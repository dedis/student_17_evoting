package shuffle

import (
	"gopkg.in/dedis/onet.v1"
)

// Prompt is sent from node to node prompting the receiver to perform
// their respective shuffle (re-encryption) of the ballots.
type Prompt struct{}

// MessagePrompt is a wrapper around Prompt.
type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

// Terminate is sent by the leaf node to the root node upon completion of
// the last shuffle, which terminates the protocol.
type Terminate struct{}

// MessageTerminate is a wrapper around Terminate.
type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
