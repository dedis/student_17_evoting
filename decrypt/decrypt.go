package decrypt

import (
	"crypto/cipher"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/crypto.v0/share"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/dkg"
)

type Protocol struct {
	*onet.TreeNodeInstance

	Secret *dkg.SharedSecret

	Shuffle    *chains.Box
	Decryption *chains.Box

	Finished chan bool
}

var suite abstract.Suite
var stream cipher.Stream

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	onet.GlobalProtocolRegister(Name, New)

	suite = ed25519.NewAES128SHA256Ed25519(false)
	stream = suite.Cipher(abstract.RandomKey)
}

func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}
	protocol.RegisterHandler(protocol.HandlePrompt)
	protocol.RegisterHandler(protocol.HandleTerminate)
	return protocol, nil
}

func (p *Protocol) decrypt(shuffle []*chains.Ballot) ([]abstract.Point, error) {
	decrypted := make([]abstract.Point, len(shuffle))
	for i := range decrypted {
		secret := suite.Point().Mul(shuffle[i].Alpha, p.Secret.V)
		message := suite.Point().Sub(shuffle[i].Beta, secret)

		decrypted[i] = message
	}

	return decrypted, nil
}

func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{p.Shuffle.Ballots}})
}

func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	if p.IsRoot() {
		return p.SendToChildren(&prompt.Prompt)
	}

	points, err := p.decrypt(prompt.Shuffle)
	if err != nil {
		return err
	}

	return p.SendTo(p.Parent(), &Terminate{p.Secret.Index, points})
}

func (p *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	points, err := p.decrypt(p.Shuffle.Ballots)
	if err != nil {
		return err
	}

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

		message, err := share.RecoverCommit(suite, shares, 2, 3)
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
