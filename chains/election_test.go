package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func TestStore(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	election := &Election{Roster: roster, Stage: RUNNING}
	_ = election.GenChain(10)

	election.Store(&Ballot{User: 1000})

	chain, _ := client.GetUpdateChain(roster, election.ID)
	_, blob, _ := network.Unmarshal(chain.Update[len(chain.Update)-1].Data)
	assert.Equal(t, uint32(1000), blob.(*Ballot).User)
}

func TestBox(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	election := &Election{Roster: roster, Stage: RUNNING}
	_ = election.GenChain(10)

	box, _ := election.Box()
	assert.Equal(t, 10, len(box.Ballots))
}

func TestMixes(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	election := &Election{Roster: roster, Stage: SHUFFLED}
	_ = election.GenChain(10)

	mixes, _ := election.Mixes()
	assert.Equal(t, 3, len(mixes))
}

func TestPartials(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	election := &Election{Roster: roster, Stage: DECRYPTED}
	_ = election.GenChain(10)

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
