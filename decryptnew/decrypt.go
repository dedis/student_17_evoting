package decryptnew

import (
	"github.com/dedis/onet/network"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/storage"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/share"
	"gopkg.in/dedis/onet.v1"
)

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	_, _ = onet.GlobalProtocolRegister(Name, New)
}

type Protocol struct {
	*onet.TreeNodeInstance

	Chain *storage.Chain

	Index    uint32
	Finished chan bool
}

func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}
	for _, handler := range []interface{}{protocol.HandlePrompt, protocol.HandleTerminate} {
		if err := protocol.RegisterHandler(handler); err != nil {
			return nil, err
		}
	}
	return protocol, nil
}

func (p *Protocol) decrypt() ([]abstract.Point, error) {
	boxes, err := p.Chain.Boxes()
	if err != nil {
		return nil, err
	}

	ballots := boxes[0].Ballots

	decrypted := make([]abstract.Point, len(ballots))
	for i := range decrypted {
		secret := api.Suite.Point().Mul(ballots[i].Alpha, p.Chain.SharedSecret.V)
		message := api.Suite.Point().Sub(ballots[i].Beta, secret)

		decrypted[i] = message
	}

	return decrypted, nil
}

func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{}})
}

func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	if p.IsRoot() {
		return p.SendToChildren(&prompt.Prompt)
	}

	points, err := p.decrypt()
	if err != nil {
		return err
	}

	return p.SendTo(p.Parent(), &Terminate{p.Chain.SharedSecret.Index, points})
}

func (p *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	points, err := p.decrypt()
	if err != nil {
		return err
	}

	clear := make([][]byte, len(points))
	for i := range points {
		shares := make([]*share.PubShare, len(terminates)+1)
		shares[0] = &share.PubShare{I: p.Chain.SharedSecret.Index, V: points[i]}
		for j, terminate := range terminates {
			shares[j+1] = &share.PubShare{
				I: terminate.Terminate.Index,
				V: terminate.Terminate.Decrypted[i],
			}
		}

		message, err := share.RecoverCommit(api.Suite, shares, 2, 3)
		if err != nil {
			return err
		}

		data, _ := message.Data()
		clear[i] = data
	}

	boxes, _ := p.Chain.Boxes()
	shuffle := boxes[0].Ballots

	ballots := make([]*api.BallotNew, len(points))
	for i := range ballots {
		ballots[i] = shuffle[i]
		ballots[i].Clear = clear[i]
	}

	index, _ := p.Chain.Store(&api.BoxNew{ballots})
	p.Index = uint32(index)
	p.Finished <- true

	return nil
}