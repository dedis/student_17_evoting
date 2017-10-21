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

	key1 := services[0].Storage.Chains["test"].Election().Key
	key2 := services[1].Storage.Chains["test"].Election().Key
	key3 := services[2].Storage.Chains["test"].Election().Key

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

	assert.Equal(t, uint32(2), cbr.Block)

	ballots1, _ := services[0].Storage.Chains["test"].Ballots()
	ballots2, _ := services[1].Storage.Chains["test"].Ballots()
	ballots3, _ := services[2].Storage.Chains["test"].Ballots()

	assert.Equal(t, ballots1[0], ballots2[0], ballots3[0])
}

func TestGetBallots(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	ge := &api.GenerateElection{Token: "", Election: election}

	response, _ := services[0].GenerateElection(ge)

	<-time.After(500 * time.Millisecond)

	alpha1, beta1 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot1 := api.BallotNew{"user1", alpha1, beta1, []byte{}}
	alpha2, beta2 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot2 := api.BallotNew{"user2", alpha2, beta2, []byte{}}
	alpha3, beta3 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot3 := api.BallotNew{"user2", alpha3, beta3, []byte{}}
	alpha4, beta4 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot4 := api.BallotNew{"user3", alpha4, beta4, []byte{}}

	_, _ = services[0].CastBallot(&api.CastBallot{"", "test", ballot1})
	_, _ = services[1].CastBallot(&api.CastBallot{"", "test", ballot2})
	_, _ = services[2].CastBallot(&api.CastBallot{"", "test", ballot3})
	_, _ = services[0].CastBallot(&api.CastBallot{"", "test", ballot4})

	gbr, err := services[0].GetBallots(&api.GetBallots{"", "test"})
	if err != nil {
		log.ErrFatal(err)
	}

	assert.Equal(t, 3, len(gbr.Ballots))
}

func TestShuffle(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	ge := &api.GenerateElection{Token: "", Election: election}

	response, _ := services[0].GenerateElection(ge)

	<-time.After(500 * time.Millisecond)

	alpha1, beta1 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot1 := api.BallotNew{"user1", alpha1, beta1, []byte{}}
	alpha2, beta2 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot2 := api.BallotNew{"user2", alpha2, beta2, []byte{}}

	_, _ = services[0].CastBallot(&api.CastBallot{"", "test", ballot1})
	_, _ = services[1].CastBallot(&api.CastBallot{"", "test", ballot2})

	shr, err := services[0].Shuffle(&api.Shuffle{"", "test"})
	if err != nil {
		log.ErrFatal(err)
	}

	assert.Equal(t, 4, int(shr.Block))
}

func TestGetShuffle(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	ge := &api.GenerateElection{Token: "", Election: election}

	response, _ := services[0].GenerateElection(ge)

	<-time.After(500 * time.Millisecond)

	_, err := services[0].GetShuffle(&api.GetShuffle{"", "test"})
	assert.NotNil(t, err)

	alpha1, beta1 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot1 := api.BallotNew{"user1", alpha1, beta1, []byte{}}
	alpha2, beta2 := encrypt(api.Suite, response.Key, []byte{1, 2, 3})
	ballot2 := api.BallotNew{"user2", alpha2, beta2, []byte{}}

	_, _ = services[0].CastBallot(&api.CastBallot{"", "test", ballot1})
	_, _ = services[1].CastBallot(&api.CastBallot{"", "test", ballot2})

	_, _ = services[0].Shuffle(&api.Shuffle{"", "test"})

	gsr, _ := services[0].GetShuffle(&api.GetShuffle{"", "test"})
	assert.Equal(t, 2, len(gsr.Box.Ballots))
}

func TestDecrypt(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	ge := &api.GenerateElection{Token: "", Election: election}

	response, _ := services[0].GenerateElection(ge)

	<-time.After(500 * time.Millisecond)

	alpha1, beta1 := encrypt(api.Suite, response.Key, []byte("user1"))
	ballot1 := api.BallotNew{"user1", alpha1, beta1, []byte{}}
	alpha2, beta2 := encrypt(api.Suite, response.Key, []byte("user2"))
	ballot2 := api.BallotNew{"user2", alpha2, beta2, []byte{}}
	alpha3, beta3 := encrypt(api.Suite, response.Key, []byte("user3"))
	ballot3 := api.BallotNew{"user3", alpha3, beta3, []byte{}}

	_, _ = services[0].CastBallot(&api.CastBallot{"", "test", ballot1})
	_, _ = services[1].CastBallot(&api.CastBallot{"", "test", ballot2})
	_, _ = services[2].CastBallot(&api.CastBallot{"", "test", ballot3})

	_, _ = services[0].Shuffle(&api.Shuffle{"", "test"})

	dr, err := services[0].Decrypt(&api.Decrypt{"", "test"})
	if err != nil {
		log.ErrFatal(err)
	}
	assert.Equal(t, uint32(6), dr.Block)

	boxes, _ := services[2].Storage.Chains["test"].Boxes()
	assert.Equal(t, 2, len(boxes))

	assert.Equal(t, boxes[1].Ballots[0].User, string(boxes[1].Ballots[0].Clear))
	assert.Equal(t, boxes[1].Ballots[1].User, string(boxes[1].Ballots[1].Clear))
	assert.Equal(t, boxes[1].Ballots[2].User, string(boxes[1].Ballots[2].Clear))
}
