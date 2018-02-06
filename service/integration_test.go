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
	reply, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	assert.NotEqual(t, 0, len(reply.Master))
}

func TestOpen(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	reply, _ := service.Open(&api.Open{lor.Token, lr.Master, election})
	assert.NotEqual(t, 0, len(reply.Genesis))
}

func TestLogin(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	assert.Equal(t, 1, len(reply.Elections))
}

func TestCast(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Cast(&api.Cast{lor.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.Equal(t, 2, int(reply.Index))
}

func TestGetBox(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	service.Cast(&api.Cast{lor.Token, or.Genesis, &chains.Ballot{User: 0}})

	gbr, _ := service.GetBox(&api.GetBox{lor.Token, or.Genesis})
	assert.Equal(t, 1, len(gbr.Box.Ballots))
}

func TestShuffle(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	lor2, _ := service.Login(&api.Login{lr.Master, 2, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
	b2 := &chains.Ballot{2, suite.Point(), suite.Point()}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
	service.Cast(&api.Cast{lor2.Token, or.Genesis, b2})

	reply, _ := service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
	assert.NotNil(t, reply)
}

func TestGetMixes(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})

	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})

	reply, _ := service.GetMixes(&api.GetMixes{lor0.Token, or.Genesis})
	assert.Equal(t, 3, len(reply.Mixes))
}

// func TestDecrypt(t *testing.T) {
// 	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
// 	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
// 	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

// 	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
// 	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

// 	k0, c0 := crypto.Encrypt(or.Key, []byte{0})
// 	k1, c1 := crypto.Encrypt(or.Key, []byte{1})
// 	b0, b1 := &chains.Ballot{0, k0, c0}, &chains.Ballot{1, k1, c1}
// 	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
// 	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
// 	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})

// 	r, _ := service.Decrypt(&api.Decrypt{lor0.Token, or.Genesis})
// 	assert.Nil(t, r.Decrypted.Ballots)
// 	assert.Equal(t, byte(r.Decrypted.Texts[0].User), r.Decrypted.Texts[0].Data[0])
// 	assert.Equal(t, byte(r.Decrypted.Texts[1].User), r.Decrypted.Texts[1].Data[0])
// }

// func TestAggregate(t *testing.T) {
// 	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
// 	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
// 	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

// 	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
// 	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

// 	b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
// 	b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
// 	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
// 	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
// 	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
// 	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
// 	service.Decrypt(&api.Decrypt{lor0.Token, or.Genesis})

// 	r0, _ := service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 0})
// 	r1, _ := service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 1})
// 	r2, _ := service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 2})
// 	assert.Equal(t, 2, len(r0.Box.Ballots), len(r1.Box.Ballots), len(r2.Box.Ballots))
// }
