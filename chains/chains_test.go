package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

func TestNew(t *testing.T) {
	local, roster := setupChain()
	defer local.CloseAll()

	block, _ := New(roster, nil)
	assert.NotNil(t, block)
}

func TestStore(t *testing.T) {
	local, roster := setupChain()
	defer local.CloseAll()

	chain, _ := New(roster, nil)
	assert.Nil(t, Store(roster, chain.Hash, &Master{}, &Link{}))
}

func setupChain() (*onet.LocalTest, *onet.Roster) {
	local := onet.NewLocalTest()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)
	return local, roster
}
