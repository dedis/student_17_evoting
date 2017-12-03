package chains

import (
	"testing"

	"github.com/dedis/cothority/skipchain"
	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

func TestChain(t *testing.T) {
	local, roster, m, _, _ := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := chain(roster, nil)
	assert.NotNil(t, err)

	c, _ := chain(roster, m)
	assert.Equal(t, 3, len(c))
}

func TestCreate(t *testing.T) {
	local, roster, _, _, _ := setup()
	defer local.CloseAll()

	// Invalid marshal
	_, err := Create(roster, 0)
	assert.NotNil(t, err)

	g, _ := Create(roster, nil)
	assert.NotNil(t, g)
}

func TestStore(t *testing.T) {
	local, roster, m, _, _ := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := Store(roster, nil, nil)
	assert.NotNil(t, err)

	// Invalid marshal
	_, err = Store(roster, m, 0)
	assert.NotNil(t, err)

	index, _ := Store(roster, m, &Link{nil})
	assert.Equal(t, 3, index)
}

func TestGetElection(t *testing.T) {
	local, roster, _, e, _ := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := GetElection(roster, nil)
	assert.NotNil(t, err)

	g, _ := GetElection(roster, e)
	assert.Equal(t, "election1", g.Name)
}

func TestGetMaster(t *testing.T) {
	local, roster, m, _, _ := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := GetMaster(roster, nil)
	assert.NotNil(t, err)

	g, _ := GetMaster(roster, m)
	assert.Equal(t, User(0), g.Admins[0])
}

func TestGetLinks(t *testing.T) {
	local, roster, m, _, _ := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := GetLinks(roster, nil)
	assert.NotNil(t, err)

	l, _ := GetLinks(roster, m)
	assert.Equal(t, 2, len(l))
}

func TestGetBallots(t *testing.T) {
	local, roster, _, e, _ := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := GetBallots(roster, nil)
	assert.NotNil(t, err)

	b, _ := GetBallots(roster, e)
	assert.Equal(t, 1, len(b))
}

func TestGetBox(t *testing.T) {
	local, roster, _, e1, e2 := setup()
	defer local.CloseAll()

	// Invalid id
	_, err := GetBox(roster, nil, BALLOTS)
	assert.NotNil(t, err)

	// Unfinalized, invalid
	_, err = GetBox(roster, e1, SHUFFLE)
	assert.NotNil(t, err)

	// Unfinalized, valid
	b, _ := GetBox(roster, e1, BALLOTS)
	assert.Equal(t, 1, len(b.Ballots))

	// Finalized
	b1, _ := GetBox(roster, e2, BALLOTS)
	b2, _ := GetBox(roster, e2, SHUFFLE)
	b3, _ := GetBox(roster, e2, DECRYPTION)
	assert.Equal(t, 0, len(b1.Ballots), len(b2.Ballots), len(b3.Ballots))
}

// Create a master skipchain with two links and two election skipchains.
// The first one unfinalized with one ballot second finalized with one ballot.
func setup() (
	*onet.LocalTest,
	*onet.Roster,
	skipchain.SkipBlockID,
	skipchain.SkipBlockID,
	skipchain.SkipBlockID) {

	local := onet.NewLocalTest()
	_, roster, _ := local.GenTree(3, true)

	client := skipchain.NewClient()
	master := &Master{Roster: roster, Admins: []User{0}}
	mGen, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, master, nil)

	election1 := &Election{Name: "election1"}
	election2 := &Election{Name: "election2"}
	ballot := &Ballot{User: User(1)}
	box := &Box{nil}

	eGen1, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, nil, nil)
	rep, _ := client.StoreSkipBlock(eGen1, roster, election1)
	client.StoreSkipBlock(rep.Latest, roster, ballot)

	eGen2, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, nil, nil)
	rep, _ = client.StoreSkipBlock(eGen2, roster, election2)
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, ballot)
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, box)
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, box)
	client.StoreSkipBlock(rep.Latest, roster, box)

	rep, _ = client.StoreSkipBlock(mGen, roster, &Link{eGen1.Hash})
	client.StoreSkipBlock(rep.Latest, roster, &Link{eGen2.Hash})

	return local, roster, mGen.Hash, eGen1.Hash, eGen2.Hash
}
