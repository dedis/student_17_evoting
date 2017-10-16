package decryptnew

import (
	"github.com/dedis/onet/log"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/storage"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/share"
	"gopkg.in/dedis/onet.v1"
)

func init() {

}

type Protocol struct {
	*onet.TreeNodeInstance

	Chain *storage.Chain

	Finished chan bool
}

func (p *Protocol) decrypt() ([]abstract.Point, error) {
	boxes, err := p.Chain.Boxes()
	if err != nil {
		return nil, err
	}

	ballots := boxes[1].Ballots

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

		data, err := message.Data()
		log.Lvl3("DATA", data, err)
	}
}
