package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
)

func TestGetBox_NotLoggedIn(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.GetBox(&api.GetBox{Token: ""})
	assert.NotNil(t, err)
}

func TestGetBox_NotPart(t *testing.T) {
	local := onet.NewLocalTest()
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

	_, err := s.GetBox(&api.GetBox{Token: "1", ID: election.ID})
	assert.NotNil(t, err)
}

func TestGetBox_Full(t *testing.T) {
	local := onet.NewLocalTest()
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

	r, _ := s.GetBox(&api.GetBox{Token: "0", ID: election.ID})
	assert.Equal(t, 3, len(r.Box.Ballots))
}
