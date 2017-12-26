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

// Protocol is the main object for the decryption protocol. All handlers are
// attached to it. Furthermore, it can be used to pass over values from the service
// upon initialization and it contains the protocol result after the protocol has
// finished.
type Protocol struct {
	*onet.TreeNodeInstance

	// Secret is the node's shared secret from the DKG protocol.
	Secret *dkg.SharedSecret

	// Shuffle is a box containing the shuffled (re-encrypted) ballots.
	Shuffle *chains.Box
	// Decryption contains the result, fully decrypted ballots, after termination.
	Decryption *chains.Box

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

// Start is called on the root node prompting it to send a Prompt message to itself
// and the leaf nodes containing the shuffled ballots.
func (p *Protocol) Start() error {
	return p.HandlePrompt(MessagePrompt{p.TreeNode(), Prompt{p.Shuffle.Ballots}})
}

// HandlePrompt is executed upon receiving a prompt message. If the receiver is the
// root node it is sent on to the leaf nodes, otherwise the leaf node perform their
// decryption on the shuffled ballots with their respective secrets and send the
// results in a Terminate message back to the root node.
func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	if p.IsRoot() {
		return p.SendToChildren(&prompt.Prompt)
	}

	points := p.decrypt(prompt.Shuffle)
	return p.SendTo(p.Parent(), &Terminate{p.Secret.Index, points})
}

// HandleTerminate is executed by the root and completes the protocol. The root node
// first decrypts the shuffled ballots with its secret before reconstructing the
// plaintexts by a Lagrange interpolation using the decrypted ballots from
// leaf nodes. The result is set in the protocol structure before signaling
// the protocol termination.
func (p *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	points := p.decrypt(p.Shuffle.Ballots)

	clear := make([][]byte, len(points))
	for i := range points {
		// Every point has to be reconstructed separately
		shares := make([]*share.PubShare, len(terminates)+1)
		shares[0] = &share.PubShare{I: p.Secret.Index, V: points[i]}
		for j, terminate := range terminates {
			shares[j+1] = &share.PubShare{
				I: terminate.Terminate.Index,
				V: terminate.Terminate.Decrypted[i],
			}
		}

		// Lagrange interpolation.
		message, err := share.RecoverCommit(crypto.Suite, shares, 2, len(points))
		if err != nil {
			return err
		}

		clear[i], _ = message.Data()
	}

	texts := make([]*chains.Text, len(points))
	for i := range texts {
		texts[i] = &chains.Text{p.Shuffle.Ballots[i].User, clear[i]}
	}

	p.Decryption = &chains.Box{Ballots: nil, Texts: texts}
	p.Finished <- true
	return nil
}

// decrypt is a helper function decrypting a a list of ballots using the nodes
// secret key.
func (p *Protocol) decrypt(shuffle []*chains.Ballot) []abstract.Point {
	decrypted := make([]abstract.Point, len(shuffle))
	for i := range decrypted {
		decrypted[i] = crypto.Decrypt(p.Secret.V, shuffle[i].Alpha, shuffle[i].Beta)
	}
	return decrypted
}
