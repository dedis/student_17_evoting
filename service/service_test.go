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

var nodes []*onet.Server
var roster *onet.Roster
var service *Service

func TestMain(m *testing.M) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	serviceID, _ = onet.RegisterNewService(Name, new)

	nodes, roster, _ = local.GenTree(3, true)
	service = local.GetServices(nodes, serviceID)[0].(*Service)

	m.Run()
}

func TestPing(t *testing.T) {
	reply, _ := service.Ping(&api.Ping{0})
	assert.Equal(t, 1, int(reply.Nonce))
}

func TestLink(t *testing.T) {
	reply, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	assert.NotEqual(t, 0, len(reply.Master))
}

func TestOpen(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Creator: 0, Users: []uint32{0}, Data: []byte{}}
	reply, _ := service.Open(&api.Open{lor.Token, lr.Master, election})
	assert.NotEqual(t, 0, len(reply.Genesis))
}

func TestLogin(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Creator: 0, Users: []uint32{0, 1}, Data: []byte{}}
	service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	assert.Equal(t, 1, len(reply.Elections))
}

func TestCast(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Creator: 0, Users: []uint32{0}, Data: []byte{}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Cast(&api.Cast{lor.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.NotNil(t, reply)
}

func TestGetBox(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Creator: 0, Users: []uint32{0}, Data: []byte{}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	service.Cast(&api.Cast{lor.Token, or.Genesis, &chains.Ballot{User: 0}})

	gbr, _ := service.GetBox(&api.GetBox{lor.Token, or.Genesis})
	assert.Equal(t, 1, len(gbr.Box.Ballots))
}

func TestShuffle(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

	election := &chains.Election{Creator: 0, Users: []uint32{0, 1}, Data: []byte{}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, crypto.Random(), crypto.Random()}
	b1 := &chains.Ballot{1, crypto.Random(), crypto.Random()}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})

	reply, _ := service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
	assert.NotNil(t, reply)
}

func TestGetMixes(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []uint32{0}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

	election := &chains.Election{Creator: 0, Users: []uint32{0, 1}, Data: []byte{}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, crypto.Random(), crypto.Random()}
	b1 := &chains.Ballot{1, crypto.Random(), crypto.Random()}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})

	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})

	reply, _ := service.GetMixes(&api.GetMixes{lor0.Token, or.Genesis})
	assert.Equal(t, 3, len(reply.Mixes))
}
