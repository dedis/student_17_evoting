package shuffle

import (
	"errors"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

type Protocol struct {
	*onet.TreeNodeInstance

	Key     abstract.Point
	Box     *chains.Box
	Shuffle *chains.Box

	Finished chan bool
}

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	onet.GlobalProtocolRegister(Name, New)
}

func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}
	protocol.RegisterHandler(protocol.HandlePrompt)
	protocol.RegisterHandler(protocol.HandleTerminate)
	return protocol, nil
}

func (p *Protocol) Start() error {
	if len(p.Box.Ballots) < 2 {
		return errors.New("Not enough (> 1) ballots to shuffle")
	}

	msg := MessagePrompt{p.TreeNode(), Prompt{p.Key, p.Box.Ballots}}
	if err := p.HandlePrompt(msg); err != nil {
		return err
	}

	return nil
}

func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	k := len(prompt.Ballots)
	alpha, beta := make([]abstract.Point, k), make([]abstract.Point, k)

	for i, ballot := range prompt.Ballots {
		alpha[i] = ballot.Alpha
		beta[i] = ballot.Beta
	}

	gamma, delta, pi, _, _ := crypto.Shuffle(prompt.Key, alpha, beta)

	// Reconstruct ballot list with shuffle permutation
	shuffled := make([]*chains.Ballot, k)
	for i := range shuffled {
		shuffled[i] = &chains.Ballot{
			User:  prompt.Ballots[pi[i]].User,
			Alpha: gamma[i],
			Beta:  delta[i],
		}
	}

	if p.IsLeaf() {
		return p.SendTo(p.Root(), &Terminate{shuffled})
	}

	return p.SendToChildren(&Prompt{prompt.Key, shuffled})
}

func (p *Protocol) HandleTerminate(terminate MessageTerminate) error {
	p.Shuffle = &chains.Box{terminate.Shuffle}
	p.Finished <- true
	return nil
}
