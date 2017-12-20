package chains

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// var roster *onet.Roster
// var genesis *skipchain.SkipBlock
// var master *Master

// func TestMain(m *testing.M) {
// 	local := onet.NewTCPTest()
// 	defer local.CloseAll()

// 	_, roster, _ = local.GenTree(3, true)
// 	genesis, _ = client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
// 	master = &Master{nil, genesis.Hash, roster, []User{0}}
// 	rep, _ := client.StoreSkipBlock(genesis, roster, master)
// 	client.StoreSkipBlock(rep.Latest, roster, &Link{nil})

// 	m.Run()
// }

func TestFetchMaster(t *testing.T) {
	_, err := FetchMaster(roster, "0")
	assert.NotNil(t, err)
	_, err = FetchMaster(roster, "0")
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
