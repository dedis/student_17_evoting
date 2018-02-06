package service

import (
	"testing"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/stretchr/testify/assert"
)

func TestLinkError(t *testing.T) {
	reply, _ := service.Link(&api.Link{"", roster, nil, []chains.User{0}})
	assert.Equal(t, 0, len(reply.Master))

	_, err := service.Link(&api.Link{"1", roster, nil, []chains.User{0}})
	assert.NotNil(t, err)
}

func TestOpenError(t *testing.T) {
	_, err := service.Open(&api.Open{"", nil, nil})
	assert.NotNil(t, err)

	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})

	_, err = service.Open(&api.Open{lor.Token, nil, nil})
	assert.NotNil(t, err)
}

func TestLoginError(t *testing.T) {
	_, err := service.Login(&api.Login{nil, 0, nil})
	assert.NotNil(t, err)
}

func TestCastError(t *testing.T) {
	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})
	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
	// lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
	lor2, _ := service.Login(&api.Login{lr.Master, 2, []byte{}})

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

	_, err := service.Cast(&api.Cast{"", or.Genesis, &chains.Ballot{User: 0}})
	assert.NotNil(t, err)

	_, err = service.Cast(&api.Cast{lor0.Token, nil, &chains.Ballot{User: 0}})
	assert.NotNil(t, err)

	_, err = service.Cast(&api.Cast{lor2.Token, or.Genesis, &chains.Ballot{User: 0}})
	assert.NotNil(t, err)

	_, err = service.Cast(&api.Cast{lor0.Token, or.Genesis, &chains.Ballot{User: 1}})
	assert.NotNil(t, err)

	// b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
	// b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
	// service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
	// service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
	// service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
	// _, err = service.Cast(&api.Cast{lor0.Token, or.Genesis, &chains.Ballot{User: 0}})
	// assert.NotNil(t, err)
}

// func TestShuffleError(t *testing.T) {
// 	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0, 1}})
// 	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
// 	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

// 	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
// 	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

// 	b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
// 	b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
// 	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
// 	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
// 	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})

// 	_, err := service.Shuffle(&api.Shuffle{"", or.Genesis})
// 	assert.NotNil(t, err)

// 	_, err = service.Shuffle(&api.Shuffle{lor0.Token, nil})
// 	assert.NotNil(t, err)

// 	_, err = service.Shuffle(&api.Shuffle{lor1.Token, or.Genesis})
// 	assert.NotNil(t, err)

// 	_, err = service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
// 	assert.NotNil(t, err)
// }

// func TestDecryptError(t *testing.T) {
// 	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0, 1}})
// 	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
// 	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})

// 	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
// 	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

// 	b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
// 	b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
// 	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
// 	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})
// 	service.Shuffle(&api.Shuffle{lor0.Token, or.Genesis})
// 	service.Decrypt(&api.Decrypt{lor0.Token, or.Genesis})

// 	_, err := service.Decrypt(&api.Decrypt{"", or.Genesis})
// 	assert.NotNil(t, err)

// 	_, err = service.Decrypt(&api.Decrypt{lor0.Token, nil})
// 	assert.NotNil(t, err)

// 	_, err = service.Decrypt(&api.Decrypt{lor1.Token, or.Genesis})
// 	assert.NotNil(t, err)

// 	_, err = service.Decrypt(&api.Decrypt{lor0.Token, or.Genesis})
// 	assert.NotNil(t, err)
// }

// func TestAggregateError(t *testing.T) {
// 	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0, 1}})
// 	lor0, _ := service.Login(&api.Login{lr.Master, 0, []byte{}})
// 	lor1, _ := service.Login(&api.Login{lr.Master, 1, []byte{}})
// 	lor2, _ := service.Login(&api.Login{lr.Master, 2, []byte{}})

// 	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0, 1}}
// 	or, _ := service.Open(&api.Open{lor0.Token, lr.Master, election})

// 	b0 := &chains.Ballot{0, suite.Point(), suite.Point()}
// 	b1 := &chains.Ballot{1, suite.Point(), suite.Point()}
// 	service.Cast(&api.Cast{lor0.Token, or.Genesis, b0})
// 	service.Cast(&api.Cast{lor1.Token, or.Genesis, b1})

// 	_, err := service.Aggregate(&api.Aggregate{"", or.Genesis, 0})
// 	assert.NotNil(t, err)

// 	_, err = service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 3})
// 	assert.NotNil(t, err)

// 	_, err = service.Aggregate(&api.Aggregate{lor0.Token, nil, 0})
// 	assert.NotNil(t, err)

// 	_, err = service.Aggregate(&api.Aggregate{lor2.Token, or.Genesis, 0})
// 	assert.NotNil(t, err)

// 	_, err = service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 1})
// 	assert.NotNil(t, err)

// 	_, err = service.Aggregate(&api.Aggregate{lor0.Token, or.Genesis, 2})
// 	assert.NotNil(t, err)
// }
