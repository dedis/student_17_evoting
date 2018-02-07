package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

func TestFetchMaster(t *testing.T) {
	local, master := setupMaster()
	defer local.CloseAll()

	m, _ := FetchMaster(master.Roster, master.ID)
	assert.Equal(t, master.ID, m.ID)
}

func TestLinks(t *testing.T) {
	local, master := setupMaster()
	defer local.CloseAll()

	links, _ := master.Links()
	assert.Equal(t, 1, len(links))
}

func TestIsAdmin(t *testing.T) {
	m := &Master{Admins: []uint32{0}}
	assert.True(t, m.IsAdmin(0))
	assert.False(t, m.IsAdmin(1))
}

func setupMaster() (*onet.LocalTest, *Master) {
	local := onet.NewLocalTest()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	chain, _ := New(roster, nil)
	master := &Master{ID: chain.Hash, Roster: roster, Admins: []uint32{0, 1}}
	Store(master.Roster, master.ID, master, &Link{})

	return local, master
}
