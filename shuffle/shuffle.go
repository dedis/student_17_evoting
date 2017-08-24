package shuffle

import (
	"crypto/cipher"
	"errors"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
	"gopkg.in/dedis/crypto.v0/shuffle"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

// Protocol ...
type Protocol struct {
	*onet.TreeNodeInstance
	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock
	Done    chan bool
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

// New ...
func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{
		TreeNodeInstance: node,
		Done:             make(chan bool),
	}

	for _, handler := range []interface{}{protocol.HandlePrompt, protocol.HandleTerminate} {
		if err := protocol.RegisterHandler(handler); err != nil {
			log.Lvl3("Could not register handler", err.Error())
			return nil, err
		}
	}

	return protocol, nil
}

// Start ...
func (protocol *Protocol) Start() error {
	log.Lvl3("Start shuffle protocol")

	prompt := Prompt{protocol.Genesis, protocol.Latest}
	message := MessagePrompt{protocol.TreeNode(), prompt}
	if err := protocol.HandlePrompt(message); err != nil {
		return err
	}

	return nil
}

func (protocol *Protocol) HandlePrompt(prompt MessagePrompt) error {
	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(prompt.Genesis.Roster, prompt.Genesis.Hash)
	if err != nil {
		return err
	}

	var alpha []abstract.Point
	var beta []abstract.Point

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
			alpha[index-1] = ballot.Alpha
			beta[index-1] = ballot.Beta
		}
	} else {
		latest := chain.Update[len(chain.Update)-1]
		_, blob, _ := network.Unmarshal(latest.Data)
		if err != nil {
			return err
		}

		collection := blob.(*api.Collection)
		alpha, beta = collection.Split()
	}

	log.Lvl3(protocol.ServerIdentity(), "Alpha:", alpha, "Beta:", beta)

	gamma, delta, _ := shuffle.Shuffle(suite, nil, nil, alpha, beta, stream)
	log.Lvl3(protocol.ServerIdentity(), "Gamma:", gamma, "Delta:", delta)

	collection := &api.Collection{}
	collection.Join(gamma, delta)
	reply, err := client.StoreSkipBlock(prompt.Latest, nil, collection)
	if err != nil {
		return err
	}

	log.Lvl3("Stored shuffle at", reply.Latest.Index)

	if protocol.IsLeaf() {
		terminate := &Terminate{Latest: reply.Latest}
		if err := protocol.SendTo(protocol.Root(), terminate); err != nil {
			return err
		}
	} else {
		forward := &Prompt{Genesis: prompt.Genesis, Latest: reply.Latest}
		if err := protocol.SendToChildren(forward); err != nil {
			return err
		}
	}

	return nil
}

func (protocol *Protocol) HandleTerminate(terminate MessageTerminate) error {
	protocol.Latest = terminate.Latest
	protocol.Done <- true

	return nil
}
