package storage

import (
	"sync"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/dkg"

	"gopkg.in/dedis/onet.v1/network"
)

type Chain struct {
	sync.Mutex

	SharedSecret *dkg.SharedSecret
	Genesis      *skipchain.SkipBlock
}

// TODO: Handle error from GetSingleBlockByIndex
func (c *Chain) Election() *api.Election {
	c.Lock()
	defer c.Unlock()

	client := skipchain.NewClient()
	block, _ := client.GetSingleBlockByIndex(c.Genesis.Roster, c.Genesis.Hash, 1)

	_, blob, _ := network.Unmarshal(block.Data)
	return blob.(*api.Election)
}

func (c *Chain) IsShuffled() bool {
	boxes, _ := c.Boxes()

	return len(boxes) >= 1
}

func (c *Chain) IsDecrypted() bool {
	boxes, _ := c.Boxes()

	return len(boxes) == 2
}

// func (c *Chain) LatestBlock() (*skipchain.SkipBlock, error) {
// 	c.Lock()
// 	defer c.Unlock()

// 	client := skipchain.NewClient()
// 	chain, err := client.GetUpdateChain(c.Genesis.Roster, c.Genesis.Hash)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return chain.Update[len(chain.Update)-1], nil
// }

func (c *Chain) Store(data interface{}) (int, error) {
	c.Lock()
	defer c.Unlock()

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(c.Genesis.Roster, c.Genesis.Hash)
	if err != nil {
		return -1, err
	}

	latest := chain.Update[len(chain.Update)-1]
	response, _ := client.StoreSkipBlock(latest, nil, data)

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
	for i := 2; i < len(chain.Update); i++ {
		block, _ := client.GetSingleBlockByIndex(c.Genesis.Roster, c.Genesis.Hash, i)

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

func (c *Chain) Boxes() ([]*api.BoxNew, error) {
	c.Lock()
	defer c.Unlock()

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(c.Genesis.Roster, c.Genesis.Hash)
	if err != nil {
		return nil, err
	}

	boxes := make([]*api.BoxNew, 0)
	for i := 1; i < len(chain.Update); i++ {
		block, _ := client.GetSingleBlockByIndex(c.Genesis.Roster, c.Genesis.Hash, i)

		_, blob, _ := network.Unmarshal(block.Data)
		box, ok := blob.(*api.BoxNew)
		if ok {
			boxes = append(boxes, box)
		}
	}

	return boxes, nil
}
