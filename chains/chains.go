package chains

import (
	"github.com/qantik/nevv/dkg"
	"gopkg.in/dedis/cothority.v1/skipchain"
	rabin "gopkg.in/dedis/crypto.v0/share/dkg"
	"gopkg.in/dedis/onet.v1"
)

var client = skipchain.NewClient()

// New creates a new skipchain for a given roster and stores data in the genesis block.
func New(roster *onet.Roster, data interface{}) (*skipchain.SkipBlock, error) {
	return client.CreateGenesis(roster, 1, 1, skipchain.VerificationStandard, data, nil)
}

// chain returns a skipchain for a given id.
func chain(roster *onet.Roster, id skipchain.SkipBlockID) ([]*skipchain.SkipBlock, error) {
	chain, err := client.GetUpdateChain(roster, id)
	if err != nil {
		return nil, err
	}
	return chain.Update, nil
}

// GenerateElectionChain creates an election and its corresponding skipchain for a given
// stage. This is only used for testing purposes to quicky set up an environment.
func GenElectionChain(roster *onet.Roster, creator uint32, users []uint32,
	numBallots, stage int) (
	*Election, []*rabin.DistKeyGenerator) {

	chain, _ := New(roster, nil)

	n := len(roster.List)
	dkgs := dkg.Simulate(n, n-1)
	s, _ := dkg.NewSharedSecret(dkgs[0])

	election := &Election{
		Key:     s.X,
		ID:      chain.Hash,
		Roster:  roster,
		Creator: creator,
		Users:   users,
		Stage:   uint32(stage),
		Data:    []byte{},
	}

	box := genBox(s.X, numBallots)
	mixes := box.genMix(s.X, n)
	partials := mixes[n-1].genPartials(dkgs)

	election.Store(election)
	election.storeBallots(box.Ballots)

	if stage == SHUFFLED {
		election.storeMixes(mixes)
	} else if stage == DECRYPTED {
		election.storeMixes(mixes)
		election.storePartials(partials)
	}
	return election, dkgs
}

// GenMasterChain creates a master object and its corresponsing skipchain for a given
// roster. This is only used for testing purposes to quicky set up an environment.
func GenMasterChain(roster *onet.Roster, links ...skipchain.SkipBlockID) *Master {
	chain, _ := New(roster, nil)

	master := &Master{ID: chain.Hash, Roster: roster}
	master.Store(master)

	for _, link := range links {
		master.Store(&Link{ID: link})
	}
	return master
}

func reconstruct() {
	// for i := 0; i < 3; i++ {
	// 	shares := make([]*share.PubShare, 3)
	// 	for j, partial := range partials {
	// 		shares[j] = &share.PubShare{I: j, V: partial.Points[i]}
	// 	}

	// 	message, _ := share.RecoverCommit(crypto.Suite, shares, 3, 3)
	// 	fmt.Println(message.Data())
	// }

}
