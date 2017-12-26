package decrypt

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
)

// Name is the protocol identifier string.
const Name = "decrypt"

// Prompt is sent to every node to perform the decryption with their
// respective secrets.
type Prompt struct {
	// Shuffle is a list of shuffled (re-encrypted) ballots.
	Shuffle []*chains.Ballot
}

// MessagePrompt is a wrapper around Prompt. It is mandated by the
// onet protocol service.
type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

// Terminate is sent to the root node by every leaf node which finishes
// the decryption protocol.
type Terminate struct {
	// Index is the order of the node within the DKG protocol.
	Index int
	// Decrypted is a list of decrypted ballots (raw points) from a node.
	Decrypted []abstract.Point
}

// MessageTerminate is a wrapper around Terminate. It is mandated by the
// onet protocol service.
type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
