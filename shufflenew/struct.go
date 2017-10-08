package shufflenew

import (
	"github.com/dedis/cothority/skipchain"
	"gopkg.in/dedis/onet.v1"
)

const Name = "shuffle"

type Prompt struct {
	Genesis *skipchain.SkipBlock
}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct {
}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
