package storage

import (
	"sync"

	"gopkg.in/dedis/onet.v1/network"
)

type Storage struct {
	sync.Mutex

	Chains map[string]*Chain
}

func init() {
	network.RegisterMessage(&Storage{})
}
