package crypto

import "gopkg.in/dedis/crypto.v0/abstract"

// Encrypt performs the standard ElGamal encryption algorithm.
// (K, C) = (g^k, public^k * m).
func Encrypt(public abstract.Point, message []byte) (abstract.Point, abstract.Point) {
	M, _ := Suite.Point().Pick(message, Stream)

	k := Suite.Scalar().Pick(Stream)
	K := Suite.Point().Mul(nil, k)
	S := Suite.Point().Mul(public, k)
	C := S.Add(S, M)
	return K, C
}

// Decrypt performs the standard ElGamal decryption algorithm.
// m = C / (K^secret).
func Decrypt(secret abstract.Scalar, K, C abstract.Point) ([]byte, error) {
	S := Suite.Point().Mul(K, secret)
	M := Suite.Point().Sub(C, S)
	return M.Data()
}
