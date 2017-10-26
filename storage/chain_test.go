package storage

import (
	"testing"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

func TestElection(t *testing.T) {
	chain, local := newChain()
	defer local.CloseAll()

	assert.Equal(t, "test", chain.Election().ID)
}

func TestIsShuffled(t *testing.T) {
	chain, local := newChain()
	defer local.CloseAll()

	assert.False(t, chain.IsShuffled())

	ballot := &api.BallotNew{"", nil, nil, nil}
	_, _ = chain.Store(ballot)
	_, _ = chain.Store(ballot)
	assert.False(t, chain.IsShuffled())

	shuffle := &api.BoxNew{nil}
	_, _ = chain.Store(shuffle)
	assert.True(t, chain.IsShuffled())
}

func TestIsDecrypted(t *testing.T) {
	chain, local := newChain()
	defer local.CloseAll()

	assert.False(t, chain.IsShuffled())

	ballot := &api.BallotNew{}
	_, _ = chain.Store(ballot)
	_, _ = chain.Store(ballot)
	assert.False(t, chain.IsDecrypted())

	shuffle := &api.BoxNew{nil}
	_, _ = chain.Store(shuffle)
	assert.False(t, chain.IsDecrypted())

	decrypt := &api.BoxNew{nil}
	_, _ = chain.Store(decrypt)
	assert.True(t, chain.IsDecrypted())
}

func TestStore(t *testing.T) {
	chain, local := newChain()
	local.CloseAll()

	block, err := chain.Store(&api.BallotNew{})
	assert.Equal(t, -1, block)
	assert.NotNil(t, err)

	chain, local = newChain()
	defer local.CloseAll()

	block, _ = chain.Store(&api.BallotNew{})
	assert.Equal(t, 2, block)
	block, _ = chain.Store(&api.BallotNew{})
	assert.Equal(t, 3, block)
}

func TestBallots(t *testing.T) {
	chain, local := newChain()
	local.CloseAll()

	ballots, err := chain.Ballots()
	assert.Nil(t, ballots)
	assert.NotNil(t, err)

	chain, local = newChain()
	defer local.CloseAll()

	ballot1 := &api.BallotNew{"b1", nil, nil, []byte("b1")}
	_, _ = chain.Store(ballot1)
	ballot2 := &api.BallotNew{"b2", nil, nil, []byte("b2")}
	_, _ = chain.Store(ballot2)

	ballots, _ = chain.Ballots()
	assert.Equal(t, 2, len(ballots))
	assert.Equal(t, string(ballots[0].Clear), ballots[0].User)
	assert.Equal(t, string(ballots[1].Clear), ballots[1].User)
}

func TestBoxes(t *testing.T) {
	chain, local := newChain()
	local.CloseAll()

	boxes, err := chain.Boxes()
	assert.Nil(t, boxes)
	assert.NotNil(t, err)

	chain, local = newChain()
	defer local.CloseAll()

	_, _ = chain.Store(&api.BallotNew{})

	boxes, _ = chain.Boxes()
	assert.Equal(t, 0, len(boxes))

	box := &api.BoxNew{make([]*api.BallotNew, 0)}
	_, _ = chain.Store(box)

	boxes, _ = chain.Boxes()
	assert.Equal(t, 1, len(boxes))
	assert.Equal(t, 0, len(boxes[0].Ballots))

	box = &api.BoxNew{[]*api.BallotNew{&api.BallotNew{}}}
	_, _ = chain.Store(box)

	boxes, _ = chain.Boxes()
	assert.Equal(t, 2, len(boxes))
	assert.Equal(t, 1, len(boxes[1].Ballots))
}

func newChain() (*Chain, *onet.LocalTest) {
	local := onet.NewTCPTest()

	_, roster, _ := local.GenTree(3, true)

	client := skipchain.NewClient()
	genesis, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, nil, nil)
	election := api.Election{
		ID:          "test",
		Admin:       "admin",
		Start:       "",
		End:         "",
		Data:        []byte{},
		Roster:      roster,
		Users:       []string{"user1", "user2"},
		Key:         nil,
		Description: "",
	}

	_, _ = client.StoreSkipBlock(genesis, nil, &election)

	return &Chain{Genesis: genesis}, local
}
