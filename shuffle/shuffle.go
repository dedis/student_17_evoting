package shuffle

import (
	"errors"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/proof"
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

	Election *chains.Election

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
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool, 1)}
	protocol.RegisterHandler(protocol.HandlePrompt)
	protocol.RegisterHandler(protocol.HandleTerminate)
	return protocol, nil
}

// Start is called on the root node prompting it to send itself a Prompt message.
func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{}})
}

// HandlePrompt is executed by each node one after another. It shuffles the ballots
// contained in the messages and passes them on to next node. When the leaf node is
// reached it sends the final shuffle back to the root node in a Terminate message.
func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	var ballots []*chains.Ballot

	if p.IsRoot() {
		box, err := p.Election.Box()
		if err != nil {
			return err
		}
		ballots = box.Ballots
	} else {
		mixes, err := p.Election.Mixes()
		if err != nil {
			return err
		}
		ballots = mixes[len(mixes)-1].Ballots
	}

	k := len(ballots)
	if k < 2 {
		return errors.New("Not enough (> 2) ballots to shuffle")
	}

	alpha, beta := make([]abstract.Point, k), make([]abstract.Point, k)
	for i, ballot := range ballots {
		alpha[i] = ballot.Alpha
		beta[i] = ballot.Beta
	}

	gamma, delta, _, prover := crypto.Shuffle(p.Election.Key, alpha, beta)
	proof, err := proof.HashProve(crypto.Suite, Name, crypto.Stream, prover)
	if err != nil {
		return err
	}

	mixed := make([]*chains.Ballot, k)
	for i := range mixed {
		mixed[i] = &chains.Ballot{
			Alpha: gamma[i],
			Beta:  delta[i],
		}
	}

	mix := &chains.Mix{Ballots: mixed, Proof: proof, Node: p.Name()}
	if _, err := chains.Store(p.Election.Roster, p.Election.ID, mix); err != nil {
		return err
	}

	if p.IsLeaf() {
		return p.SendTo(p.Root(), &Terminate{})
	}

	return p.SendToChildren(&Prompt{})
}

func (p *Protocol) HandleTerminate(terminate MessageTerminate) error {
	p.Finished <- true
	return nil
}
