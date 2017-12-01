package shuffle

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
)

const Name = "shuffle"

type Prompt struct {
	Key     abstract.Point
	Ballots []*chains.Ballot
}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct {
	Shuffle []*chains.Ballot
}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
