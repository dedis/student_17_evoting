package shuffle

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestSimulate(t *testing.T) {
	key := crypto.Random()
	ballot := &chains.Ballot{Alpha: crypto.Random(), Beta: crypto.Random()}

	mixes := Simulate(3, key, []*chains.Ballot{ballot, ballot})
	assert.Equal(t, 3, len(mixes))
}
