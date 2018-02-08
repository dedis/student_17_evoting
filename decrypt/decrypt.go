package decrypt

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"

	"github.com/dedis/onet/network"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
)

// Name is the protocol identifier string.
const Name = "decrypt"

// Protocol is the core structure of the protocol.
type Protocol struct {
	*onet.TreeNodeInstance

	Secret   *dkg.SharedSecret // Secret is the private key share from the DKG.
	Election *chains.Election  // Election to be decrypted.
	Finished chan bool         // Flag to signal protocol termination.
}

func init() {
	network.RegisterMessages(Prompt{}, Terminate{})
	onet.GlobalProtocolRegister(Name, New)
}

// New initializes the protocol object and registers all the handlers.
func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}
	protocol.RegisterHandlers(protocol.HandlePrompt, protocol.HandleTerminate)
	return protocol, nil
}

// Start is called on the root node prompting it to send itself a Prompt message.
func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{}})
}

// HandlePrompt retrieves the mixes, verifies them and performs a partial decryption
// on the last mix before appending it to the election skipchain.
func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	box, err := p.Election.Box()
	if err != nil {
		return err
	}
	mixes, err := p.Election.Mixes()
	if err != nil {
		return err
	}

	var partial *chains.Partial
	if !Verify(p.Election.Key, box, mixes) {
		partial = &chains.Partial{Flag: true, Node: p.Name()}
	} else {
		last := mixes[len(mixes)-1].Ballots
		points := make([]abstract.Point, len(box.Ballots))
		for i := range points {
			points[i] = crypto.Decrypt(p.Secret.V, last[i].Alpha, last[i].Beta)
		}
		partial = &chains.Partial{Points: points, Flag: false, Node: p.Name()}
	}

	if err = p.Election.Store(partial); err != nil {
		return err
	}

	if p.IsLeaf() {
		return p.SendTo(p.Root(), &Terminate{})
	}
	return p.SendToChildren(&Prompt{})
}

// HandleTerminate concludes to the protocol.
func (p *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	p.Finished <- true
	return nil
}

// Verify iteratively checks the integrity of each mix.
func Verify(key abstract.Point, box *chains.Box, mixes []*chains.Mix) bool {
	x, y := chains.Split(box.Ballots)
	v, w := chains.Split(mixes[0].Ballots)
	if crypto.Verify(mixes[0].Proof, key, x, y, v, w) != nil {
		return false
	}

	for i := 0; i < len(mixes)-1; i++ {
		x, y = chains.Split(mixes[i].Ballots)
		v, w = chains.Split(mixes[i+1].Ballots)
		if crypto.Verify(mixes[i+1].Proof, key, x, y, v, w) != nil {
			return false
		}
	}
	return true
}
