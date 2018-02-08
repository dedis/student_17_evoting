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

var serviceID onet.ServiceID
var client = skipchain.NewClient()

func init() {
	serviceID, _ = onet.RegisterNewService(Name, new)
}

func TestLink_WrongPin(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	_, err := s.Link(&api.Link{Pin: "0"})
	assert.NotNil(t, err)
}

func TestLink_InvalidRoster(t *testing.T) {
	local := onet.NewLocalTest()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	local.CloseAll()

	_, err := s.Link(&api.Link{Pin: s.pin, Roster: roster})
	assert.NotNil(t, err)
}

func TestLink_Full(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	r, _ := s.Link(&api.Link{Pin: s.pin, Roster: roster})
	assert.NotNil(t, r)

	chain, _ := client.GetUpdateChain(roster, r.ID)
	_, blob, _ := network.Unmarshal(chain.Update[1].Data)
	assert.Equal(t, r.ID, blob.(*chains.Master).ID)
}

func TestOpen_NotLoggedIn(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.Open(&api.Open{Token: ""})
	assert.NotNil(t, err)
	_, err = s.Open(&api.Open{Token: "0"})
	assert.NotNil(t, err)
}

func TestOpen_NotAdmin(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.Open(&api.Open{Token: "0"})
	assert.NotNil(t, err)
}

func TestOpen_InvalidMasterID(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}

	_, err := s.Open(&api.Open{Token: "0"})
	assert.NotNil(t, err)
}

func TestOpen_CloseConnection(t *testing.T) {
	local := onet.NewLocalTest()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}

	master := chains.GenMasterChain(roster, nil)

	local.CloseAll()
	_, err := s.Open(&api.Open{Token: "0", ID: master.ID})
	assert.NotNil(t, err)
}

func TestOpen_Full(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}

	master := chains.GenMasterChain(roster, nil)
	election := &chains.Election{Data: []byte{}}

	r, _ := s.Open(&api.Open{Token: "0", ID: master.ID, Election: election})
	assert.NotNil(t, r)

	chain, _ := client.GetUpdateChain(roster, r.ID)
	_, blob, _ := network.Unmarshal(chain.Update[1].Data)
	assert.Equal(t, r.ID, blob.(*chains.Election).ID)

	assert.Equal(t, r.Key, s.secrets[r.ID.Short()].X)
}

func TestLogin_InvalidMasterID(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	_, err := s.Login(&api.Login{ID: nil})
	assert.NotNil(t, err)
}

func TestLogin_InvalidLink(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	master := chains.GenMasterChain(roster, []byte{})

	_, err := s.Login(&api.Login{ID: master.ID})
	assert.NotNil(t, err)
}

func TestLogin_Full(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	election, _ := chains.GenElectionChain(roster, 0, []uint32{0}, 3, chains.RUNNING)
	master := chains.GenMasterChain(roster, election.ID)

	r, _ := s.Login(&api.Login{User: 0, ID: master.ID})
	assert.Equal(t, election.ID, r.Elections[0].ID)
	assert.Equal(t, uint32(0), s.state.log[r.Token].user)
}

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

	chain, _ := client.GetUpdateChain(roster, election.ID)
	_, blob, _ := network.Unmarshal(chain.Update[len(chain.Update)-1].Data)
	assert.Equal(t, ballot.User, blob.(*chains.Ballot).User)
}
