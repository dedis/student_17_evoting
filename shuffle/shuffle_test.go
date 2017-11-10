package shuffle

import (
	"testing"
	"time"

	"github.com/dedis/cothority/skipchain"
	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/storage"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestProtocol(t *testing.T) {
	for _, nodes := range []int{3, 5, 7} {
		local := onet.NewLocalTest()
		_, roster, tree := local.GenBigTree(nodes, nodes, 1, true)
		instance, err := local.CreateProtocol(Name, tree)
		log.ErrFatal(err)

		protocol := instance.(*Protocol)
		protocol.Chain = newChain(roster)
		log.ErrFatal(protocol.Start())

		timeout := network.WaitRetry * 2 * time.Second
		select {
		case <-protocol.Finished:
			assert.Equal(t, 4, protocol.Index)
		case <-time.After(timeout):
			t.Fatal("Shuffle timeout")
		}

		local.CloseAll()
	}
}

func newChain(roster *onet.Roster) *storage.Chain {
	client := skipchain.NewClient()
	genesis, _ := client.CreateGenesis(roster, 1, 1, skipchain.VerificationNone, nil, nil)
	election := api.Election{
		ID:          "test",
		Admin:       "admin",
		Start:       "",
		End:         "",
		Data:        []byte{},
		Roster:      roster,
		Users:       []string{"user1", "user2"},
		Key:         nil,
		Description: "",
	}

	_, _ = client.StoreSkipBlock(genesis, nil, &election)
	chain := &storage.Chain{Genesis: genesis}

	ballot1 := &api.Ballot{"b1", suite.Point(), suite.Point(), []byte("b1")}
	_, _ = chain.Store(ballot1)
	ballot2 := &api.Ballot{"b2", suite.Point(), suite.Point(), []byte("b2")}
	_, _ = chain.Store(ballot2)

	return chain
}
