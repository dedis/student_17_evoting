package service

import (
	"crypto/cipher"
	"encoding/base64"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
)

var suite abstract.Suite
var stream cipher.Stream

func init() {
	suite = ed25519.NewAES128SHA256Ed25519(false)
	stream = suite.Cipher(abstract.RandomKey)
}

func TestMain(m *testing.M) {
	log.MainTest(m)
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
	_, err := services[0].Link(&api.Link{"", nil, nil, nil})
	assert.Nil(t, err)

	// Wrong pin
	_, err = services[0].Link(&api.Link{"000000", nil, nil, nil})
	assert.NotNil(t, err)

	// Valid link
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	assert.NotNil(t, lr.Master)
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

	// Valid generation
	e := &chains.Election{"", 123, []chains.User{654}, roster, nil, nil, "", ""}
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})
	<-time.After(200 * time.Millisecond)

	// Check equality of dkg key
	id, _ := base64.StdEncoding.DecodeString(or.Genesis)
	pk1 := services[0].secrets[string(id)].X
	pk2 := services[1].secrets[string(id)].X
	pk3 := services[2].secrets[string(id)].X
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

	// Valid login
	lor, _ := services[0].Login(&api.Login{lr.Master, 654, nil})
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

	// Valid cast
	cr, _ := services[0].Cast(&api.Cast{"1", or.Genesis, &chains.Ballot{}})
	assert.Equal(t, uint32(2), cr.Index)
}

func TestAggregate(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	admin := &stamp{123, true, 0}
	user1 := &stamp{654, false, 0}
	user2 := &stamp{789, false, 0}
	services[0].state = &state{map[string]*stamp{"0": admin, "1": user1, "2": user2}}

	e := &chains.Election{"", 123, []chains.User{654, 789}, roster, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})

	b1 := &chains.Ballot{654, suite.Point(), suite.Point(), nil}
	b2 := &chains.Ballot{789, suite.Point(), suite.Point(), nil}
	b3 := &chains.Ballot{654, suite.Point(), suite.Point(), nil}
	services[0].Cast(&api.Cast{"1", or.Genesis, b1})
	services[0].Cast(&api.Cast{"1", or.Genesis, b2})
	services[0].Cast(&api.Cast{"1", or.Genesis, b3})

	// Valid aggregation (ballots)
	ar, _ := services[0].Aggregate(&api.Aggregate{"1", or.Genesis, chains.BALLOTS})
	assert.Equal(t, 2, len(ar.Box.Ballots))

	// Valid aggregation (shuffle)
	services[0].Finalize(&api.Finalize{"0", or.Genesis})
	ar, _ = services[0].Aggregate(&api.Aggregate{"1", or.Genesis, chains.SHUFFLE})
	assert.Equal(t, 2, len(ar.Box.Ballots))
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

	e := &chains.Election{"", 123, []chains.User{654, 789}, roster, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, _ := services[0].Open(&api.Open{"0", lr.Master, e})

	a1, b1 := encrypt(or.Key, []byte{1, 2, 3})
	a2, b2 := encrypt(or.Key, []byte{1, 2, 3})
	ballot1 := &chains.Ballot{654, a1, b1, nil}
	ballot2 := &chains.Ballot{789, a2, b2, nil}
	services[0].Cast(&api.Cast{"1", or.Genesis, ballot1})
	services[0].Cast(&api.Cast{"2", or.Genesis, ballot2})

	// Valid finalize
	fr, _ := services[0].Finalize(&api.Finalize{"0", or.Genesis})
	assert.Equal(t, []byte{1, 2, 3}, fr.Decryption.Ballots[0].Text)
	assert.Equal(t, []byte{1, 2, 3}, fr.Decryption.Ballots[1].Text)

	// Invalid second finalize
	fr, err := services[0].Finalize(&api.Finalize{"0", or.Genesis})
	assert.NotNil(t, err)
}

func Test100(t *testing.T) {
	t.Skip()
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))
	services[0].pin = "123456"

	logging := make(map[string]*stamp)
	logging["0"] = &stamp{0, true, 0}
	users := make([]chains.User, 0)
	for i := 1; i <= 200; i++ {
		logging[strconv.Itoa(i)] = &stamp{chains.User(i), false, 0}
		users = append(users, chains.User(i))
	}
	services[0].state = &state{logging}

	log.Lvl2("Prepare skipchain...")

	e := &chains.Election{"", 0, users, roster, nil, nil, "", ""}
	lr, _ := services[0].Link(&api.Link{"123456", roster, suite.Point(), nil})
	or, err := services[0].Open(&api.Open{"0", lr.Master, e})

	for i := 1; i <= 200; i++ {
		a, b := encrypt(or.Key, []byte{byte(i)})
		ballot := &chains.Ballot{chains.User(i), a, b, nil}
		services[0].Cast(&api.Cast{strconv.Itoa(i), or.Genesis, ballot})
	}

	log.Lvl2("Start finalize...")
	start := time.Now()

	fr, err := services[0].Finalize(&api.Finalize{"0", or.Genesis})
	assert.Nil(t, err)

	for _, ballot := range fr.Decryption.Ballots {
		assert.Equal(t, byte(ballot.User), ballot.Text[0])
	}

	elapsed := time.Since(start)
	log.Printf("Finalized finished: %s", elapsed)
}

func encrypt(key abstract.Point, msg []byte) (K, C abstract.Point) {
	M, _ := suite.Point().Pick(msg, random.Stream)

	k := suite.Scalar().Pick(random.Stream)
	K = suite.Point().Mul(nil, k)
	S := suite.Point().Mul(key, k)
	C = S.Add(S, M)
	return
}

func castServices(services []onet.Service) []*Service {
	cast := make([]*Service, len(services))
	for i, service := range services {
		cast[i] = service.(*Service)
	}

	return cast
}
