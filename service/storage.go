package service

import (
	"errors"
	"sync"

	"gopkg.in/dedis/onet.v1/network"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
)

// Storage offers the possibilty to store elections permanently on
// the disk. This is especially useful when multiple elections have to
// kept alive after potential shutdowns of the conode.
type Storage struct {
	sync.Mutex

	Elections map[string]*Election
	Chains    map[string]*skipchain.SkipBlock
}

// Get retrieves an election for a given name.
func (storage *Storage) Get(name string) (*Election, error) {
	storage.Lock()
	defer storage.Unlock()

	election, found := storage.Elections[name]
	if !found {
		return nil, errors.New("Election " + name + " not found")
	}

	return election, nil
}

func (s *Storage) GetElection(id string) *api.Election {
	s.Lock()
	defer s.Unlock()

	_, blob, _ := network.Unmarshal(s.Chains[id].Data)
	return blob.(*api.Election)
}

func (s *Storage) GetLatestBlock(id string) (*skipchain.SkipBlock, error) {
	s.Lock()
	defer s.Unlock()

	genesis := s.Chains[id]

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(genesis.Roster, genesis.Hash)
	if err != nil {
		return nil, err
	}

	return chain.Update[len(chain.Update)-1], nil
}

func (s *Storage) AppendToChain(id string, data interface{}) (int, error) {
	s.Lock()
	defer s.Unlock()

	genesis := s.Chains[id]

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(genesis.Roster, genesis.Hash)
	if err != nil {
		return -1, err
	}

	latest := chain.Update[len(chain.Update)-1]

	response, err := client.StoreSkipBlock(latest, nil, data)
	if err != nil {
		return -1, err
	}

	return response.Latest.Index, nil
}

func (s *Storage) GetBallots(id string) ([]*api.BallotNew, error) {
	s.Lock()
	defer s.Unlock()

	election := s.Chains[id]

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(election.Roster, election.Hash)
	if err != nil {
		return nil, err
	}

	mapping := make(map[string]*api.BallotNew)
	for i := 1; i < len(chain.Update); i++ {
		block, err := client.GetSingleBlockByIndex(election.Roster, election.Hash, i)
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

// UpdateLatest replaces the latest SkipBlock of an election by a given SkipBlock.
func (storage *Storage) SetLatest(name string, latest *skipchain.SkipBlock) {
	storage.Lock()
	defer storage.Unlock()

	storage.Elections[name].Latest = latest
}

func (storage *Storage) SetElection(election *Election) {
	storage.Lock()
	defer storage.Unlock()

	storage.Elections[election.Name] = election
}
