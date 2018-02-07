package decrypt

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
	"github.com/qantik/nevv/shuffle"
	"github.com/stretchr/testify/assert"
)

var serviceID onet.ServiceID

type service struct {
	*onet.ServiceProcessor
	secret   *dkg.SharedSecret
	election *chains.Election
}

func init() {
	new := func(ctx *onet.Context) onet.Service {
		return &service{ServiceProcessor: onet.NewServiceProcessor(ctx)}
	}
	serviceID, _ = onet.RegisterNewService(Name, new)
}

func (s *service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	switch node.ProtocolName() {
	case Name:
		instance, _ := New(node)
		protocol := instance.(*Protocol)
		protocol.Secret = s.secret
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

	dkgs := dkg.Simulate(n, n-1)

	election := &chains.Election{Key: crypto.Random(), Data: []byte{}}

	services := local.GetServices(nodes, serviceID)
	for i := range services {
		services[i].(*service).secret, _ = dkg.NewSharedSecret(dkgs[i])
		services[i].(*service).election = election
	}

	chain, _ := chains.New(roster, nil)
	election.ID = chain.Hash
	election.Roster = roster

	b1 := &chains.Ballot{User: 0, Alpha: crypto.Random(), Beta: crypto.Random()}
	b2 := &chains.Ballot{User: 1, Alpha: crypto.Random(), Beta: crypto.Random()}
	box := &chains.Box{Ballots: []*chains.Ballot{b1, b2}}
	mixes := shuffle.Simulate(n, election.Key, box.Ballots)

	chains.Store(roster, election.ID, election, b1, b2, box)
	for _, mix := range mixes {
		chains.Store(roster, election.ID, mix)
	}

	instance, _ := services[0].(*service).CreateProtocol(Name, tree)
	protocol := instance.(*Protocol)
	protocol.Secret, _ = dkg.NewSharedSecret(dkgs[0])
	protocol.Election = election
	protocol.Start()

	select {
	case <-protocol.Finished:
		partials, _ := election.Partials()
		for _, partial := range partials {
			fmt.Println(partial)
			assert.False(t, partial.Flag)
		}
	case <-time.After(5 * time.Second):
		assert.True(t, false)
	}
}
