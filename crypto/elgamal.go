package crypto

import (
	"github.com/dedis/kyber"
	"github.com/dedis/kyber/proof"
	"github.com/dedis/kyber/shuffle"
	"github.com/dedis/kyber/util/random"
)

func Encrypt(public kyber.Point, message []byte) (K, C kyber.Point) {
	M := Suite.Point().Embed(message, random.New())

	// ElGamal-encrypt the point to produce ciphertext (K,C).
	k := Suite.Scalar().Pick(random.New()) // ephemeral private key
	K = Suite.Point().Mul(k, nil)          // ephemeral DH public key
	S := Suite.Point().Mul(k, public)      // ephemeral DH shared secret
	C = S.Add(S, M)                        // message blinded with secret
	return
}

func Decrypt(private kyber.Scalar, K, C kyber.Point) kyber.Point {
	// ElGamal-decrypt the ciphertext (K,C) to reproduce the message.
	S := Suite.Point().Mul(private, K) // regenerate shared secret
	return Suite.Point().Sub(C, S)     // use to un-blind the message
}

// // Decrypt performs the standard ElGamal decryption algorithm.
// // m = beta / (alpha^secret).
// func Decrypt(secret abstract.Scalar, alpha, beta abstract.Point) abstract.Point {
// 	S := Suite.Point().Mul(alpha, secret)
// 	return Suite.Point().Sub(beta, S)
// }

// Shuffle permutes and reencrypts ElGamal ciphertext pairs and returns it with
// prover function for verifiability.
// func Shuffle(public kyber.Point, alpha, beta []abstract.Point) (
// 	[]abstract.Point, []abstract.Point, []int, proof.Prover) {

// 	k := len(alpha)

// 	ps := shuffle.PairShuffle{}
// 	ps.Init(Suite, k)

// 	pi := make([]int, k)
// 	for i := 0; i < k; i++ {
// 		pi[i] = i
// 	}

// 	for i := k - 1; i > 0; i-- {
// 		j := int(random.Uint64(Stream) % uint64(i+1))
// 		if j != i {
// 			t := pi[j]
// 			pi[j] = pi[i]
// 			pi[i] = t
// 		}
// 	}

// 	tau := make([]abstract.Scalar, k)
// 	for i := 0; i < k; i++ {
// 		tau[i] = Suite.Scalar().Pick(Stream)
// 	}

// 	gamma, delta := make([]abstract.Point, k), make([]abstract.Point, k)
// 	for i := 0; i < k; i++ {
// 		gamma[i] = Suite.Point().Mul(nil, tau[pi[i]])
// 		gamma[i].Add(gamma[i], alpha[pi[i]])
// 		delta[i] = Suite.Point().Mul(public, tau[pi[i]])
// 		delta[i].Add(delta[i], beta[pi[i]])
// 	}

// 	prover := func(ctx proof.ProverContext) error {
// 		return ps.Prove(pi, nil, public, tau, alpha, beta, Stream, ctx)
// 	}
// 	return gamma, delta, pi, prover
// }

func Verify(tag []byte, public kyber.Point, x, y, v, w []kyber.Point) error {
	verifier := shuffle.Verifier(Suite, nil, public, x, y, v, w)
	return proof.HashVerify(Suite, "", verifier, tag)
}
