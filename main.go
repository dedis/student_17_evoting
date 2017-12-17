package main

import (
	"crypto/cipher"

	_ "github.com/dedis/cothority/cosi/service"
	_ "github.com/qantik/nevv/service"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/onet.v1/app"
)

var Suite abstract.Suite
var Stream cipher.Stream

func init() {
	Suite = ed25519.NewAES128SHA256Ed25519(false)
	Stream = Suite.Cipher(abstract.RandomKey)
}

func main() {
	app.Server()
}
