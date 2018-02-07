package shuffle

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/proof"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

// Simulate performs and offline version of the shuffle protocol.
func Simulate(n int, key abstract.Point, ballots []*chains.Ballot) []*chains.Mix {
	mixes := make([]*chains.Mix, n)
	for i := range mixes {
		x, y := chains.Split(ballots)
		v, w, _, prover := crypto.Shuffle(key, x, y)

		proof, _ := proof.HashProve(crypto.Suite, "", crypto.Stream, prover)

		ballots = chains.Combine(v, w)
		mixes[i] = &chains.Mix{Ballots: ballots, Proof: proof, Node: string(i)}
	}
	return mixes
}
