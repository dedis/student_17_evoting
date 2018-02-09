package crypto

import (
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/xof/blake"
)

var (
	Suite  = edwards25519.NewBlakeSHA256Ed25519WithRand(blake.New(nil))
	Stream = Suite.RandomStream()
	Base   = Suite.Point().Base()
)

// Random returns an arbitrary Ed25519 curve point.
// func Random() kyber.Point {
// 	point, _ := Suite.Point().Pick(Stream())
// 	return point
// }
