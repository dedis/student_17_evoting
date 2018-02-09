package decrypt

import (
	"github.com/dedis/onet"
)

// Prompt is sent from node to node prompting the receiver to perform
// their respective partial decryption of the last mix.
type Prompt struct{}

// MessagePrompt is a wrapper around Prompt.
type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

// Terminate is sent by the leaf node to the root node upon completion of
// the last partial decryption, which terminates the protocol.
type Terminate struct{}

// MessageTerminate is a wrapper around Terminate.
type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
