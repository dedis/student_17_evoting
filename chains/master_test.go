package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchMaster(t *testing.T) {
	m, _ := FetchMaster(master.Roster, master.ID)
	assert.Equal(t, master.ID, m.ID)
}

func TestLinks(t *testing.T) {
	links, _ := master.Links()
	assert.Equal(t, 1, len(links))
}

func TestIsAdmin(t *testing.T) {
	master := &Master{Admins: []uint32{0}}
	assert.True(t, master.IsAdmin(0))
	assert.False(t, master.IsAdmin(1))
}
