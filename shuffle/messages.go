package shuffle

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
)

// Name is the protocol identifier string.
const Name = "shuffle"

// Prompt is from node to node propting the receiver the perform
// their respective shuffle (re-encryption) of the ballots.
type Prompt struct {
	// Key is the election's public key.
	Key abstract.Point
	// Ballots is the list of ballots to be shuffled (re-encrypted).
	Ballots []*chains.Ballot
}

// MessagePrompt is a wrapper around Prompt. It is mandated by the
// onet protocol service.
type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

// Terminate is sent by the leaf node to the root node upon completion of
// the last shuffle.
type Terminate struct {
	// Shuffle is the last re-encrypted list of ballots.
	Shuffle []*chains.Ballot
}

// MessageTerminate is a wrapper around Terminate. It is mandated by the
// onet protocol service.
type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
