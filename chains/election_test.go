package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchElection(t *testing.T) {
	e, _ := FetchElection(election.Roster, election.ID)
	assert.Equal(t, STAGE_SHUFFLED, int(e.Stage))
}

func TestBallots(t *testing.T) {
	box, _ := election.Ballots()
	assert.Equal(t, 2, len(box.Ballots))
}

func TestBox(t *testing.T) {
	box, _ := election.Box()
	assert.Equal(t, 2, len(box.Ballots))
}

func TestLatest(t *testing.T) {
	msg, _ := election.Latest()
	_, ok := msg.(*Mix)
	assert.True(t, ok)
}

func TestMixes(t *testing.T) {
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
