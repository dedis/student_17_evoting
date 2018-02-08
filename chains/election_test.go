package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

func TestBox(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)
	election, _ := GenElectionChain(roster, 0, []uint32{0}, 10, RUNNING)

	box, _ := election.Box()
	assert.Equal(t, 10, len(box.Ballots))
}

func TestMixes(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)
	election, _ := GenElectionChain(roster, 0, []uint32{0}, 10, SHUFFLED)

	mixes, _ := election.Mixes()
	assert.Equal(t, 3, len(mixes))
}

func TestPartials(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)
	election, _ := GenElectionChain(roster, 0, []uint32{0}, 10, DECRYPTED)

	partials, _ := election.Partials()
	assert.Equal(t, 3, len(partials))
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
