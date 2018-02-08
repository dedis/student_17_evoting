package shuffle

import (
	"errors"

	"gopkg.in/dedis/crypto.v0/proof"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

// Name is the protocol identifier string.
const Name = "shuffle"

// Protocol is the core structure of the protocol.
type Protocol struct {
	*onet.TreeNodeInstance

	Election *chains.Election // Election to be shuffled.
	Finished chan bool        // Flag to signal protocol termination.
}

func init() {
	network.RegisterMessages(Prompt{}, Terminate{})
	onet.GlobalProtocolRegister(Name, New)
}

// New initializes the protocol object and registers all the handlers.
func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool, 1)}
	protocol.RegisterHandlers(protocol.HandlePrompt, protocol.HandleTerminate)
	return protocol, nil
}

// Start is called on the root node prompting it to send itself a Prompt message.
func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{}})
}

// HandlePrompt retrieves, shuffles and stores the mix back on the skipchain.
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

	if len(ballots) < 2 {
		return errors.New("Not enough (> 2) ballots to shuffle")

	}

	alpha, beta := chains.Split(ballots)
	gamma, delta, _, prover := crypto.Shuffle(p.Election.Key, alpha, beta)

	proof, err := proof.HashProve(crypto.Suite, "", crypto.Stream, prover)
	if err != nil {
		return err
	}

	mix := &chains.Mix{Ballots: chains.Combine(gamma, delta), Proof: proof, Node: p.Name()}
	if err := p.Election.Store(mix); err != nil {
		return err
	}

	if p.IsLeaf() {
		return p.SendTo(p.Root(), &Terminate{})
	}
	return p.SendToChildren(&Prompt{})
}

// HandleTerminate concludes the protocol.
func (p *Protocol) HandleTerminate(terminate MessageTerminate) error {
	p.Finished <- true
	return nil
}
