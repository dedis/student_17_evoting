package decrypt

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

const Name = "decrypt"

type Prompt struct {
	ElectionName string
}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct {
	Index     int
	Decrypted []abstract.Point
}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
