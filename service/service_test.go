package service

import (
	"crypto/cipher"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
)

var suite abstract.Suite
var stream cipher.Stream

func init() {
	suite = ed25519.NewAES128SHA256Ed25519(false)
	stream = suite.Cipher(abstract.RandomKey)
}

func TestAssertLevel(t *testing.T) {
	local, services, _, _ := setup(3)
	defer local.CloseAll()

	u, _ := services[0].assertLevel("0", false)
	assert.Equal(t, chains.User(0), u)
	u, _ = services[0].assertLevel("0", true)
	assert.Equal(t, chains.User(0), u)
	u, _ = services[0].assertLevel("1", false)
	assert.Equal(t, chains.User(1), u)
	_, err := services[0].assertLevel("10", true)
	assert.NotNil(t, err)
}

func TestFetchElection(t *testing.T) {
	local, services, _, election := setup(3)
	defer local.CloseAll()

	// Invalid id
	_, _, err := services[0].fetchElection("", 1, false)
	assert.NotNil(t, err)

	// Not the creator
	_, _, err = services[0].fetchElection(election, 1, true)
	assert.NotNil(t, err)

	e, _, _ := services[0].fetchElection(election, 1, false)
	assert.Equal(t, "election", e.Name)
}

func TestFetchMaster(t *testing.T) {
	local, services, master, _ := setup(3)
	defer local.CloseAll()

	// Invalid id
	_, _, err := services[0].fetchMaster("")
	assert.NotNil(t, err)

	m, _, _ := services[0].fetchMaster(master)
	assert.Equal(t, chains.User(0), m.Admins[0])
}

func TestPing(t *testing.T) {
	local, services, _, _ := setup(3)
	defer local.CloseAll()

	ping := &api.Ping{0}

	p1, _ := services[0].Ping(ping)
	p2, _ := services[1].Ping(ping)
	p3, _ := services[2].Ping(ping)
	assert.Equal(t, uint32(1), p1.Nonce, p2.Nonce, p3.Nonce)
}

func TestLink(t *testing.T) {
	local, services, _, _ := setup(3)
	defer local.CloseAll()

	_, err := services[0].Link(&api.Link{"", nil, nil, nil})
	assert.Nil(t, err)

	_, err = services[0].Link(&api.Link{"000000", nil, nil, nil})
	assert.NotNil(t, err)

	lr, _ := services[0].Link(&api.Link{"123456", services[0].node, suite.Point(), nil})
	assert.NotNil(t, lr.Master)
}

func TestOpen(t *testing.T) {
	local, services, master, _ := setup(3)
	defer local.CloseAll()

	// Valid generation
	e := &chains.Election{"", 0, []chains.User{}, services[0].node, nil, nil, "", ""}
	or, err := services[0].Open(&api.Open{"0", master, e})
	assert.Nil(t, err)
	<-time.After(200 * time.Millisecond)

	// Check equality of dkg key
	id, _ := base64.StdEncoding.DecodeString(or.Genesis)
	pk1 := services[0].secrets[string(id)].X
	pk2 := services[1].secrets[string(id)].X
	pk3 := services[2].secrets[string(id)].X
	assert.Equal(t, pk1.String(), pk2.String(), pk3.String())
}

func TestLogin(t *testing.T) {
	local, services, master, _ := setup(3)
	defer local.CloseAll()

	// Valid login
	lor, _ := services[0].Login(&api.Login{master, 1, nil})
	assert.Equal(t, 32, len(lor.Token))
	assert.Equal(t, 1, len(lor.Elections))
}

func TestCast(t *testing.T) {
	local, services, _, genesis := setup(3)
	defer local.CloseAll()

	cr, err := services[0].Cast(&api.Cast{"3", genesis, &chains.Ballot{}})
	assert.Nil(t, err)
	assert.Equal(t, uint32(5), cr.Index)
}

func TestAggregate(t *testing.T) {
	local, services, _, genesis := setup(3)
	defer local.CloseAll()

	ar, err := services[0].Aggregate(&api.Aggregate{"1", genesis, chains.BALLOTS})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(ar.Box.Ballots))
}

func TestFinalize(t *testing.T) {
	local, services, _, genesis := setup(3)
	defer local.CloseAll()

	fr, err := services[0].Finalize(&api.Finalize{"0", genesis})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(fr.Shuffle.Ballots))
	assert.Equal(t, 3, len(fr.Decryption.Ballots))
}
