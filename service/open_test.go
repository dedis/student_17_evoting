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

	client := skipchain.NewClient()
	chain, _ := client.GetUpdateChain(roster, r.ID)
	_, blob, _ := network.Unmarshal(chain.Update[1].Data)
	assert.Equal(t, r.ID, blob.(*chains.Election).ID)

	assert.Equal(t, r.Key, s.secrets[r.ID.Short()].X)
}
