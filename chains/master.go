package chains

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

// Master is the foundation object of the entire service.
// It contains mission critical information that can only be accessed and
// set by an administrators.
type Master struct {
	ID     skipchain.SkipBlockID // ID is the hash of the genesis skipblock.
	Roster *onet.Roster          // Roster is the set of responsible conodes.

	Admins []uint32 // Admins is the list of administrators.

	Key abstract.Point // Key is the front-end public key.
}

// Link is a wrapper around the genesis Skipblock identifier of an
// election. Every newly created election adds a new link to the master Skipchain.
type Link struct {
	ID skipchain.SkipBlockID
}

func init() {
	network.RegisterMessages(Master{}, Link{})
}

// FetchMaster retrieves the master object from its skipchain.
func FetchMaster(roster *onet.Roster, id skipchain.SkipBlockID) (*Master, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	_, blob, _ := network.Unmarshal(chain[1].Data)
	return blob.(*Master), nil
}

// Links returns all the links appended to the master skipchain.
func (m *Master) Links() ([]*Link, error) {
	chain, err := chain(m.Roster, m.ID)
	if err != nil {
		return nil, err
	}

	links := make([]*Link, 0)
	for i := 2; i < len(chain); i++ {
		_, blob, _ := network.Unmarshal(chain[i].Data)
		links = append(links, blob.(*Link))
	}
	return links, nil
}

// IsAdmin checks if a given user is part of the administrator list.
func (m *Master) IsAdmin(user uint32) bool {
	for _, admin := range m.Admins {
		if admin == user {
			return true
		}
	}
	return false
}
