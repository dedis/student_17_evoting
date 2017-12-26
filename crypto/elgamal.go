package crypto

import (
	"errors"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/proof"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/crypto.v0/shuffle"
)

// Encrypt performs the standard ElGamal encryption algorithm.
// (alpha, beta) = (g^k, public^k * m).
func Encrypt(public abstract.Point, message []byte) (abstract.Point, abstract.Point) {
	M, _ := Suite.Point().Pick(message, Stream)

	k := Suite.Scalar().Pick(Stream)
	alpha := Suite.Point().Mul(nil, k)
	S := Suite.Point().Mul(public, k)
	beta := S.Add(S, M)
	return alpha, beta
}

// Decrypt performs the standard ElGamal decryption algorithm.
// m = beta / (alpha^secret).
func Decrypt(secret abstract.Scalar, alpha, beta abstract.Point) abstract.Point {
	S := Suite.Point().Mul(alpha, secret)
	return Suite.Point().Sub(beta, S)
}

// Shuffle permutes and reencrypts ElGamal ciphertext pairs and returns it with
// prover function for verifiability.
func Shuffle(public abstract.Point, alpha, beta []abstract.Point) (
	[]abstract.Point, []abstract.Point, []int, proof.Prover, error) {

	k := len(alpha)

	if k != len(beta) {
		return nil, nil, nil, nil, errors.New("ElGamal pair lists not of equal length")
	}
	if k < 2 {
		return nil, nil, nil, nil, errors.New("Not enough elements (> 1) to permute")
	}

	ps := shuffle.PairShuffle{}
	ps.Init(Suite, k)

	pi := make([]int, k)
	for i := 0; i < k; i++ {
		pi[i] = i
	}

	for i := k - 1; i > 0; i-- {
		j := int(random.Uint64(Stream) % uint64(i+1))
		if j != i {
			t := pi[j]
			pi[j] = pi[i]
			pi[i] = t
		}
	}

	tau := make([]abstract.Scalar, k)
	for i := 0; i < k; i++ {
		tau[i] = Suite.Scalar().Pick(Stream)
	}

	gamma, delta := make([]abstract.Point, k), make([]abstract.Point, k)
	for i := 0; i < k; i++ {
		gamma[i] = Suite.Point().Mul(nil, tau[pi[i]])
		gamma[i].Add(gamma[i], alpha[pi[i]])
		delta[i] = Suite.Point().Mul(public, tau[pi[i]])
		delta[i].Add(delta[i], beta[pi[i]])
	}

	prover := func(ctx proof.ProverContext) error {
		return ps.Prove(pi, nil, public, tau, alpha, beta, Stream, ctx)
	}
	return gamma, delta, pi, prover, nil
}
