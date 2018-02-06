package crypto

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
)

var (
	Suite  = ed25519.NewAES128SHA256Ed25519(false)
	Stream = Suite.Cipher(abstract.RandomKey)
	Base   = Suite.Point()
)

// Random returns an arbitrary Ed25519 curve point.
func Random() abstract.Point {
	point, _ := Suite.Point().Pick(nil, Suite.Cipher(abstract.RandomKey))
	return point
}
