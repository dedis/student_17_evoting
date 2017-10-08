package storage

import (
	"sync"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"

	"gopkg.in/dedis/onet.v1/network"
)

type Chain struct {
	sync.Mutex

	Genesis *skipchain.SkipBlock
}

func (c *Chain) Election() *api.Election {
	c.Lock()
	defer c.Unlock()

	_, blob, _ := network.Unmarshal(c.Genesis.Data)
	return blob.(*api.Election)
}

func (c *Chain) LatestBlock() (*skipchain.SkipBlock, error) {
	c.Lock()
	defer c.Unlock()

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(c.Genesis.Roster, c.Genesis.Hash)
	if err != nil {
		return nil, err
	}

	return chain.Update[len(chain.Update)-1], nil
}

func (c *Chain) Store(data interface{}) (int, error) {
	latest, err := c.LatestBlock()
	if err != nil {
		return -1, err
	}
	client := skipchain.NewClient()
	response, err := client.StoreSkipBlock(latest, nil, data)
	if err != nil {
		return -1, err
	}

	return response.Latest.Index, nil
}

func (c *Chain) Ballots() ([]*api.BallotNew, error) {
	c.Lock()
	defer c.Unlock()

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(c.Genesis.Roster, c.Genesis.Hash)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]*api.BallotNew)
	for i := 1; i < len(chain.Update); i++ {
		block, err := client.GetSingleBlockByIndex(c.Genesis.Roster, c.Genesis.Hash, i)
		if err != nil {
			return nil, err
		}

		_, blob, _ := network.Unmarshal(block.Data)
		ballot, ok := blob.(*api.BallotNew)
		if !ok {
			break
		}

		mapping[ballot.User] = ballot
	}

	ballots, index := make([]*api.BallotNew, len(mapping)), 0
	for _, ballot := range mapping {
		ballots[index] = ballot
		index++
	}

	return ballots, nil
}
