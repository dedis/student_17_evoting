package shuffle

import (
	"errors"
	"testing"
	"time"

	"github.com/dedis/onet"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

var serviceID onet.ServiceID

type service struct {
	*onet.ServiceProcessor
	election *chains.Election
}

func init() {
	new := func(ctx *onet.Context) (onet.Service, error) {
		return &service{ServiceProcessor: onet.NewServiceProcessor(ctx)}, nil
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
	for _, nodes := range []int{3} {
		run(t, nodes)
	}
}

func run(t *testing.T, n int) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, tree := local.GenBigTree(n, n, 1, true)

	election := &chains.Election{Roster: roster, Stage: chains.RUNNING}
	_ = election.GenChain(n)

	services := local.GetServices(nodes, serviceID)
	for i := range services {
		services[i].(*service).election = election
	}

	instance, _ := services[0].(*service).CreateProtocol(Name, tree)
	protocol := instance.(*Protocol)
	protocol.Election = election
	protocol.Start()

	select {
	case <-protocol.Finished:
		// box, _ := election.Box()
		// mixes, _ := election.Mixes()

		// x, y := chains.Split(box.Ballots)
		// v, w := chains.Split(mixes[0].Ballots)
		// fmt.Println(crypto.Verify(mixes[0].Proof, election.Key, x, y, v, w))
	case <-time.After(3 * time.Second):
		t.Fatal("Protocol timeout")
	}
}
