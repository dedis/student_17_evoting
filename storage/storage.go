package storage

import (
	"errors"
	"sync"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/dkg"
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

// CreateElection adds a new election structure to the storage map.
func (storage *Storage) CreateElection(name string, genesis, latest *skipchain.SkipBlock,
	shared *dkg.SharedSecret) {

	storage.Lock()
	defer storage.Unlock()

	if latest == nil {
		storage.Elections[name] = &Election{genesis, genesis, shared}
	} else {
		storage.Elections[name] = &Election{genesis, latest, shared}
	}
}

// UpdateLatest replaces the latest SkipBlock of an election by a given SkipBlock.
func (storage *Storage) UpdateLatest(name string, latest *skipchain.SkipBlock) {
	storage.Lock()
	defer storage.Unlock()

	storage.Elections[name].Latest = latest
}
