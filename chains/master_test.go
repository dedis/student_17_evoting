package chains

import (
	"testing"

	"github.com/qantik/nevv/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/onet"
)

func TestLinks(t *testing.T) {
	local := onet.NewLocalTest(crypto.Suite)
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	master := &Master{Roster: roster}
	master.GenChain([]byte{0}, []byte{1})

	links, _ := master.Links()
	assert.Equal(t, 2, len(links))
	assert.Equal(t, skipchain.SkipBlockID([]byte{0}), links[0].ID)
	assert.Equal(t, skipchain.SkipBlockID([]byte{1}), links[1].ID)
}
