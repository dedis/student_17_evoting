package service

import (
	"errors"
	"sync"

	"github.com/dedis/cothority/skipchain"
)

// Storage offers the possibilty to store elections permanently on
// the disk. This is especially useful when multiple elections have to
// kept alive after potential shutdowns of the conode.
type Storage struct {
	sync.Mutex

	Elections map[string]*Election
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
