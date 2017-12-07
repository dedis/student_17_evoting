package chains

import (
	"errors"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

// Create creates a new Skipchain with standard verfication and marshable
// data in the genesis block. It returns said genesis Skipblock.
func Create(roster *onet.Roster, data interface{}) (*skipchain.SkipBlock, error) {
	client := skipchain.NewClient()
	genesis, err := client.CreateGenesis(roster, 1, 1,
		skipchain.VerificationStandard, data, nil)
	if err != nil {
		return nil, err
	}
	return genesis, nil
}

// Stores appends a new Skipblock containing marshable data to a Skipchain
// identified by id. It returns the index of the new Skipblock.
func Store(roster *onet.Roster, id skipchain.SkipBlockID, data interface{}) (int, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return -1, err
	}

	client := skipchain.NewClient()
	reply, err := client.StoreSkipBlock(chain[len(chain)-1], roster, data)
	if err != nil {
		return -1, err
	}
	return reply.Latest.Index, nil
}

// GetElection retrieves the election object from its Skipchain identified by
// id. It then returns said election object.
func GetElection(roster *onet.Roster, id skipchain.SkipBlockID) (*Election, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	// By definition the election object is in the second Skipblock.
	_, blob, err := network.Unmarshal(chain[1].Data)
	if err != nil {
		return nil, err
	}
	return blob.(*Election), nil
}

// GetMaster retrieves the master object from its Skipchain identified by id.
// It then return said master object.
func GetMaster(roster *onet.Roster, id skipchain.SkipBlockID) (*Master, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	// By definition the master object is stored in the genesis Skipblock.
	_, blob, err := network.Unmarshal(chain[0].Data)
	if err != nil {
		return nil, err
	}
	return blob.(*Master), nil
}

// GetLinks retrieves the links from a master Skipchain identified by id.
// It then returns them in an array.
func GetLinks(roster *onet.Roster, id skipchain.SkipBlockID) ([]*Link, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	// Links immediately follow the genesis Skipblock.
	links := make([]*Link, 0)
	for i := 1; i < len(chain); i++ {
		_, blob, err := network.Unmarshal(chain[i].Data)
		if err != nil {
			return nil, err
		}

		links = append(links, blob.(*Link))
	}
	return links, nil
}

// GetBallots retrieves the ballots from their (unfinalized) election Skipchain
// identified by id. Only the last casted ballot of a user is regarded as valid
// and thus included in the returned array.
func GetBallots(roster *onet.Roster, id skipchain.SkipBlockID) ([]*Ballot, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	// Use map to only included a user's last ballot.
	mapping := make(map[User]*Ballot)
	for i := 2; i < len(chain); i++ {
		_, blob, err := network.Unmarshal(chain[i].Data)
		if err != nil {
			return nil, err
		}

		ballot := blob.(*Ballot)
		mapping[ballot.User] = blob.(*Ballot)
	}

	ballots := make([]*Ballot, 0)
	for _, ballot := range mapping {
		ballots = append(ballots, ballot)
	}

	return ballots, nil
}

// GetBox retrieves a box of the given kind from a (finalized or unfinalized)
// election Skipchain identified by id and then returned.
func GetBox(roster *onet.Roster, id skipchain.SkipBlockID, kind int32) (*Box, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	// Aggregate boxes
	boxes := make([]*Box, 0)
	for i := 2; i < len(chain); i++ {
		_, blob, err := network.Unmarshal(chain[i].Data)
		if err != nil {
			return nil, err
		}

		box, ok := blob.(*Box)
		if ok {
			boxes = append(boxes, box)
		}
	}

	// Check if box is available.
	size := len(boxes)
	if size >= 1 && kind == BALLOTS {
		return boxes[0], nil
	}
	if size >= 2 && kind == SHUFFLE {
		return boxes[1], nil
	}
	if size >= 3 && kind == DECRYPTION {
		return boxes[2], nil
	}

	// Aggregate ballots if not finalized yet
	if size == 0 && kind == BALLOTS {
		ballots, err := GetBallots(roster, id)
		if err != nil {
			return nil, err
		}

		return &Box{ballots}, nil
	}

	return nil, errors.New("Aggregation not available, need to finalize first")
}

// chain is a helper function that retrieves a Skipchain. Returning it
// as a list of SkipBlocks
func chain(roster *onet.Roster, id skipchain.SkipBlockID) ([]*skipchain.SkipBlock, error) {
	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(roster, id)
	if err != nil {
		return nil, err
	}
	return chain.Update, nil
}
