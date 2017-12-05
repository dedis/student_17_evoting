package chains

import (
	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(&Master{})
	network.RegisterMessage(&Link{})
}

type Master struct {
	Key    abstract.Point
	Roster *onet.Roster
	Admins []User
}

type Link struct {
	Genesis skipchain.SkipBlockID
}

func (m *Master) IsAdmin(user User) bool {
	for _, admin := range m.Admins {
		if admin == user {
			return true
		}
	}
	return false
}
