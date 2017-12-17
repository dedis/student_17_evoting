package chains

import (
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
