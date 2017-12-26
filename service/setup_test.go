package service

import (
	"crypto/cipher"
	"testing"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/onet.v1"
)

var nodes []*onet.Server
var roster *onet.Roster
var service *Service

var suite abstract.Suite
var stream cipher.Stream

func TestMain(m *testing.M) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	nodes, roster, _ = local.GenTree(3, true)
	service = local.GetServices(nodes, ServiceID)[0].(*Service)

	suite = ed25519.NewAES128SHA256Ed25519(false)
	stream = suite.Cipher(abstract.RandomKey)
	m.Run()
}
