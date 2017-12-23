package decrypt

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
)

var serviceID onet.ServiceID

type service struct {
	*onet.ServiceProcessor

	secret *dkg.SharedSecret
}

func init() {
	serviceID, _ = onet.RegisterNewService(Name, newService)
}

func TestProtocol(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()
	nodes, _, tree := local.GenBigTree(3, 3, 3, true)

	dkgs := dkg.Simulate(3, 2)

	services := local.GetServices(nodes, serviceID)
	for i := range services {
		services[i].(*service).secret, _ = dkg.NewSharedSecret(dkgs[i])
	}

	ballots := make([]*chains.Ballot, 10)
	for i := 0; i < 10; i++ {
		k, c := crypto.Encrypt(services[0].(*service).secret.X, []byte{byte(i)})
		ballots[i] = &chains.Ballot{chains.User(i), k, c, nil}
	}

	instance, _ := services[0].(*service).CreateProtocol(Name, tree)
	protocol := instance.(*Protocol)
	protocol.Secret, _ = dkg.NewSharedSecret(dkgs[0])
	protocol.Shuffle = &chains.Box{ballots}
	protocol.Start()

	select {
	case <-protocol.Finished:
		for _, b := range protocol.Decryption.Ballots {
			assert.Equal(t, byte(b.User), b.Text[0])
		}
	case <-time.After(2 * time.Second):
		assert.True(t, false)
	}
}

func (s *service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	switch node.ProtocolName() {
	case Name:
		instance, err := New(node)
		if err != nil {
			return nil, err
		}
		protocol := instance.(*Protocol)
		protocol.Secret = s.secret
		return protocol, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

func newService(ctx *onet.Context) onet.Service {
	return &service{ServiceProcessor: onet.NewServiceProcessor(ctx)}
}
