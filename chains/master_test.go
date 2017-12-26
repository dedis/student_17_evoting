package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchMaster(t *testing.T) {
	_, err := FetchMaster(roster, []byte{})
	assert.NotNil(t, err)

	block, _ := FetchMaster(roster, master.ID)
	assert.Equal(t, master.ID, block.ID)
}

func TestLinks(t *testing.T) {
	links, _ := master.Links()
	assert.Equal(t, 1, len(links))
}

func TestIsAdmin(t *testing.T) {
	master := &Master{nil, nil, nil, []User{123456}}
	assert.True(t, master.IsAdmin(123456))
	assert.False(t, master.IsAdmin(654321))
}
