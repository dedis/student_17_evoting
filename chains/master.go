package chains

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
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
	log.Lvl3(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", m.Admins)
	for _, admin := range m.Admins {
		log.Lvl3("******", admin, user, admin == user)
		if admin == user {
			return true
		}
	}
	return false
}
