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
