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
	protocol.RegisterHandler(protocol.HandlePrompt)
	protocol.RegisterHandler(protocol.HandleTerminate)
	return protocol, nil
}

// Start is called on the root node prompting it to send itself a Prompt message.
func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{}})
}

// HandlePrompt retrieves, shuffles and stores the mix back on the skipchain.
func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	latest, err := p.Election.Latest()
	if err != nil {
		return err
	}

	var ballots []*chains.Ballot
	if p.IsRoot() {
		ballots = latest.(*chains.Box).Ballots
	} else {
		ballots = latest.(*chains.Mix).Ballots
	}

	if len(ballots) < 2 {
		return errors.New("Not enough (> 2) ballots to shuffle")
	}

	alpha, beta := split(ballots)
	gamma, delta, _, prover := crypto.Shuffle(p.Election.Key, alpha, beta)

	proof, err := proof.HashProve(crypto.Suite, Name, crypto.Stream, prover)
	if err != nil {
		return err
	}

	mix := &chains.Mix{Ballots: combine(gamma, delta), Proof: proof, Node: p.Name()}
	if err := chains.Store(p.Election.Roster, p.Election.ID, mix); err != nil {
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

// split separates the ElGamal pairs of a list of ballots into separate lists.
func split(ballots []*chains.Ballot) (alpha, beta []abstract.Point) {
	n := len(ballots)
	alpha, beta = make([]abstract.Point, n), make([]abstract.Point, n)
	for i, ballot := range ballots {
		alpha[i] = ballot.Alpha
		beta[i] = ballot.Beta
	}
	return
}

// combine creates a list of ballots from two lists of points.
func combine(alpha, beta []abstract.Point) []*chains.Ballot {
	ballots := make([]*chains.Ballot, len(alpha))
	for i := range ballots {
		ballots[i] = &chains.Ballot{
			Alpha: alpha[i],
			Beta:  beta[i],
		}
	}
	return ballots
}
