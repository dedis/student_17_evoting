package shuffle

import (
	"errors"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

// Protocol is the core structure of the protocol. All handlers are are
// attached to it. Furthermore, it can be used to pass over values from the service
// upon initialization and it contains the protocol result after the protocol has
// finished.
type Protocol struct {
	*onet.TreeNodeInstance

	// Key is the public key of the election.
	Key abstract.Point

	// Box contains the encrypted ballots to be shuffled (re-encrypted).
	Box *chains.Box
	// Shuffle contains the shuffled (re-encrypted) ballots upon protocol termination.
	Shuffle *chains.Box

	// Finished is a channel through which termination can be signaled to the service.
	Finished chan bool
}

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	onet.GlobalProtocolRegister(Name, New)
}

// New initializes the protocol object and registers all the handlers at the
// onet service.
func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}
	protocol.RegisterHandler(protocol.HandlePrompt)
	protocol.RegisterHandler(protocol.HandleTerminate)
	return protocol, nil
}

// Start is called on the root node prompting it to send itself a Prompt message.
func (p *Protocol) Start() error {
	if len(p.Box.Ballots) < 2 {
		return errors.New("Not enough (> 1) ballots to shuffle")
	}

	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{p.Key, p.Box.Ballots}})
}

// HandlePrompt is executed by each node one after another. It shuffles the ballots
// contained in the messages and passes them on to next node. When the leaf node is
// reached it sends the final shuffle back to the root node in a Terminate message.
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

// HandleTerminate is executed by the root node upon protocol termination. It sets the
// final shuffle in the protocol object and signals the termination to the service.
func (p *Protocol) HandleTerminate(terminate MessageTerminate) error {
	p.Shuffle = &chains.Box{terminate.Shuffle}
	p.Finished <- true
	return nil
}
