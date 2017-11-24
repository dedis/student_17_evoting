package service

import (
	"github.com/dedis/cothority/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1/network"
)

type master struct {
	Key    abstract.Point
	Admins []uint32
}

type link struct {
	Genesis skipchain.SkipBlockID
}

func init() {
	network.RegisterMessage(&master{})
	network.RegisterMessage(&link{})
}

func (m *master) admin(sciper uint32) bool {
	for _, admin := range m.Admins {
		if admin == sciper {
			return true
		}
	}
	return false
}
