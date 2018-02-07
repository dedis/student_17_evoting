package decrypt

import (
	"gopkg.in/dedis/onet.v1"
)

// Name is the protocol identifier string.
const Name = "decrypt"

type Prompt struct{}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct{}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
