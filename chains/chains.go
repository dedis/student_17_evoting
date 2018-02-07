package chains

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
)

var client = skipchain.NewClient()

// New creates a new skipchain for a given roster and stores data in the genesis block.
func New(roster *onet.Roster, data interface{}) (*skipchain.SkipBlock, error) {
	return client.CreateGenesis(roster, 1, 1, skipchain.VerificationStandard, data, nil)
}

// Store appends a new block containing the given data to a Skipchain identified by id.
func Store(roster *onet.Roster, id skipchain.SkipBlockID, data ...interface{}) error {
	chain, err := chain(roster, id)
	if err != nil {
		return err
	}

	latest := chain[len(chain)-1]
	for _, obj := range data {
		if resp, err := client.StoreSkipBlock(latest, roster, obj); err != nil {
			return err
		} else {
			latest = resp.Latest
		}
	}
	return nil
}

// chain returns a skipchain for a given id.
func chain(roster *onet.Roster, id skipchain.SkipBlockID) ([]*skipchain.SkipBlock, error) {
	chain, err := client.GetUpdateChain(roster, id)
	if err != nil {
		return nil, err
	}
	return chain.Update, nil
}
