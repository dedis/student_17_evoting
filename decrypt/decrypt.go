package decrypt

import (
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/storage"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/share"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	_, _ = onet.GlobalProtocolRegister(Name, New)
}

type Protocol struct {
	*onet.TreeNodeInstance

	Storage      *storage.Storage
	Election     *storage.Election
	ElectionName string

	Done chan bool
}

type Config struct {
}

func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Done: make(chan bool)}
	for _, handler := range []interface{}{protocol.HandlePrompt, protocol.HandleTerminate} {
		if err := protocol.RegisterHandler(handler); err != nil {
			return nil, err
		}
	}
	return protocol, nil
}

func (protocol *Protocol) Start() error {
	log.Lvl3("Starting decryption")
	prompt := Prompt{protocol.ElectionName}
	return protocol.HandlePrompt(MessagePrompt{protocol.TreeNode(), prompt})
}

func (protocol *Protocol) HandlePrompt(prompt MessagePrompt) error {
	election, _ := protocol.Storage.Get(prompt.ElectionName)
	protocol.Election = election

	if protocol.IsRoot() {
		return protocol.SendToChildren(&prompt.Prompt)
	}

	points, index, err := protocol.decryptShuffle()
	if err != nil {
		return err
	}

	return protocol.SendTo(protocol.Parent(), &Terminate{index, points})
}

// Load the last shuffle from the skipchain and perform the ElGamal decryption
// algorithm on each ballot. Returns a slice of encrypted points and the index of
// conode within the DKG.
func (protocol *Protocol) decryptShuffle() ([]abstract.Point, int, error) {
	latest, err := protocol.Election.GetLastBlock()
	if err != nil {
		return nil, -1, err
	}

	_, blob, _ := network.Unmarshal(latest.Data)
	box := blob.(*api.Box)
	alpha, beta := box.Split()

	decrypted := make([]abstract.Point, len(alpha))
	for index := range decrypted {
		secret := api.Suite.Point().Mul(alpha[index], protocol.Election.SharedSecret.V)
		message := api.Suite.Point().Sub(beta[index], secret)

		decrypted[index] = message
	}

	return decrypted, protocol.Election.SharedSecret.Index, nil
}

func (protocol *Protocol) HandleTerminate(terminates []MessageTerminate) error {
	points, index, err := protocol.decryptShuffle()
	if err != nil {
		return err
	}

	for i := range points {
		shares := make([]*share.PubShare, len(terminates)+1)
		shares[0] = &share.PubShare{I: index, V: points[i]}
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

		point := api.Point{}
		point.UnpackNorm(message)
		point.Out()
	}

	protocol.Done <- true

	return nil
}
