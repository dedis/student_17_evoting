package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchElection(t *testing.T) {
	_, err := FetchElection(roster, []byte{})
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

func TestIsUser(t *testing.T) {
	e := &Election{"", 100, []User{200, 300}, []byte{}, nil, nil, nil, 0, "", ""}
	assert.True(t, e.IsUser(200))
	assert.False(t, e.IsUser(100))
	assert.False(t, e.IsUser(400))
}

func TestIsCreator(t *testing.T) {
	e := &Election{"", 100, []User{200, 300}, []byte{}, nil, nil, nil, 0, "", ""}
	assert.True(t, e.IsCreator(100))
	assert.False(t, e.IsCreator(200))
	assert.False(t, e.IsCreator(400))
}
