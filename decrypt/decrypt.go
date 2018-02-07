package decrypt

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"

	"github.com/dedis/onet/network"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
)

type Protocol struct {
	*onet.TreeNodeInstance

	Secret *dkg.SharedSecret

	Election *chains.Election

	Finished chan bool
}

func init() {
	network.RegisterMessages(Prompt{}, Terminate{})
	onet.GlobalProtocolRegister(Name, New)
}

func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}
	protocol.RegisterHandlers(protocol.HandlePrompt, protocol.HandleTerminate)
	return protocol, nil
}

func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{}})
}

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

	if err = chains.Store(p.Election.Roster, p.Election.ID, partial); err != nil {
		return err
	}

	if p.IsLeaf() {
		return p.SendTo(p.Root(), &Terminate{})
	}
	return p.SendToChildren(&Prompt{})
}

func (p *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	p.Finished <- true
	return nil
}

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
