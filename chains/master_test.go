package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
)

func TestLinks(t *testing.T) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)
	master := GenMasterChain(roster, []byte{0}, []byte{1})

	links, _ := master.Links()
	assert.Equal(t, 2, len(links))
	assert.Equal(t, skipchain.SkipBlockID([]byte{0}), links[0].ID)
	assert.Equal(t, skipchain.SkipBlockID([]byte{1}), links[1].ID)
}
