package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElGamal(t *testing.T) {
	secret := Suite.Scalar().Pick(Stream)
	public := Suite.Point().Mul(nil, secret)
	message := []byte("nevv")

	K, C := Encrypt(public, message)
	dec, _ := Decrypt(secret, K, C)
	assert.Equal(t, message, dec)
}
