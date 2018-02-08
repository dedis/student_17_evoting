package service

import (
	"testing"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/stretchr/testify/assert"
)

func TestCast_InvalidElectionID(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0}

	_, err := s.Cast(&api.Cast{Token: "0", ID: []byte{}})
	assert.NotNil(t, err)
}

func TestCast_UserNotPart(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1"] = &stamp{user: 1}

	election, _ := chains.GenElectionChain(roster, 0, []uint32{0}, 3, chains.RUNNING)

	_, err := s.Cast(&api.Cast{Token: "1", ID: election.ID})
	assert.NotNil(t, err)
}

func TestCast_ElectionAlreadyClosed(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0}

	election, _ := chains.GenElectionChain(roster, 0, []uint32{0}, 3, chains.SHUFFLED)
	_, err := s.Cast(&api.Cast{Token: "0", ID: election.ID})
	assert.NotNil(t, err)

	election, _ = chains.GenElectionChain(roster, 0, []uint32{0}, 3, chains.DECRYPTED)
	_, err = s.Cast(&api.Cast{Token: "0", ID: election.ID})
	assert.NotNil(t, err)
}

func TestCast_Full(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["1000"] = &stamp{user: 1000}

	election, _ := chains.GenElectionChain(roster, 0, []uint32{1000}, 3, chains.RUNNING)
	ballot := &chains.Ballot{User: 1000}

	r, _ := s.Cast(&api.Cast{Token: "1000", ID: election.ID, Ballot: ballot})
	assert.NotNil(t, r)

	client := skipchain.NewClient()
	chain, _ := client.GetUpdateChain(roster, election.ID)
	_, blob, _ := network.Unmarshal(chain.Update[len(chain.Update)-1].Data)
	assert.Equal(t, ballot.User, blob.(*chains.Ballot).User)
}
