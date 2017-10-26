package storage

import (
	"sync"

	"github.com/qantik/nevv/api"

	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(&Storage{})
}

// Storage holds all currently available election chains that are
// kept in permanent storage at the conodes.
type Storage struct {
	sync.Mutex

	Chains map[string]*Chain
}

// GetElections returns all elections for which the given user is
// either the administrator or part of the election's user list.
func (s *Storage) GetElections(user string) []*api.Election {
	elections := make([]*api.Election, 0)
	for _, c := range s.Chains {
		election := c.Election()
		if election.Admin == user {
			elections = append(elections, election)
			continue
		}

		for _, u := range election.Users {
			if u == user {
				elections = append(elections, election)
				break
			}
		}
	}

	return elections
}
