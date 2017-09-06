package decrypt

/*
The `NewProtocol` method is used to define the protocol and to register
the handlers that will be called if a certain type of message is received.
The handlers will be treated according to their signature.

The protocol-file defines the actions that the protocol needs to do in each
step. The root-node will call the `Start`-method of the protocol. Each
node will only use the `Handle`-methods, and not call `Start` again.
*/

import (
	"errors"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/dkg"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/share"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(Announce{})
	network.RegisterMessage(Reply{})
	_, _ = onet.GlobalProtocolRegister(Name, NewProtocol)
}

// Template just holds a message that is passed to all children. It
// also defines a channel that will receive the number of children. Only the
// root-node will write to the channel.
type Template struct {
	*onet.TreeNodeInstance
	Message    string
	ChildCount chan int
	Shared     *dkg.SharedSecret
	Ballot     *api.Ballot
	Genesis    *skipchain.SkipBlock
}

type Config struct {
	Election string
}

// NewProtocol initialises the structure for use in one round
func NewProtocol(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	t := &Template{
		TreeNodeInstance: n,
		ChildCount:       make(chan int),
	}
	for _, handler := range []interface{}{t.HandleAnnounce, t.HandleReply} {
		if err := t.RegisterHandler(handler); err != nil {
			return nil, errors.New("couldn't register handler: " + err.Error())
		}
	}
	return t, nil
}

// Start sends the Announce-message to all children
func (p *Template) Start() error {
	log.Lvl3("Starting Template")
	return p.HandleAnnounce(StructAnnounce{p.TreeNode(),
		Announce{"cothority rulez!"}})
}

// HandleAnnounce is the first message and is used to send an ID that
// is stored in all nodes.
func (p *Template) HandleAnnounce(msg StructAnnounce) error {
	p.Message = msg.Message
	if !p.IsLeaf() {
		// If we have children, send the same message to all of them
		_ = p.SendToChildren(&msg.Announce)
	} else {
		// If we're the leaf, start to reply
		_ = p.HandleReply(nil)
	}
	return nil
}

func decrypt(secret abstract.Scalar, K, C abstract.Point) abstract.Point {
	S := api.Suite.Point().Mul(K, secret)
	M := api.Suite.Point().Sub(C, S)

	// point := api.Point{}
	// point.UnpackNorm(M)
	// point.Out()
	return M
}

// HandleReply is the message going up the tree and holding a counter
// to verify the number of nodes.
func (p *Template) HandleReply(reply []StructReply) error {
	defer p.Done()

	children := 1
	for _, c := range reply {
		children += c.ChildrenCount
	}

	if !p.IsRoot() {
		log.Lvl3("Sending to parent")
		return p.SendTo(p.Parent(), &Reply{children, p.Shared.Index, p.Shared.V})
	}

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(p.Genesis.Roster, p.Genesis.Hash)
	if err != nil {
		return err
	}

	// latest := chain.Update[2]
	latest := chain.Update[len(chain.Update)-1]
	_, blob, _ := network.Unmarshal(latest.Data)

	// collection := blob.(*api.Ballot)
	collection := blob.(*api.Box)
	alpha, beta := collection.Split()
	// K, C := collection.Alpha.Pack(), collection.Beta.Pack()
	K, C := alpha[0], beta[0]

	snek := p.Shared.V
	// K := p.Ballot.Alpha.Pack()
	// C := p.Ballot.Beta.Pack()

	shares := make([]*share.PubShare, 0)

	shares = append(shares, &share.PubShare{I: 0, V: decrypt(snek, K, C)})
	for _, c := range reply {
		shares = append(shares, &share.PubShare{I: c.I, V: decrypt(c.Secret, K, C)})
		ii, err := share.RecoverCommit(api.Suite, shares, 2, 3)
		log.Lvl3(err)

		point := api.Point{}
		point.UnpackNorm(ii)
		point.Out()
	}

	log.Lvl3("Root-node is done - nbr of children found:", children)
	p.ChildCount <- children
	return nil
}
