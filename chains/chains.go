package chains

import (
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

	links := make([]*Link, len(chain)-1)
	for i := 1; i <= len(links); i++ {
		_, blob, err := network.Unmarshal(chain[i].Data)
		if err != nil {
			return nil, err
		}

		links[i-1] = blob.(*Link)
	}
	return links, nil
}

func GetBallots(roster *onet.Roster, id skipchain.SkipBlockID) error {
	return nil
}

func GetBoxes(roster *onet.Roster, id skipchain.SkipBlockID) error {
	return nil
}
