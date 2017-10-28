package api

import (
	"crypto/cipher"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
)

// ID is used for registration on the onet.
const ID = "nevv"

var Suite abstract.Suite
var Stream cipher.Stream

func init() {
	Suite = ed25519.NewAES128SHA256Ed25519(false)
	Stream = Suite.Cipher(abstract.RandomKey)
}

// Client structure for communication with the CoSi service.
// type Client struct {
// 	*onet.Client
// }

// // NewClient instantiates a new cosi.Client.
// func NewClient() *Client {
// 	return &Client{Client: onet.NewClient(ID)}
// }
