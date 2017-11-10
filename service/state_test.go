package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonce(t *testing.T) {
	n1, n2, n3 := nonce(10), nonce(10), nonce(10)
	assert.Equal(t, 10, len(n1), len(n2), len(n3))
	assert.NotEqual(t, n1, n2, n3)
}
