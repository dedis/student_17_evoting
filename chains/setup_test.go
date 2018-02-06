package chains

import (
	"testing"

	"gopkg.in/dedis/onet.v1"
)

var roster *onet.Roster

var master *Master
var election *Election

func TestMain(m *testing.M) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	_, roster, _ = local.GenTree(3, true)

	// Setup master Skipchain.
	genesis, _ := client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
	master = &Master{nil, genesis.Hash, roster, []User{0}}
	rep, _ := client.StoreSkipBlock(genesis, roster, master)
	client.StoreSkipBlock(rep.Latest, roster, &Link{nil})

	// Setup election Skipchain.
	genesis, _ = client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
	election = &Election{"", 0, []User{0}, genesis.Hash, roster, nil, nil, 0, "", ""}
	rep, _ = client.StoreSkipBlock(genesis, roster, election)
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, &Ballot{0, nil, nil})
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, &Box{nil})
	client.StoreSkipBlock(rep.Latest, roster, &Box{nil})

	m.Run()
}
