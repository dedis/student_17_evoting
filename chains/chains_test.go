package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

func TestNew(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	_, roster, _ := local.GenTree(3, true)

	block, err := New(roster, nil)
	assert.NotNil(t, block)
	assert.Nil(t, err)
}

func TestChain(t *testing.T) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	_, roster, _ := local.GenTree(3, true)

	genesis, _ := New(roster, nil)
	_, err := chain(roster, nil)
	assert.NotNil(t, err)

	chain, _ := chain(roster, genesis.Hash)
	assert.NotNil(t, chain)
}
