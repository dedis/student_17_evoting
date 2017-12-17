package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/service"
)

type client struct {
	*onet.Client
}

func new() *client {
	return &client{onet.NewClient(service.Name)}
}

func (c *client) send(r *onet.Roster, msg, rep interface{}) onet.ClientError {
	return c.SendProtobuf(r.List[0], msg, rep)
}

func TestPing(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	_, roster, _ := local.GenTree(3, true)

	for _ = range roster.List {
		reply := &api.Ping{}
		require.Nil(t, new().send(roster, &api.Ping{0}, reply))
		assert.Equal(t, uint32(1), reply.Nonce)
	}
}

func TestLink(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	hosts, roster, _ := local.GenTree(3, true)
	services := local.GetServices(hosts, service.ServiceID)

	link := &api.Link{services[0].(*service.Service).Pin, roster, nil, []chains.User{0}}
	reply := &api.LinkReply{}

	require.Nil(t, new().send(roster, link, reply))
	assert.NotEqual(t, 0, len(reply.Master))
}

func TestLogin(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	hosts, roster, _ := local.GenTree(3, true)
	services := local.GetServices(hosts, service.ServiceID)

	link := &api.Link{services[0].(*service.Service).Pin, roster, nil, []chains.User{0}}
	linkReply := &api.LinkReply{}
	require.Nil(t, new().send(roster, link, linkReply))

	login := &api.Login{linkReply.Master, 0, []byte{}}
	loginReply := &api.LoginReply{}
	require.Nil(t, new().send(roster, login, loginReply))

	assert.Equal(t, 32, len(loginReply.Token))
	assert.Equal(t, 0, len(loginReply.Elections))
	assert.True(t, loginReply.Admin)
}

func TestOpen(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	hosts, roster, _ := local.GenTree(3, true)
	services := local.GetServices(hosts, service.ServiceID)

	link := &api.Link{services[0].(*service.Service).Pin, roster, nil, []chains.User{0}}
	linkReply := &api.LinkReply{}
	require.Nil(t, new().send(roster, link, linkReply))

	login := &api.Login{linkReply.Master, 0, []byte{}}
	loginReply := &api.LoginReply{}
	require.Nil(t, new().send(roster, login, loginReply))

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	open := &api.Open{loginReply.Token, linkReply.Master, election}
	openReply := &api.OpenReply{}
	require.Nil(t, new().send(roster, open, openReply))

	assert.NotEqual(t, "", openReply.Genesis)
	assert.NotNil(t, openReply.Key)
}

func TestCast(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	hosts, roster, _ := local.GenTree(3, true)
	services := local.GetServices(hosts, service.ServiceID)

	link := &api.Link{services[0].(*service.Service).Pin, roster, nil, []chains.User{0}}
	linkReply := &api.LinkReply{}
	require.Nil(t, new().send(roster, link, linkReply))

	login := &api.Login{linkReply.Master, 0, []byte{}}
	loginReply := &api.LoginReply{}
	require.Nil(t, new().send(roster, login, loginReply))

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	open := &api.Open{loginReply.Token, linkReply.Master, election}
	openReply := &api.OpenReply{}
	require.Nil(t, new().send(roster, open, openReply))

	ballot := &chains.Ballot{0, nil, nil, nil}
	cast := &api.Cast{loginReply.Token, openReply.Genesis, ballot}
	castReply := &api.CastReply{}
	require.Nil(t, new().send(roster, cast, castReply))

	assert.Equal(t, 2, int(castReply.Index))
}

func TestAggregate(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	hosts, roster, _ := local.GenTree(3, true)
	services := local.GetServices(hosts, service.ServiceID)

	link := &api.Link{services[0].(*service.Service).Pin, roster, nil, []chains.User{0}}
	linkReply := &api.LinkReply{}
	require.Nil(t, new().send(roster, link, linkReply))

	login := &api.Login{linkReply.Master, 0, []byte{}}
	loginReply := &api.LoginReply{}
	require.Nil(t, new().send(roster, login, loginReply))

	election := &chains.Election{Name: "", Creator: 0, Users: []chains.User{0}}
	open := &api.Open{loginReply.Token, linkReply.Master, election}
	openReply := &api.OpenReply{}
	require.Nil(t, new().send(roster, open, openReply))

	ballot := &chains.Ballot{0, nil, nil, nil}
	cast := &api.Cast{loginReply.Token, openReply.Genesis, ballot}
	castReply := &api.CastReply{}
	require.Nil(t, new().send(roster, cast, castReply))

	aggregate := &api.Aggregate{loginReply.Token, openReply.Genesis, chains.BALLOTS}
	aggregateReply := &api.AggregateReply{}
	require.Nil(t, new().send(roster, aggregate, aggregateReply))

	assert.Equal(t, 1, len(aggregateReply.Box.Ballots))
	assert.Equal(t, chains.User(0), aggregateReply.Box.Ballots[0].User)
}

func TestFinalize(t *testing.T) {
	n := 10
	local := onet.NewTCPTest()
	defer local.CloseAll()

	hosts, roster, _ := local.GenTree(3, true)
	services := local.GetServices(hosts, service.ServiceID)

	link := &api.Link{services[0].(*service.Service).Pin, roster, nil, []chains.User{0}}
	linkReply := &api.LinkReply{}
	require.Nil(t, new().send(roster, link, linkReply))

	logins := make([]*api.LoginReply, n)
	users := make([]chains.User, n)
	for i := 0; i < n; i++ {
		login := &api.Login{linkReply.Master, chains.User(i), []byte{}}
		loginReply := &api.LoginReply{}
		require.Nil(t, new().send(roster, login, loginReply))
		logins[i] = loginReply
		users[i] = chains.User(i)
	}

	election := &chains.Election{Name: "", Creator: 0, Users: users}
	open := &api.Open{logins[0].Token, linkReply.Master, election}
	openReply := &api.OpenReply{}
	require.Nil(t, new().send(roster, open, openReply))

	for i := 0; i < n; i++ {
		M, _ := Suite.Point().Pick([]byte{byte(i)}, Stream)
		k := Suite.Scalar().Pick(Stream)
		K := Suite.Point().Mul(nil, k)
		S := Suite.Point().Mul(openReply.Key, k)
		C := S.Add(S, M)
		ballot := &chains.Ballot{chains.User(i), K, C, nil}
		cast := &api.Cast{logins[i].Token, openReply.Genesis, ballot}
		require.Nil(t, new().send(roster, cast, &api.CastReply{}))
	}

	finalize := &api.Finalize{logins[0].Token, openReply.Genesis}
	fr := &api.FinalizeReply{}
	require.Nil(t, new().send(roster, finalize, fr))

	assert.Equal(t, n, len(fr.Shuffle.Ballots))
	assert.Equal(t, n, len(fr.Decryption.Ballots))
	for i := 0; i < n; i++ {
		ballot := fr.Decryption.Ballots[0]
		assert.Equal(t, int(ballot.User), int(ballot.Text[0]))
	}
}
