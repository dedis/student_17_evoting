package service

import (
	"testing"

	"github.com/dedis/onet"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestShuffle_UserNotLoggedIn(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.Shuffle(&api.Shuffle{Token: ""})
	assert.Equal(t, ERR_NOT_LOGGED_IN, err)
}

func TestShuffle_UserNotAdmin(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1"] = &stamp{user: 1, admin: false}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	_, err := s.Shuffle(&api.Shuffle{Token: "1", ID: election.ID})
	assert.Equal(t, ERR_NOT_ADMIN, err)
}

func TestShuffle_UserNotCreator(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1"] = &stamp{user: 1, admin: true}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0, 1},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	_, err := s.Shuffle(&api.Shuffle{Token: "1", ID: election.ID})
	assert.Equal(t, ERR_NOT_CREATOR, err)
}

func TestShuffle_ElectionClosed(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.SHUFFLED,
	}
	_ = election.GenChain(3)

	_, err := s.Shuffle(&api.Shuffle{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_ALREADY_SHUFFLED, err)

	election = &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.DECRYPTED,
	}
	_ = election.GenChain(3)

	_, err = s.Shuffle(&api.Shuffle{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_ALREADY_SHUFFLED, err)
}

func TestShuffle_Full(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	r, _ := s.Shuffle(&api.Shuffle{Token: "0", ID: election.ID})
	assert.NotNil(t, r)
}
