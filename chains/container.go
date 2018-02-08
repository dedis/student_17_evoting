package chains

import (
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/proof"
	rabin "gopkg.in/dedis/crypto.v0/share/dkg"
)

// Ballot represents an encrypted vote.
type Ballot struct {
	User uint32 // User identifier.

	// ElGamal ciphertext pair.
	Alpha abstract.Point
	Beta  abstract.Point
}

// Box is a wrapper around a list of encrypted ballots.
type Box struct {
	Ballots []*Ballot
}

// genMix generates n mixes with corresponding proofs out of the ballots.
func (b *Box) genMix(key abstract.Point, n int) []*Mix {
	mixes := make([]*Mix, n)

	x, y := Split(b.Ballots)
	for i := range mixes {
		v, w, _, prover := crypto.Shuffle(key, x, y)
		proof, _ := proof.HashProve(crypto.Suite, "", crypto.Stream, prover)
		mixes[i] = &Mix{Ballots: Combine(v, w), Proof: proof, Node: string(i)}
		x, y = v, w
	}
	return mixes
}

// Mix contains the shuffled ballots.
type Mix struct {
	Ballots []*Ballot // Ballots are permuted and re-encrypted.
	Proof   []byte    // Proof of the shuffle.

	Node string // Node signifies the creator of the mix.
}

// Partial contains the partially decrypted ballots.
type Partial struct {
	Points []abstract.Point // Points are the partially decrypted plaintexts.

	Flag bool   // Flag signals if the mixes could not be verified.
	Node string // Node signifies the creator of this partial decryption.
}

// genPartials generates partial decryptions for a given list of shared secrets.
func (m *Mix) genPartials(dkgs []*rabin.DistKeyGenerator) []*Partial {
	partials := make([]*Partial, len(dkgs))

	for i, gen := range dkgs {
		secret, _ := dkg.NewSharedSecret(gen)
		points := make([]abstract.Point, len(m.Ballots))
		for j, ballot := range m.Ballots {
			points[j] = crypto.Decrypt(secret.V, ballot.Alpha, ballot.Beta)
		}
		partials[i] = &Partial{Points: points, Node: string(i)}
	}
	return partials
}

// genBox generates a box of encrypted ballots.
func genBox(key abstract.Point, n int) *Box {
	ballots := make([]*Ballot, n)
	for i := range ballots {
		a, b := crypto.Encrypt(key, []byte{byte(i)})
		ballots[i] = &Ballot{User: uint32(i), Alpha: a, Beta: b}
	}
	return &Box{Ballots: ballots}
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
