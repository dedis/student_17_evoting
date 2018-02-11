package service

import (
	"testing"

	"github.com/dedis/onet"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
)

func TestDecrypt_UserNotLoggedIn(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.Decrypt(&api.Decrypt{Token: ""})
	assert.Equal(t, ERR_NOT_LOGGED_IN, err)
}

func TestDecrypt_UserNotAdmin(t *testing.T) {
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

	_, err := s.Decrypt(&api.Decrypt{Token: "1", ID: election.ID})
	assert.Equal(t, ERR_NOT_ADMIN, err)
}

func TestDecrypt_UserNotCreator(t *testing.T) {
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

	_, err := s.Decrypt(&api.Decrypt{Token: "1", ID: election.ID})
	assert.Equal(t, ERR_NOT_CREATOR, err)
}

func TestDecrypt_ElectionNotShuffled(t *testing.T) {
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

	_, err := s.Decrypt(&api.Decrypt{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_NOT_SHUFFLED, err)
}

func TestDecrypt_ElectionClosed(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.DECRYPTED,
	}
	_ = election.GenChain(3)

	_, err := s.Decrypt(&api.Decrypt{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_ALREADY_DECRYPTED, err)
}

func TestDecrypt_Full(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s0 := local.GetServices(nodes, serviceID)[0].(*Service)
	s1 := local.GetServices(nodes, serviceID)[1].(*Service)
	s2 := local.GetServices(nodes, serviceID)[2].(*Service)
	s0.state.log["0"] = &stamp{user: 0, admin: true}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.SHUFFLED,
	}
	dkgs := election.GenChain(3)
	s0.secrets[election.ID.Short()], _ = dkg.NewSharedSecret(dkgs[0])
	s1.secrets[election.ID.Short()], _ = dkg.NewSharedSecret(dkgs[1])
	s2.secrets[election.ID.Short()], _ = dkg.NewSharedSecret(dkgs[2])

	r, _ := s0.Decrypt(&api.Decrypt{Token: "0", ID: election.ID})
	assert.NotNil(t, r)
}
