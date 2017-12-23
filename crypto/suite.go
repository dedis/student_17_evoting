package crypto

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
)

var (
	Suite  = ed25519.NewAES128SHA256Ed25519(false)
	Stream = Suite.Cipher(abstract.RandomKey)
)
