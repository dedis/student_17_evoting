package chains

import (
	"gopkg.in/dedis/crypto.v0/abstract"
)

type Ballot struct {
	User uint32

	Alpha abstract.Point
	Beta  abstract.Point
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
	Points []abstract.Point
	Index  uint32

	Flag bool
	Node string
}

// Split separates the ElGamal pairs of a list of ballots into separate lists.
func Split(ballots []*Ballot) (alpha, beta []abstract.Point) {
	n := len(ballots)
	alpha, beta = make([]abstract.Point, n), make([]abstract.Point, n)
	for i, ballot := range ballots {
		alpha[i] = ballot.Alpha
		beta[i] = ballot.Beta
	}
	return
}

// Combine creates a list of ballots from two lists of points.
func Combine(alpha, beta []abstract.Point) []*Ballot {
	ballots := make([]*Ballot, len(alpha))
	for i := range ballots {
		ballots[i] = &Ballot{Alpha: alpha[i], Beta: beta[i]}
	}
	return ballots
}
