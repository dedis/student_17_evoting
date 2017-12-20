package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
)

func TestPing(t *testing.T) {
	reply, _ := service.Ping(&api.Ping{0})
	assert.Equal(t, 1, int(reply.Nonce))
}

func TestLink(t *testing.T) {
	reply, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	assert.NotEqual(t, 0, len(reply.Master))
}

func TestOpen(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	reply, _ := service.Open(&api.Open{lor.Token, lr.Master, election})
	assert.NotEqual(t, 0, len(reply.Genesis))
}

func TestLogin(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	assert.Equal(t, 1, len(reply.Elections))
}

func TestCast(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Cast(&api.Cast{lor.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.Equal(t, 2, int(reply.Index))
}

func TestShuffle(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point(), nil}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point(), nil}
	service.Cast(&api.Cast{lor.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor.Token, or.Genesis, b1})

	reply, _ := service.Shuffle(&api.Shuffle{lor.Token, or.Genesis})
	assert.Equal(t, 2, len(reply.Shuffled.Ballots))
}

func TestDecrypt(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	k0, c0 := encrypt(or.Key, []byte{0})
	k1, c1 := encrypt(or.Key, []byte{1})
	b0, b1 := &chains.Ballot{0, k0, c0, nil}, &chains.Ballot{1, k1, c1, nil}
	service.Cast(&api.Cast{lor.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor.Token, or.Genesis, b1})
	service.Shuffle(&api.Shuffle{lor.Token, or.Genesis})

	r, _ := service.Decrypt(&api.Decrypt{lor.Token, or.Genesis})
	assert.Equal(t, byte(r.Decrypted.Ballots[0].User), r.Decrypted.Ballots[0].Text[0])
	assert.Equal(t, byte(r.Decrypted.Ballots[1].User), r.Decrypted.Ballots[1].Text[0])
}

func TestAggregate(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point(), nil}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point(), nil}
	service.Cast(&api.Cast{lor.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor.Token, or.Genesis, b1})
	service.Shuffle(&api.Shuffle{lor.Token, or.Genesis})
	service.Decrypt(&api.Decrypt{lor.Token, or.Genesis})

	r0, _ := service.Aggregate(&api.Aggregate{lor.Token, or.Genesis, 0})
	r1, _ := service.Aggregate(&api.Aggregate{lor.Token, or.Genesis, 1})
	r2, _ := service.Aggregate(&api.Aggregate{lor.Token, or.Genesis, 2})
	assert.Equal(t, 2, len(r0.Box.Ballots), len(r1.Box.Ballots), len(r2.Box.Ballots))
}
