package service

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/dedis/cothority/skipchain"
	"github.com/stretchr/testify/assert"

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
