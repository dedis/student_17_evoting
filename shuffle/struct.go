package shuffle

import (
	"github.com/qantik/nevv/api"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

const Name = "shuffle"

type Prompt struct {
	Key     abstract.Point
	Ballots []*api.Ballot
}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct {
	Ballots []*api.Ballot
}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
