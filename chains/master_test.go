package chains

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchMaster(t *testing.T) {
	_, err := FetchMaster(roster, "0")
	assert.NotNil(t, err)
	_, err = FetchMaster(roster, "")
	assert.NotNil(t, err)

	block, _ := FetchMaster(roster, base64.StdEncoding.EncodeToString(masterGenesis.Hash))
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
