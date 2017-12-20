package chains

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
)

var client *skipchain.Client
var verifier []skipchain.VerifierID

func init() {
	client = skipchain.NewClient()
	verifier = skipchain.VerificationStandard
}

func New(roster *onet.Roster, data interface{}) (*skipchain.SkipBlock, error) {
	return client.CreateGenesis(roster, 1, 1, verifier, data, nil)
}

func chain(roster *onet.Roster, id skipchain.SkipBlockID) ([]*skipchain.SkipBlock, error) {
	chain, err := client.GetUpdateChain(roster, id)
	if err != nil {
		return nil, err
	}
	return chain.Update, nil
}
