package decrypt

import (
	"errors"
	"testing"
	"time"

	"gopkg.in/dedis/crypto.v0/abstract"
	gen "gopkg.in/dedis/crypto.v0/share/dkg"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/dkg"
	"github.com/stretchr/testify/assert"
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

	dkgs, _ := runDKG(3, 2)

	services := local.GetServices(nodes, serviceID)
	for i := range services {
		services[i].(*service).secret, _ = dkg.NewSharedSecret(dkgs[i])
	}

	ballots := make([]*chains.Ballot, 10)
	for i := 0; i < 10; i++ {
		k, c := encrypt(services[0].(*service).secret.X, []byte{byte(i)})
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

func runDKG(nbrNodes, threshold int) (dkgs []*gen.DistKeyGenerator, err error) {
	dkgs = make([]*gen.DistKeyGenerator, nbrNodes)
	scalars := make([]abstract.Scalar, nbrNodes)
	points := make([]abstract.Point, nbrNodes)

	// 1a - initialisation
	for i := range scalars {
		scalars[i] = suite.Scalar().Pick(stream)
		points[i] = suite.Point().Mul(nil, scalars[i])
	}

	// 1b - key-sharing
	for i := range dkgs {
		dkgs[i], err = gen.NewDistKeyGenerator(suite,
			scalars[i], points, stream, threshold)
		if err != nil {
			return
		}
	}
	// Exchange of Deals
	responses := make([][]*gen.Response, nbrNodes)
	for i, p := range dkgs {
		responses[i] = make([]*gen.Response, nbrNodes)
		deals, err := p.Deals()
		if err != nil {
			return nil, err
		}
		for j, d := range deals {
			responses[i][j], err = dkgs[j].ProcessDeal(d)
			if err != nil {
				return nil, err
			}
		}
	}
	// ProcessResponses
	for _, resp := range responses {
		for j, r := range resp {
			for k, p := range dkgs {
				if r != nil && j != k {
					justification, err := p.ProcessResponse(r)
					if err != nil {
						return nil, err
					}
					if justification != nil {
						return nil, errors.New("invalid justification")
					}
				}
			}
		}
	}

	// Secret commits
	for _, p := range dkgs {
		commit, err := p.SecretCommits()
		if err != nil {
			return nil, err
		}
		for _, p2 := range dkgs {
			compl, err := p2.ProcessSecretCommits(commit)
			if err != nil {
				return nil, err
			}
			if compl != nil {
				return nil, errors.New("there should be no complaint")
			}
		}
	}

	// Verify if all is OK
	for _, p := range dkgs {
		if !p.Finished() {
			return nil, errors.New("one of the dkgs is not finished yet")
		}
	}
	return
}

func encrypt(key abstract.Point, msg []byte) (K, C abstract.Point) {
	M, _ := suite.Point().Pick(msg, stream)
	k := suite.Scalar().Pick(stream)
	K = suite.Point().Mul(nil, k)
	S := suite.Point().Mul(key, k)
	C = S.Add(S, M)
	return
}
