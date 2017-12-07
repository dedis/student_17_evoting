package service

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/dkg"
)

func Test50(t *testing.T) {
	local, services, _, genesis := setup(50)
	defer local.CloseAll()

	// start := time.Now()
	fr, err := services[0].Finalize(&api.Finalize{"0", genesis})
	assert.Nil(t, err)
	assert.Equal(t, 50, len(fr.Shuffle.Ballots))
	assert.Equal(t, 50, len(fr.Decryption.Ballots))
	// fmt.Printf("%s\n", time.Since(start))
}

func TestIntegration(t *testing.T) {
	encrypt := func(key abstract.Point, msg []byte) (K, C abstract.Point) {
		M, _ := suite.Point().Pick(msg, random.Stream)

		k := suite.Scalar().Pick(random.Stream)
		K = suite.Point().Mul(nil, k)
		S := suite.Point().Mul(key, k)
		C = S.Add(S, M)
		return
	}

	local := onet.NewTCPTest()
	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := make([]*Service, 3)
	for i, service := range local.GetServices(hosts, serviceID) {
		services[i] = service.(*Service)
	}
	services[0].pin = "123456"

	admin := &stamp{123, true, 0}
	user1 := &stamp{654, false, 0}
	user2 := &stamp{789, false, 0}
	services[0].state = &state{map[string]*stamp{"0": admin, "1": user1, "2": user2}}

	e := &chains.Election{"", 123, []chains.User{654, 789}, "", nil, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})

	a1, b1 := encrypt(or.Key, []byte{1, 2, 3, 10, 100})
	a2, b2 := encrypt(or.Key, []byte{1, 2, 3, 10, 100})
	ballot1 := &chains.Ballot{654, a1, b1, nil}
	ballot2 := &chains.Ballot{789, a2, b2, nil}
	services[0].Cast(&api.Cast{"1", or.Genesis, ballot1})
	services[0].Cast(&api.Cast{"2", or.Genesis, ballot2})

	// Not logged in
	fr, err := services[0].Finalize(&api.Finalize{"", or.Genesis})
	assert.Nil(t, fr)
	assert.NotNil(t, err)

	// Invalid genesis
	fr, err = services[0].Finalize(&api.Finalize{"0", ""})
	assert.Nil(t, fr)
	assert.NotNil(t, err)

	// Not the creator
	fr, err = services[0].Finalize(&api.Finalize{"1", or.Genesis})
	assert.Nil(t, fr)
	assert.NotNil(t, err)

	// Valid finalize
	fr, err = services[0].Finalize(&api.Finalize{"0", or.Genesis})
	assert.NotNil(t, fr)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3, 10, 100}, fr.Decryption.Ballots[0].Text)
	assert.Equal(t, []byte{1, 2, 3, 10, 100}, fr.Decryption.Ballots[1].Text)
}

// Create master Skipchain with admin 0 and one election Skipchain
// with n users [1, n]. All keys are set to the zero element.
func setup(n int) (*onet.LocalTest, []*Service, string, string) {
	local := onet.NewTCPTest()
	hosts, roster, _ := local.GenTree(3, true)

	services := make([]*Service, 3)
	for i, service := range local.GetServices(hosts, serviceID) {
		services[i] = service.(*Service)
	}

	client := skipchain.NewClient()
	master := &chains.Master{Roster: roster, Admins: []chains.User{0}}
	mGen, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, master, nil)

	logs := make(map[string]*stamp)
	logs["0"] = &stamp{0, true, 0}
	users := make([]chains.User, n)
	for i := 0; i < n; i++ {
		logs[strconv.Itoa(i+1)] = &stamp{chains.User(i + 1), false, 0}
		users[i] = chains.User(i + 1)
	}
	services[0].state = &state{logs}
	services[0].pin = "123456"

	election := &chains.Election{
		Name:    "election",
		Creator: chains.User(0),
		Users:   users,
		Roster:  roster,
		Key:     suite.Point(),
	}

	eGen, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, nil, nil)
	rep, _ := client.StoreSkipBlock(eGen, roster, election)
	client.StoreSkipBlock(mGen, roster, &chains.Link{eGen.Hash})

	for i := 0; i < n; i++ {
		ballot := &chains.Ballot{chains.User(i + 1), suite.Point(), suite.Point(), nil}
		rep, _ = client.StoreSkipBlock(rep.Latest, roster, ballot)
	}

	s1 := &dkg.SharedSecret{0, suite.Scalar(), suite.Point()}
	s2 := &dkg.SharedSecret{1, suite.Scalar(), suite.Point()}
	s3 := &dkg.SharedSecret{2, suite.Scalar(), suite.Point()}
	services[0].secrets = map[string]*dkg.SharedSecret{string(eGen.Hash): s1}
	services[1].secrets = map[string]*dkg.SharedSecret{string(eGen.Hash): s2}
	services[2].secrets = map[string]*dkg.SharedSecret{string(eGen.Hash): s3}

	mGenStr := base64.StdEncoding.EncodeToString(mGen.Hash)
	eGenStr := base64.StdEncoding.EncodeToString(eGen.Hash)
	return local, services, mGenStr, eGenStr
}
