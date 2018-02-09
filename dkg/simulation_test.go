package dkg

import (
	"testing"

	"github.com/dedis/kyber"

	"github.com/stretchr/testify/assert"
)

func TestSimulate(t *testing.T) {
	dkgs, _ := Simulate(5, 4)
	assert.Equal(t, 5, len(dkgs))

	secrets := make([]*SharedSecret, 5)
	for i, dkg := range dkgs {
		secrets[i], _ = NewSharedSecret(dkg)
	}

	var public kyber.Point
	var private kyber.Scalar
	for _, secret := range secrets {
		if public != nil && private != nil {
			assert.Equal(t, public.String(), secret.X.String())
			assert.NotEqual(t, private.String(), secret.V.String())
		}
		public, private = secret.X, secret.V
	}
}
