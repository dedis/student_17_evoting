package crypto

import (
	"testing"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/proof"
	"gopkg.in/dedis/crypto.v0/shuffle"

	"github.com/stretchr/testify/assert"
)

func TestElGamal(t *testing.T) {
	secret := Suite.Scalar().Pick(Stream)
	public := Suite.Point().Mul(nil, secret)
	message := []byte("nevv")

	K, C := Encrypt(public, message)
	dec, _ := Decrypt(secret, K, C).Data()
	assert.Equal(t, message, dec)
}

func TestShuffle(t *testing.T) {
	secret := Suite.Scalar().Pick(Stream)
	public := Suite.Point().Mul(nil, secret)
	message := []byte("nevv")

	n := 100

	alpha, beta := make([]abstract.Point, n), make([]abstract.Point, n)
	for i := 0; i < n; i++ {
		alpha[i], beta[i] = Encrypt(public, message)
	}

	gamma, delta, _, prover := Shuffle(public, alpha, beta)
	prove, err := proof.HashProve(Suite, "", Stream, prover)
	assert.Nil(t, err)

	verifier := shuffle.Verifier(Suite, nil, public, alpha, beta, gamma, delta)
	err = proof.HashVerify(Suite, "", verifier, prove)
	assert.Nil(t, err)

	for i := 0; i < n; i++ {
		dec, _ := Decrypt(secret, gamma[i], delta[i]).Data()
		assert.Equal(t, message, dec)
	}
}
