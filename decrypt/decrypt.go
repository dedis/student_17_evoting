package decrypt

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/share"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/dkg"
)

type Protocol struct {
	*onet.TreeNodeInstance

	Secret *dkg.SharedSecret

	Shuffle    *chains.Box
	Decryption *chains.Box

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

func (p *Protocol) decrypt(shuffle []*chains.Ballot) []abstract.Point {
	decrypted := make([]abstract.Point, len(shuffle))
	for i := range decrypted {
		secret := crypto.Suite.Point().Mul(shuffle[i].Alpha, p.Secret.V)
		message := crypto.Suite.Point().Sub(shuffle[i].Beta, secret)
		decrypted[i] = message
	}
	return decrypted
}

func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{p.Shuffle.Ballots}})
}

func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	if p.IsRoot() {
		return p.SendToChildren(&prompt.Prompt)
	}

	points := p.decrypt(prompt.Shuffle)
	return p.SendTo(p.Parent(), &Terminate{p.Secret.Index, points})
}

func (p *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	points := p.decrypt(p.Shuffle.Ballots)

	clear := make([][]byte, len(points))
	for i := range points {
		shares := make([]*share.PubShare, len(terminates)+1)
		shares[0] = &share.PubShare{I: p.Secret.Index, V: points[i]}
		for j, terminate := range terminates {
			shares[j+1] = &share.PubShare{
				I: terminate.Terminate.Index,
				V: terminate.Terminate.Decrypted[i],
			}
		}

		message, err := share.RecoverCommit(crypto.Suite, shares, 2, 3)
		if err != nil {
			return err
		}

		data, _ := message.Data()
		clear[i] = data
	}

	ballots := make([]*chains.Ballot, len(points))
	for i := range ballots {
		ballots[i] = p.Shuffle.Ballots[i]
		ballots[i].Text = clear[i]
	}

	p.Decryption = &chains.Box{ballots}
	p.Finished <- true

	return nil
}
