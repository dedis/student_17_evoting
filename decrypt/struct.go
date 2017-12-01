package decrypt

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
)

const Name = "decrypt"

type Prompt struct {
	Shuffle []*chains.Ballot
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
