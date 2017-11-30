package service

import (
	"crypto/cipher"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1"

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
	local := onet.NewTCPTest()

	hosts, _, _ := local.GenTree(1, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	admin := &stamp{123, true, 0}
	voter := &stamp{654, false, 0}
	services[0].state = &state{map[string]*stamp{"0": admin, "1": voter}}

	u, _ := services[0].assertLevel("0", false)
	assert.Equal(t, chains.User(123), u)
	u, _ = services[0].assertLevel("0", true)
	assert.Equal(t, chains.User(123), u)
	u, _ = services[0].assertLevel("1", false)
	assert.Equal(t, chains.User(654), u)
	_, err := services[0].assertLevel("2", true)
	assert.NotNil(t, err)
}

func TestPing(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, _, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	ping := &api.Ping{0}

	p1, _ := services[0].Ping(ping)
	p2, _ := services[1].Ping(ping)
	p3, _ := services[2].Ping(ping)

	assert.Equal(t, uint32(1), p1.Nonce, p2.Nonce, p3.Nonce)
}

func TestLink(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	// Probe request
	lr, err := services[0].Link(&api.Link{"", nil, nil, nil})
	assert.Nil(t, err)
	assert.Nil(t, lr.Master)

	// Wrong pin
	lr, err = services[0].Link(&api.Link{"000000", nil, nil, nil})
	assert.Nil(t, lr)
	assert.NotNil(t, err)

	// Valid link
	lr, err = services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	assert.NotNil(t, lr.Master)
	assert.Nil(t, err)

}

func TestOpen(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	admin := &stamp{123, true, 0}
	voter := &stamp{654, false, 0}
	services[0].state = &state{map[string]*stamp{"0": admin, "1": voter}}

	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})

	// Not logged in
	or, err := services[0].Open(&api.Open{"", nil, nil})
	assert.Nil(t, or)
	assert.NotNil(t, err)

	// Not admin
	or, err = services[0].Open(&api.Open{"1", nil, nil})
	assert.Nil(t, or)
	assert.NotNil(t, err)

	// Invalid master
	or, err = services[0].Open(&api.Open{"0", nil, nil})
	assert.Nil(t, or)
	assert.NotNil(t, err)

	// Valid generation
	e := &chains.Election{"", 123, []chains.User{654}, roster, nil, nil, "", ""}
	or, err = services[0].Open(&api.Open{"0", lr.Master, e})
	assert.NotNil(t, or)
	assert.Nil(t, err)

	<-time.After(200 * time.Millisecond)

	// Check equality of dkg key
	pk1 := services[0].secrets[string(or.Genesis)].X
	pk2 := services[1].secrets[string(or.Genesis)].X
	pk3 := services[2].secrets[string(or.Genesis)].X
	assert.Equal(t, pk1.String(), pk2.String(), pk3.String())
}

func TestLogin(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	admin := &stamp{123456, true, 0}
	services[0].state = &state{map[string]*stamp{"0": admin}}

	e := &chains.Election{"", 123, []chains.User{654}, roster, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})

	<-time.After(200 * time.Millisecond)

	// Invalid master
	lor, err := services[0].Login(&api.Login{nil, 654, nil})
	assert.Nil(t, lor)
	assert.NotNil(t, err)

	// Valid login
	lor, err = services[0].Login(&api.Login{lr.Master, 654, nil})
	assert.NotNil(t, lor)
	assert.Nil(t, err)

	assert.Equal(t, 32, len(lor.Token))
	assert.Equal(t, 1, len(lor.Elections))
	assert.Equal(t, e.Name, lor.Elections[0].Name)
	assert.Equal(t, or.Key.String(), lor.Elections[0].Key.String())
}

func TestCast(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	admin := &stamp{123, true, 0}
	user1 := &stamp{654, false, 0}
	user2 := &stamp{789, false, 0}
	services[0].state = &state{map[string]*stamp{"0": admin, "1": user1, "2": user2}}

	e := &chains.Election{"", 123, []chains.User{654}, roster, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})

	// Not logged in
	cr, err := services[0].Cast(&api.Cast{"", or.Genesis, nil})
	assert.Nil(t, cr)
	assert.NotNil(t, err)

	// Invalid genesis
	cr, err = services[0].Cast(&api.Cast{"0", nil, nil})
	assert.Nil(t, cr)
	assert.NotNil(t, err)

	// Invalid user
	cr, err = services[0].Cast(&api.Cast{"2", or.Genesis, nil})
	assert.Nil(t, cr)
	assert.NotNil(t, err)

	// Valid cast
	cr, err = services[0].Cast(&api.Cast{"1", or.Genesis, &api.Ballot{}})
	assert.NotNil(t, cr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), cr.Index)
}

func TestFinalize(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	admin := &stamp{123, true, 0}
	user1 := &stamp{654, false, 0}
	user2 := &stamp{789, false, 0}
	services[0].state = &state{map[string]*stamp{"0": admin, "1": user1, "2": user2}}

	e := &chains.Election{"", 123, []chains.User{654}, roster, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})

	// Not logged in
	fr, err := services[0].Finalize(&api.Finalize{"", or.Genesis})
	assert.Nil(t, fr)
	assert.NotNil(t, err)
}

func castServices(services []onet.Service) []*Service {
	cast := make([]*Service, len(services))
	for i, service := range services {
		cast[i] = service.(*Service)
	}

	return cast
}

func encrypt(suite abstract.Suite, pub abstract.Point, msg []byte) (K, C abstract.Point) {
	M, _ := suite.Point().Pick(msg, random.Stream)

	k := suite.Scalar().Pick(random.Stream)
	K = suite.Point().Mul(nil, k)
	S := suite.Point().Mul(pub, k)
	C = S.Add(S, M)

	return
}
