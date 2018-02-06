package chains

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
)

var client = skipchain.NewClient()
var verifier = skipchain.VerificationStandard

func New(roster *onet.Roster, data interface{}) (*skipchain.SkipBlock, error) {
	return client.CreateGenesis(roster, 1, 1, verifier, data, nil)
}

// Store appends a new block containing the given data to a Skipchain identified by id.
func Store(roster *onet.Roster, id skipchain.SkipBlockID, data interface{}) error {
	chain, err := chain(roster, id)
	if err != nil {
		return err
	}
	_, err = client.StoreSkipBlock(chain[len(chain)-1], roster, data)
	return err
}

func chain(roster *onet.Roster, id skipchain.SkipBlockID) ([]*skipchain.SkipBlock, error) {
	chain, err := client.GetUpdateChain(roster, id)
	if err != nil {
		return nil, err
	}
	return chain.Update, nil
}
