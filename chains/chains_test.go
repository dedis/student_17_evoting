package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	block, _ := New(election.Roster, nil)
	assert.NotNil(t, block)
}

func TestStore(t *testing.T) {
	chain, _ := New(election.Roster, nil)
	assert.Nil(t, Store(election.Roster, chain.Hash, &Link{}))
}
