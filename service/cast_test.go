package service

import (
	"testing"

	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/onet"
	"github.com/dedis/onet/network"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestCast_InvalidElectionID(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0}

	_, err := s.Cast(&api.Cast{Token: "0", ID: []byte{}})
	assert.NotNil(t, err)
}

func TestCast_UserNotPart(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1"] = &stamp{user: 1}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	_, err := s.Cast(&api.Cast{Token: "1", ID: election.ID})
	assert.Equal(t, ERR_NOT_PART, err)
}

func TestCast_ElectionAlreadyClosed(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.SHUFFLED,
	}
	_ = election.GenChain(3)

	_, err := s.Cast(&api.Cast{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_ALREADY_CLOSED, err)

	election = &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.DECRYPTED,
	}
	_ = election.GenChain(3)

	_, err = s.Cast(&api.Cast{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_ALREADY_CLOSED, err)
}

func TestCast_Full(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1000"] = &stamp{user: 1000}

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{1000},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	ballot := &chains.Ballot{User: 1000}
	r, _ := s.Cast(&api.Cast{Token: "1000", ID: election.ID, Ballot: ballot})
	assert.NotNil(t, r)

	client := skipchain.NewClient()
	chain, _ := client.GetUpdateChain(roster, election.ID)
	_, blob, _ := network.Unmarshal(chain.Update[len(chain.Update)-1].Data, crypto.Suite)
	assert.Equal(t, ballot.User, blob.(*chains.Ballot).User)
}
