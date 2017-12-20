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

func TestAppendMaster(t *testing.T) {
	genesis, _ := client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
	master := &Master{nil, genesis.Hash, roster, []User{0}}

	index, _ := master.Append(&Link{nil})
	assert.Equal(t, 1, index)
}

func TestIsAdmin(t *testing.T) {
	master := &Master{nil, nil, nil, []User{123456}}
	assert.True(t, master.IsAdmin(123456))
	assert.False(t, master.IsAdmin(654321))
}
