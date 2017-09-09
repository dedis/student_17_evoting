package shuffle

import (
	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/onet.v1"
)

// Name defines the protocol identifier in the onet service.
const Name = "shuffle"

// Prompt is the message sent from one node to another to invoke a new
// shuffle procedure at the receiver.
type Prompt struct {
	Latest *skipchain.SkipBlock
}

// MessagePrompt wraps the Prompt message. For compatibilty reasons demanded
// by the onet framework.
type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

// Terminate is sent by the leaf node to the root node to signal that the last
// shuffle has been appended to the SkipChain.
type Terminate struct {
	Latest *skipchain.SkipBlock
}

// MessageTerminate wraps the Terminate message. For compatibility reasons
// demanded by the onet framework.
type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
