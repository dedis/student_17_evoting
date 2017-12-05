package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUser(t *testing.T) {
	election := &Election{"", 100000, []User{200000, 300000}, "", nil, nil, nil, "", ""}
	assert.True(t, election.IsUser(200000))
	assert.False(t, election.IsUser(100000))
	assert.False(t, election.IsUser(400000))
}

func TestIsCreator(t *testing.T) {
	election := &Election{"", 100000, []User{200000, 300000}, "", nil, nil, nil, "", ""}
	assert.True(t, election.IsCreator(100000))
	assert.False(t, election.IsCreator(200000))
	assert.False(t, election.IsCreator(400000))
}
