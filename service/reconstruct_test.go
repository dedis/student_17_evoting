package service

import (
	"sort"
	"testing"

	"github.com/dedis/onet"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestReconstruct_UserNotLoggedIn(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	nodes, _, _ := local.GenBigTree(3, 3, 1, true)
	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: false}

	_, err := s.Reconstruct(&api.Reconstruct{Token: ""})
	assert.Equal(t, ERR_NOT_LOGGED_IN, err)
}

func TestReconstruct_ElectionNotDecrypted(t *testing.T) {
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

	_, err := s.Reconstruct(&api.Reconstruct{Token: "0", ID: election.ID})
	assert.Equal(t, ERR_NOT_DECRYPTED, err)
}

func TestReconstruct_Full(t *testing.T) {
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
	_ = election.GenChain(7)

	r, _ := s.Reconstruct(&api.Reconstruct{Token: "0", ID: election.ID})
	assert.Equal(t, 7, len(r.Points))

	messages := make([]int, 7)
	for i, point := range r.Points {
		data, _ := point.Data()
		messages[i] = int(data[0])
	}
	sort.Ints(messages)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6}, messages)
}
