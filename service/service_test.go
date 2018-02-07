package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

var serviceID onet.ServiceID

func init() {
	serviceID, _ = onet.RegisterNewService(Name, new)
}

func TestPing(t *testing.T) {
	local, service, _, _ := setup(0)
	defer local.CloseAll()

	reply, _ := service.Ping(&api.Ping{Nonce: 0})
	assert.Equal(t, 1, int(reply.Nonce))
}

func TestLink(t *testing.T) {
	local, service, master, _ := setup(0)
	defer local.CloseAll()

	l := &api.Link{Pin: service.pin, Roster: master.Roster, Admins: []uint32{0}}
	r, _ := service.Link(l)
	assert.NotEqual(t, 0, len(r.Master))
}

func TestOpen(t *testing.T) {
	local, service, master, _ := setup(0)
	defer local.CloseAll()

	e := &chains.Election{Creator: 0, Users: []uint32{0}, Data: []byte{}}
	r, _ := service.Open(&api.Open{Token: "0", Master: master.ID, Election: e})
	assert.NotEqual(t, 0, len(r.Genesis))
}

func TestLogin(t *testing.T) {
	local, service, master, _ := setup(chains.STAGE_RUNNING)
	defer local.CloseAll()

	r, _ := service.Login(&api.Login{Master: master.ID, User: 1, Signature: []byte{}})
	assert.Equal(t, 1, len(r.Elections))
}

func TestCast(t *testing.T) {
	local, service, _, election := setup(chains.STAGE_RUNNING)
	defer local.CloseAll()

	c := &api.Cast{Token: "0", Genesis: election.ID, Ballot: &chains.Ballot{User: 0}}
	r, _ := service.Cast(c)
	assert.NotNil(t, r)
}

func TestGetBox(t *testing.T) {
	local, service, _, election := setup(chains.STAGE_RUNNING)
	defer local.CloseAll()

	r, _ := service.GetBox(&api.GetBox{Token: "0", Genesis: election.ID})
	assert.Equal(t, 2, len(r.Box.Ballots))
}

func TestShuffle(t *testing.T) {
	local, service, _, election := setup(chains.STAGE_RUNNING)
	defer local.CloseAll()

	r, _ := service.Shuffle(&api.Shuffle{Token: "0", Genesis: election.ID})
	assert.NotNil(t, r)
}

func TestGetMixes(t *testing.T) {
	local, service, _, election := setup(chains.STAGE_SHUFFLED)
	defer local.CloseAll()

	r, _ := service.GetMixes(&api.GetMixes{Token: "0", Genesis: election.ID})
	assert.Equal(t, 3, len(r.Mixes))
}

func setup(stage int) (*onet.LocalTest, *Service, *chains.Master, *chains.Election) {
	local := onet.NewLocalTest()

	nodes, roster, _ := local.GenBigTree(3, 3, 1, true)

	chain, _ := chains.New(roster, nil)
	e := &chains.Election{
		ID:      chain.Hash,
		Roster:  roster,
		Key:     crypto.Random(),
		Creator: 0,
		Users:   []uint32{0, 1},
		Data:    []byte{},
	}
	b1 := &chains.Ballot{User: 0, Alpha: crypto.Random(), Beta: crypto.Random()}
	b2 := &chains.Ballot{User: 1, Alpha: crypto.Random(), Beta: crypto.Random()}
	box := &chains.Box{Ballots: []*chains.Ballot{b1, b2}}
	mix := &chains.Mix{Ballots: []*chains.Ballot{b1, b2}, Proof: []byte{}}

	if stage == chains.STAGE_RUNNING {
		chains.Store(e.Roster, e.ID, e, b1, b2)
	} else if stage == chains.STAGE_SHUFFLED {
		chains.Store(e.Roster, e.ID, e, b1, b2, box, mix, mix, mix)
	} else if stage == chains.STAGE_DECRYPTED {

	}

	chain, _ = chains.New(roster, nil)
	m := &chains.Master{ID: chain.Hash, Roster: roster, Admins: []uint32{0}}
	chains.Store(m.Roster, m.ID, m, &chains.Link{Genesis: e.ID})

	s := local.GetServices(nodes, serviceID)[0].(*Service)
	s.state.log["0"] = &stamp{user: 0, admin: true}
	return local, s, m, e
}
