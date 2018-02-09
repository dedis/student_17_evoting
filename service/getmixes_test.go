package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dedis/onet"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestGetMixes_UserNotLoggedIn(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.GetMixes(&api.GetMixes{Token: ""})
	assert.NotNil(t, ERR_NOT_LOGGED_IN, err)
}

func TestGetMixes_UserNotPart(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1"] = &stamp{user: 1, admin: false}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.SHUFFLED,
	}
	_ = election.GenChain(3)

	_, err := s.GetMixes(&api.GetMixes{Token: "1", ID: election.ID})
	assert.NotNil(t, ERR_NOT_PART, err)
}

func TestGetMixes_ElectionNotShuffled(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	_, err := s.GetMixes(&api.GetMixes{Token: "0", ID: election.ID})
	assert.NotNil(t, ERR_NOT_SHUFFLED, err)
}

func TestGetMixes_Full(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.SHUFFLED,
	}
	_ = election.GenChain(10)

	r, _ := s.GetMixes(&api.GetMixes{Token: "0", ID: election.ID})
	assert.Equal(t, 3, len(r.Mixes))
}
