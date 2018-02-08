package service

import (
	"testing"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/api"
	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/chains"
)

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

	client := skipchain.NewClient()
	chain, _ := client.GetUpdateChain(roster, r.ID)
	_, blob, _ := network.Unmarshal(chain.Update[1].Data)
	assert.Equal(t, r.ID, blob.(*chains.Master).ID)
}
