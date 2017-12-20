package chains

import (
	"encoding/base64"
	"testing"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
)

var roster *onet.Roster

var masterGenesis *skipchain.SkipBlock
var master *Master

var electionGenesis *skipchain.SkipBlock
var election *Election

func TestMain(m *testing.M) {
	local := onet.NewTCPTest()
	defer local.CloseAll()

	_, roster, _ = local.GenTree(3, true)

	// Setup master Skipchain.
	masterGenesis, _ = client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
	master = &Master{nil, masterGenesis.Hash, roster, []User{0}}
	rep, _ := client.StoreSkipBlock(masterGenesis, roster, master)
	client.StoreSkipBlock(rep.Latest, roster, &Link{nil})

	// Setup election Skipchain.
	electionGenesis, _ = client.CreateGenesis(roster, 1, 1, verifier, nil, nil)
	id := base64.StdEncoding.EncodeToString(electionGenesis.Hash)
	election = &Election{"", 0, []User{0}, id, roster, nil, nil, 0, "", ""}
	rep, _ = client.StoreSkipBlock(electionGenesis, roster, election)
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, &Ballot{0, nil, nil, nil})
	rep, _ = client.StoreSkipBlock(rep.Latest, roster, &Box{nil})
	client.StoreSkipBlock(rep.Latest, roster, &Box{nil})

	m.Run()
}
