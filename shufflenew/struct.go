package shufflenew

import (
	"github.com/qantik/nevv/api"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

const Name = "shuffle"

type Prompt struct {
	Key     abstract.Point
	Ballots []*api.BallotNew
}

type MessagePrompt struct {
	*onet.TreeNode
	Prompt
}

type Terminate struct {
	Ballots []*api.BallotNew
}

type MessageTerminate struct {
	*onet.TreeNode
	Terminate
}
