package service

import (
	"crypto/cipher"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
)

var suite abstract.Suite
var stream cipher.Stream

func init() {
	suite = ed25519.NewAES128SHA256Ed25519(false)
	stream = suite.Cipher(abstract.RandomKey)
}

func TestAssertLevel(t *testing.T) {
	log := map[string]*stamp{"0": &stamp{0, true, 0}, "1": &stamp{1, false, 0}}
	service := &Service{state: &state{log}}

	// Not logged in
	_, err := service.assertLevel("2", false)
	assert.NotNil(t, err)

	// Not priviledged
	_, err = service.assertLevel("1", true)
	assert.NotNil(t, err)

	// Valid assertion
	user, _ := service.assertLevel("0", true)
	assert.Equal(t, 0, int(user))
	user, _ = service.assertLevel("1", false)
	assert.Equal(t, 1, int(user))
}
