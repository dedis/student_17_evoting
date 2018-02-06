package chains

import "gopkg.in/dedis/crypto.v0/abstract"

// Ballot represents a vote and is created by the frontend when a
// user casts his decision.
type Ballot struct {
	// User identifier.
	User uint32

	// Alpha is the first element in the ElGamal ciphertext.
	Alpha abstract.Point
	// Beta is the second element in the ElGamal ciphertext.
	Beta abstract.Point
}

type Box struct {
	Ballots []*Ballot
}

type Mix struct {
	Ballots []*Ballot
	Proof   []byte

	Node string
}

type Partial struct {
	Points []*abstract.Point
	Proof  []byte

	Node string
}
