package chains

import (
	"encoding/base64"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(&Master{})
	network.RegisterMessage(&Link{})
}

// Master is the foundation object of the entire service.
// It contains mission critical information that can only be
// set by an administrator that has access the conodes.
type Master struct {
	// Key is the front-end public for authenticity control.
	Key abstract.Point
	// ID is the identifier of the master Skipchain.
	ID skipchain.SkipBlockID
	// Roster is a list of conodes handling the service.
	Roster *onet.Roster
	// Admins is list of users that can execute priviledged instructions.
	Admins []User
}

// Link is a wrapper around the genesis Skipblock identifier of an
// election. Every newly created election adds a new link to the
// master Skipchain.
type Link struct {
	Genesis skipchain.SkipBlockID
}

func FetchMaster(roster *onet.Roster, id string) (*Master, error) {
	conv, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, err
	}
	chain, err := chain(roster, conv)
	if err != nil {
		return nil, err
	}

	// By definition the master object is stored right after the genesis Skipblock.
	_, blob, _ := network.Unmarshal(chain[1].Data)
	return blob.(*Master), nil
}

func (m *Master) Append(data interface{}) (int, error) {
	chain, _ := chain(m.Roster, m.ID)
	block, err := client.StoreSkipBlock(chain[len(chain)-1], m.Roster, data)
	return block.Latest.Index, err
}

func (m *Master) Links() ([]*Link, error) {
	chain, _ := chain(m.Roster, m.ID)

	links := make([]*Link, 0)
	for i := 2; i < len(chain); i++ {
		_, blob, _ := network.Unmarshal(chain[i].Data)
		links = append(links, blob.(*Link))
	}
	return links, nil
}

// IsAdmin checks if a given user is part of the administrator list
// of the master Skipchain.
func (m *Master) IsAdmin(user User) bool {
	for _, admin := range m.Admins {
		if admin == user {
			return true
		}
	}
	return false
}
