package shuffle

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/proof"
	"gopkg.in/dedis/crypto.v0/shuffle"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

var serviceID onet.ServiceID

var box *chains.Box
var election *chains.Election

type service struct {
	*onet.ServiceProcessor
}

func init() {
	new := func(ctx *onet.Context) onet.Service {
		return &service{ServiceProcessor: onet.NewServiceProcessor(ctx)}
	}
	serviceID, _ = onet.RegisterNewService(Name, new)
}

func TestProtocol(t *testing.T) {
	b1 := &chains.Ballot{Alpha: crypto.Random(), Beta: crypto.Random()}
	b2 := &chains.Ballot{Alpha: crypto.Random(), Beta: crypto.Random()}
	box = &chains.Box{Ballots: []*chains.Ballot{b1, b2}}
	election = &chains.Election{Key: crypto.Random()}

	for _, nodes := range []int{3, 5, 10} {
		run(t, nodes)
	}
}

func verify(t *testing.T, box *chains.Box, mixes []*chains.Mix) {
	a, b := split(box.Ballots)
	c, d := split(mixes[0].Ballots)

	verifier := shuffle.Verifier(crypto.Suite, nil, election.Key, a, b, c, d)
	assert.Nil(t, proof.HashVerify(crypto.Suite, Name, verifier, mixes[0].Proof))

	for i := 0; i < len(mixes)-1; i++ {
		a, b = split(mixes[i].Ballots)
		c, d = split(mixes[i+1].Ballots)

		verifier := shuffle.Verifier(crypto.Suite, nil, election.Key, a, b, c, d)
		assert.Nil(t, proof.HashVerify(crypto.Suite, Name, verifier, mixes[i+1].Proof))
	}
}

func run(t *testing.T, n int) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, tree := local.GenBigTree(n, n, 1, true)

	chain, _ := chains.New(roster, nil)

	election.ID = chain.Hash
	election.Roster = roster
	chains.Store(election.Roster, election.ID, election)
	chains.Store(election.Roster, election.ID, box)

	services := local.GetServices(nodes, serviceID)

	instance, _ := services[0].(*service).CreateProtocol(Name, tree)
	protocol := instance.(*Protocol)
	protocol.Election = election
	protocol.Start()

	select {
	case <-protocol.Finished:
		box, _ := election.Box()
		mixes, _ := election.Mixes()
		verify(t, box, mixes)
	case <-time.After(2 * time.Second):
		t.Fatal("Protocol timeout")
	}
}

func (s *service) NewProtocol(n *onet.TreeNodeInstance, c *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	switch n.ProtocolName() {
	case Name:
		instance, _ := New(n)
		protocol := instance.(*Protocol)
		protocol.Election = election
		return protocol, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}
