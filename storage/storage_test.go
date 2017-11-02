package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElectionsForUser(t *testing.T) {
	chain, local := newChain()
	defer local.CloseAll()

	storage := Storage{Chains: map[string]*Chain{"el1": chain}}

	elections := storage.ElectionsForUser("user3")
	assert.Equal(t, 0, len(elections))
	elections = storage.ElectionsForUser("admin")
	assert.Equal(t, 1, len(elections))
	elections = storage.ElectionsForUser("user1")
	assert.Equal(t, 1, len(elections))
}
