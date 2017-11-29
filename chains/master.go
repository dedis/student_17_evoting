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

func Unmarshal(roster *onet.Roster, genesis skipchain.SkipBlockID) (
	*Master, []*Link, []*skipchain.SkipBlock, error) {

	client := skipchain.NewClient()
	chain, cerr := client.GetUpdateChain(roster, genesis)
	if cerr != nil {
		return nil, nil, nil, cerr
	}
	_, blob, err := network.Unmarshal(chain.Update[0].Data)
	if err != nil {
		return nil, nil, nil, err
	}

	links := make([]*Link, len(chain.Update)-1)
	for i := 1; i <= len(links); i++ {
		_, blob, err := network.Unmarshal(chain.Update[i].Data)
		if err != nil {
			return nil, nil, nil, err
		}

		links[i-1] = blob.(*Link)
	}

	return blob.(*Master), links, chain.Update, nil
}
