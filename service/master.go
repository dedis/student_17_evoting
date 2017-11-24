package service

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1/network"
)

type master struct {
	Key    abstract.Point
	Admins []uint32
}

func init() {
	network.RegisterMessage(&master{})
}
