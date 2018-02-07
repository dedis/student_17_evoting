package shuffle

import (
	"errors"
	"testing"
	"time"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/decrypt"
	"github.com/stretchr/testify/assert"
)

var serviceID onet.ServiceID

type service struct {
	*onet.ServiceProcessor
	election *chains.Election
}

func init() {
	new := func(ctx *onet.Context) onet.Service {
		return &service{ServiceProcessor: onet.NewServiceProcessor(ctx)}
	}
	serviceID, _ = onet.RegisterNewService(Name, new)
}

func (s *service) NewProtocol(n *onet.TreeNodeInstance, c *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	switch n.ProtocolName() {
	case Name:
		instance, _ := New(n)
		protocol := instance.(*Protocol)
		protocol.Election = s.election
		return protocol, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

func TestProtocol(t *testing.T) {
	for _, nodes := range []int{3, 5, 7} {
		run(t, nodes)
	}
}

func run(t *testing.T, n int) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, tree := local.GenBigTree(n, n, 1, true)

	election := &chains.Election{Key: crypto.Random(), Data: []byte{}}

	services := local.GetServices(nodes, serviceID)
	for i := range services {
		services[i].(*service).election = election
	}

	chain, _ := chains.New(roster, nil)

	election.ID = chain.Hash
	election.Roster = roster

	b1 := &chains.Ballot{User: 0, Alpha: crypto.Random(), Beta: crypto.Random()}
	b2 := &chains.Ballot{User: 1, Alpha: crypto.Random(), Beta: crypto.Random()}
	box := &chains.Box{Ballots: []*chains.Ballot{b1, b2}}

	chains.Store(roster, election.ID, election, b1, b2, box)

	instance, _ := services[0].(*service).CreateProtocol(Name, tree)
	protocol := instance.(*Protocol)
	protocol.Election = election
	protocol.Start()

	select {
	case <-protocol.Finished:
		box, _ := election.Box()
		mixes, _ := election.Mixes()
		assert.True(t, decrypt.Verify(election.Key, box, mixes))
	case <-time.After(5 * time.Second):
		t.Fatal("Protocol timeout")
	}
}
