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

func TestLinkError(t *testing.T) {
	reply, _ := service.Link(&api.Link{"", roster, nil, []chains.User{0}})
	assert.Equal(t, "", reply.Master)

	_, err := service.Link(&api.Link{"1", roster, nil, []chains.User{0}})
	assert.NotNil(t, err)
}

func TestLink(t *testing.T) {
	reply, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	assert.NotEqual(t, 0, len(reply.Master))
}

func TestOpenError(t *testing.T) {
	_, err := service.Open(&api.Open{"", "", nil})
	assert.NotNil(t, err)

	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	_, err = service.Open(&api.Open{lor.Token, "", nil})
	assert.NotNil(t, err)
}

func TestOpen(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	reply, _ := service.Open(&api.Open{lor.Token, lr.Master, election})
	assert.NotEqual(t, 0, len(reply.Genesis))
}

func TestLoginError(t *testing.T) {
	_, err := service.Login(&api.Login{"", 0, []byte{}})
	assert.NotNil(t, err)
}

func TestLogin(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	assert.Equal(t, 1, len(reply.Elections))
}

func TestCastError(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	lor2, _ := service.Login(&api.Login{lr.Master, 2, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	_, err := service.Cast(&api.Cast{"", or.Genesis, &chains.Ballot{User: 0}})
	assert.NotNil(t, err)

	_, err = service.Cast(&api.Cast{lor0.Token, "", &chains.Ballot{User: 0}})
	assert.NotNil(t, err)

	_, err = service.Cast(&api.Cast{lor2.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.NotNil(t, err)

	b0 := &chains.Ballot{0, suite.Point(), suite.Point(), nil}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point(), nil}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
	_, err = service.Cast(&api.Cast{lor0.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.NotNil(t, err)
}

func TestCast(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	or, _ := service.Open(&api.Open{lor.Token, lr.Master, election})

	reply, _ := service.Cast(&api.Cast{lor.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.Equal(t, 2, int(reply.Index))
}

func TestShuffleError(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0, 1}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point(), nil}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point(), nil}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})

	_, err := service.Shuffle(&api.Shuffle{"", or.Genesis})
	assert.NotNil(t, err)

	_, err = service.Shuffle(&api.Shuffle{lor0.Token, ""})
	assert.NotNil(t, err)

	_, err = service.Shuffle(&api.Shuffle{lor1.Token, or.Genesis})
	assert.NotNil(t, err)

	_, err = service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
	assert.NotNil(t, err)
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

func TestDecryptError(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0, 1}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point(), nil}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point(), nil}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
	service.Decrypt(&api.Decrypt{lor0.Token, or.Genesis})

	_, err := service.Decrypt(&api.Decrypt{"", or.Genesis})
	assert.NotNil(t, err)

	_, err = service.Decrypt(&api.Decrypt{lor0.Token, ""})
	assert.NotNil(t, err)

	_, err = service.Decrypt(&api.Decrypt{lor1.Token, or.Genesis})
	assert.NotNil(t, err)

	_, err = service.Decrypt(&api.Decrypt{lor0.Token, or.Genesis})
	assert.NotNil(t, err)
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

func TestAggregateError(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.Pin, roster, nil, []chains.User{0, 1}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	lor2, _ := service.Login(&api.Login{lr.Master, 2, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	b0 := &chains.Ballot{0, suite.Point(), suite.Point(), nil}
	b1 := &chains.Ballot{1, suite.Point(), suite.Point(), nil}
	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})

	_, err := service.Aggregate(&api.Aggregate{"", or.Genesis, 0})
	assert.NotNil(t, err)

	_, err = service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 3})
	assert.NotNil(t, err)

	_, err = service.Aggregate(&api.Aggregate{lor0.Token, "", 0})
	assert.NotNil(t, err)

	_, err = service.Aggregate(&api.Aggregate{lor2.Token, or.Genesis, 0})
	assert.NotNil(t, err)

	_, err = service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 1})
	assert.NotNil(t, err)

	_, err = service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 2})
	assert.NotNil(t, err)
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
