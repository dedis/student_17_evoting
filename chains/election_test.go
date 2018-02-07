package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/crypto"
)

func TestFetchElection(t *testing.T) {
	local, election := setupElection()
	defer local.CloseAll()

	e, _ := FetchElection(election.Roster, election.ID)
	assert.Equal(t, STAGE_SHUFFLED, int(e.Stage))
}

func TestBallots(t *testing.T) {
	local, election := setupElection()
	defer local.CloseAll()

	box, _ := election.Ballots()
	assert.Equal(t, 2, len(box.Ballots))
}

func TestBox(t *testing.T) {
	local, election := setupElection()
	defer local.CloseAll()

	box, _ := election.Box()
	assert.Equal(t, 2, len(box.Ballots))
}

func TestLatest(t *testing.T) {
	local, election := setupElection()
	defer local.CloseAll()

	msg, _ := election.Latest()
	_, ok := msg.(*Mix)
	assert.True(t, ok)
}

func TestMixes(t *testing.T) {
	local, election := setupElection()
	defer local.CloseAll()

	mixes, _ := election.Mixes()
	assert.Equal(t, 3, len(mixes))
}

func TestIsUser(t *testing.T) {
	e := &Election{Creator: 0, Users: []uint32{0}}
	assert.True(t, e.IsUser(0))
	assert.False(t, e.IsUser(1))
}

func TestIsCreator(t *testing.T) {
	e := &Election{Creator: 0, Users: []uint32{0, 1}}
	assert.True(t, e.IsCreator(0))
	assert.False(t, e.IsCreator(1))
}

func setupElection() (*onet.LocalTest, *Election) {
	local := onet.NewLocalTest()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	chain, _ := New(roster, nil)
	b1 := &Ballot{User: 0, Alpha: crypto.Random(), Beta: crypto.Random()}
	b2 := &Ballot{User: 1, Alpha: crypto.Random(), Beta: crypto.Random()}
	box := &Box{Ballots: []*Ballot{b1, b2}}
	mix := &Mix{Ballots: []*Ballot{b1, b2}, Proof: []byte{}}

	election := &Election{
		ID:     chain.Hash,
		Roster: roster,
		Key:    crypto.Random(),
		Data:   []byte{},
	}

	Store(election.Roster, election.ID, election, b1, b2, box, mix, mix, mix)
	return local, election
}
