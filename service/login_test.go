package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
)

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
