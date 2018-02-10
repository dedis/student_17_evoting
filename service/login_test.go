package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dedis/onet"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestLogin_InvalidMasterID(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	_, err := s.Login(&api.Login{ID: nil})
	assert.NotNil(t, err)
}

func TestLogin_InvalidLink(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	master := &chains.Master{Roster: roster}
	master.GenChain([]byte{})

	_, err := s.Login(&api.Login{ID: master.ID})
	assert.NotNil(t, err)
}

func TestLogin_InvalidSignature(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	x, X := crypto.RandomKeyPair()
	master := &chains.Master{Roster: roster, Key: X}
	master.GenChain()

	l := &api.Login{User: 0, ID: master.ID}
	l.Sign(x)
	l.Signature = append(l.Signature, byte(0))

	_, err := s.Login(l)
	assert.NotNil(t, err)
}

func TestLogin_Full(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)

	election := &chains.Election{
		Roster:  roster,
		Creator: 0,
		Users:   []uint32{0},
		Stage:   chains.RUNNING,
	}
	_ = election.GenChain(3)

	x, X := crypto.RandomKeyPair()
	master := &chains.Master{Roster: roster, Key: X}
	master.GenChain(election.ID)

	l := &api.Login{User: 0, ID: master.ID}
	l.Sign(x)

	r, _ := s.Login(l)
	assert.Equal(t, election.ID, r.Elections[0].ID)
	assert.Equal(t, uint32(0), s.state.log[r.Token].user)
}
