package shuffle

import (
	"crypto/cipher"
	"errors"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/dkg"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/crypto.v0/shuffle"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

// Protocol is the main structure for the shuffle procedure. It is initialized
// by the service and further used to retrieve the latest SkipBlock when the
// protocol has finished.
type Protocol struct {
	*onet.TreeNodeInstance

	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock
	Shared  *dkg.SharedSecret
	Done    chan bool
}

// Config is used to distributed a message from the root node
// to the other participants at inception of the protocol.
type Config struct {
}

var suite abstract.Suite
var stream cipher.Stream

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	_, _ = onet.GlobalProtocolRegister(Name, New)

	suite = ed25519.NewAES128SHA256Ed25519(false)
	stream = suite.Cipher(abstract.RandomKey)
}

// New creates a new shuffle protocol instance used by the service.
func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{
		TreeNodeInstance: node,
		Done:             make(chan bool),
	}

	for _, handler := range []interface{}{protocol.HandlePrompt, protocol.HandleTerminate} {
		if err := protocol.RegisterHandler(handler); err != nil {
			return nil, err
		}
	}

	return protocol, nil
}

// Start is the beginning point of the protocol. The root node creates a new Prompt
// message that is then passed to itself to effectily engage in the process.
func (protocol *Protocol) Start() error {
	log.Lvl3("Start shuffle protocol")

	prompt := Prompt{protocol.Genesis, protocol.Latest, protocol.Shared.X}
	message := MessagePrompt{protocol.TreeNode(), prompt}
	if err := protocol.HandlePrompt(message); err != nil {
		return err
	}

	return nil
}

// HandlePrompt is the handler function for the Prompt message. If the receiver
// is the root node it collects all the votes from the SkipChain, shuffles them
// and appends the shuffle to the chain before sending a Prompt to its child node.
// A child node only retrieves the latest block with the previous shuffle and appends
// its mix before again prompting its child node.
// In case the node is the leaf it sends a Terminate to the root after perfoming its
// shuffle.
func (protocol *Protocol) HandlePrompt(prompt MessagePrompt) error {
	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(prompt.Genesis.Roster, prompt.Genesis.Hash)
	if err != nil {
		return err
	}

	var alpha, beta []abstract.Point

	if protocol.IsRoot() {
		number := len(chain.Update) - 1
		if number < 2 {
			return errors.New("Not enough ballots (>= 2) to shuffle")
		}

		alpha = make([]abstract.Point, number)
		beta = make([]abstract.Point, number)

		for index := 1; index < number+1; index++ {
			data := chain.Update[index].Data
			_, blob, err := network.Unmarshal(data)
			if err != nil {
				return err
			}

			ballot := blob.(*api.Ballot)
			alpha[index-1] = ballot.Alpha.Pack()
			beta[index-1] = ballot.Beta.Pack()
		}
	} else {
		latest := chain.Update[len(chain.Update)-1]
		_, blob, err := network.Unmarshal(latest.Data)
		if err != nil {
			return err
		}

		collection := blob.(*api.Box)
		alpha, beta = collection.Split()
	}

	gamma, delta, _ := shuffle.Shuffle(suite, nil, prompt.Key, alpha, beta, stream)

	collection := &api.Box{}
	collection.Join(gamma, delta)
	reply, err := client.StoreSkipBlock(prompt.Latest, nil, collection)
	if err != nil {
		return err
	}

	if protocol.IsLeaf() {
		terminate := &Terminate{Latest: reply.Latest}
		if err := protocol.SendTo(protocol.Root(), terminate); err != nil {
			return err
		}
	} else {
		forward := &Prompt{Genesis: prompt.Genesis, Latest: reply.Latest, Key: prompt.Key}
		if err := protocol.SendToChildren(forward); err != nil {
			return err
		}
	}

	return nil
}

// HandleTerminate is used by the root node after receiving a Terminate message
// from the leaf node to switch the channel boolean to true which in turn invokes
// the service that was waiting for the protocol to complete.
func (protocol *Protocol) HandleTerminate(terminate MessageTerminate) error {
	protocol.Latest = terminate.Latest
	protocol.Done <- true

	return nil
}
