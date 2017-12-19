package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
