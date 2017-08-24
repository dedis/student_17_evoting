package shuffle

import (
	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/onet.v1"
)

// Name ...
const Name = "Shuffle"

type Prompt struct {
	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock
}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct {
	Latest *skipchain.SkipBlock
}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
