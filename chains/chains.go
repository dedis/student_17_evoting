package chains

import (
	"errors"

	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func chain(roster *onet.Roster, id skipchain.SkipBlockID) ([]*skipchain.SkipBlock, error) {
	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(roster, id)
	if err != nil {
		return nil, err
	}
	return chain.Update, nil
}

func Create(roster *onet.Roster, data interface{}) (*skipchain.SkipBlock, error) {
	client := skipchain.NewClient()
	genesis, err := client.CreateGenesis(roster, 1, 1,
		skipchain.VerificationStandard, data, nil)
	if err != nil {
		return nil, err
	}
	return genesis, nil
}

func Store(roster *onet.Roster, id skipchain.SkipBlockID, data interface{}) (int, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return -1, nil
	}

	client := skipchain.NewClient()
	reply, err := client.StoreSkipBlock(chain[len(chain)-1], roster, data)
	if err != nil {
		return -1, err
	}
	return reply.Latest.Index, nil
}

func GetElection(roster *onet.Roster, id skipchain.SkipBlockID) (*Election, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	_, blob, err := network.Unmarshal(chain[1].Data)
	if err != nil {
		return nil, err
	}
	return blob.(*Election), nil
}

func GetMaster(roster *onet.Roster, id skipchain.SkipBlockID) (*Master, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	_, blob, err := network.Unmarshal(chain[0].Data)
	if err != nil {
		return nil, err
	}
	return blob.(*Master), nil
}

func GetLinks(roster *onet.Roster, id skipchain.SkipBlockID) ([]*Link, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

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

func GetBallots(roster *onet.Roster, id skipchain.SkipBlockID) ([]*Ballot, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

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
