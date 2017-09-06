package storage

import (
	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/dkg"
)

// Election is the base data structure of the application. It comprises
// for each involved conode the genesis and the latest appended block as
// well as the generated shared secret from the distributed key generation
// protocol which is run at the inception of a new election.
type Election struct {
	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock

	*dkg.SharedSecret
}

func (election *Election) GetSkipChain() (*skipchain.GetUpdateChainReply, error) {
	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(election.Genesis.Roster, election.Genesis.Hash)
	if err != nil {
		return nil, err
	}

	return chain, nil
}

func (election *Election) GetLastBlock() (*skipchain.SkipBlock, error) {
	chain, err := election.GetSkipChain()
	if err != nil {
		return nil, err
	}

	return chain.Update[len(chain.Update)-1], nil
}
