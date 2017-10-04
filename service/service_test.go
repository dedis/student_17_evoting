package service

import (
	"testing"
	"time"

	"github.com/qantik/nevv/api"
	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
)

func castServices(services []onet.Service) []*Service {
	cast := make([]*Service, len(services))
	for i, service := range services {
		cast[i] = service.(*Service)
	}

	return cast
}

func encrypt(suite abstract.Suite, pub abstract.Point, msg []byte) (K, C abstract.Point) {
	M, _ := suite.Point().Pick(msg, random.Stream)

	k := suite.Scalar().Pick(random.Stream)
	K = suite.Point().Mul(nil, k)
	S := suite.Point().Mul(pub, k)
	C = S.Add(S, M)

	return
}

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestGenerateElection(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	msg := &api.GenerateElection{Token: "", Election: election}

	response, err := services[0].GenerateElection(msg)
	if err != nil {
		log.ErrFatal(err)
	}

	<-time.After(500 * time.Millisecond)

	key1 := services[0].Storage.GetElection("test").Key
	key2 := services[1].Storage.GetElection("test").Key
	key3 := services[2].Storage.GetElection("test").Key

	assert.Equal(t, key1, key2, key3, response.Key)
}

func TestCastBallot(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	ge := &api.GenerateElection{Token: "", Election: election}

	response, err := services[0].GenerateElection(ge)
	if err != nil {
		log.ErrFatal(err)
	}

	<-time.After(500 * time.Millisecond)

	alpha, beta := encrypt(api.Suite, response.Key, []byte{1, 2, 3})

	ballot := api.BallotNew{"user", alpha, beta, []byte{}}
	cb := &api.CastBallot{"", "test", ballot}

	cbr, err := services[0].CastBallot(cb)
	if err != nil {
		log.ErrFatal(err)
	}

	assert.Equal(t, uint32(1), cbr.Block)

	ballots1, _ := services[0].Storage.GetBallots("test")
	ballots2, _ := services[1].Storage.GetBallots("test")
	ballots3, _ := services[2].Storage.GetBallots("test")

	assert.Equal(t, ballots1[0], ballots2[0], ballots3[0])
}
