package chains

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchElection(t *testing.T) {
	_, err := FetchElection(roster, "0")
	assert.NotNil(t, err)

	_, err = FetchElection(roster, "")
	assert.NotNil(t, err)

	e, _ := FetchElection(roster, election.ID)
	assert.Equal(t, e.ID, election.ID)
	assert.Equal(t, 2, int(e.Stage))
}

func TestBallots(t *testing.T) {
	box, _ := election.Ballots()
	assert.Equal(t, 1, len(box.Ballots))
}

func TestShuffle(t *testing.T) {
	e, _ := FetchElection(roster, election.ID)
	e.Stage = 0
	_, err := e.Shuffle()
	assert.NotNil(t, err)

	e, _ = FetchElection(roster, election.ID)
	box, _ := e.Shuffle()
	assert.Equal(t, 0, len(box.Ballots))
}

func TestDecryption(t *testing.T) {
	e, _ := FetchElection(roster, election.ID)
	e.Stage = 1
	_, err := e.Decryption()
	assert.NotNil(t, err)

	e, _ = FetchElection(roster, election.ID)
	box, _ := e.Decryption()
	assert.Equal(t, 0, len(box.Ballots))
}

func TestAppendElection(t *testing.T) {
	genesis, _ := client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
	id := base64.StdEncoding.EncodeToString(genesis.Hash)
	election := &Election{Name: "", Roster: roster, ID: id}

	index, _ := election.Append(&Ballot{})
	assert.Equal(t, 1, index)
}

func TestIsUser(t *testing.T) {
	e := &Election{"", 100000, []User{200000, 300000}, "", nil, nil, nil, 0, "", ""}
	assert.True(t, e.IsUser(200000))
	assert.False(t, e.IsUser(100000))
	assert.False(t, e.IsUser(400000))
}

func TestIsCreator(t *testing.T) {
	e := &Election{"", 100000, []User{200000, 300000}, "", nil, nil, nil, 0, "", ""}
	assert.True(t, e.IsCreator(100000))
	assert.False(t, e.IsCreator(200000))
	assert.False(t, e.IsCreator(400000))
}
