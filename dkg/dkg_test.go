package dkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestProtocol(t *testing.T) {
	for _, nodes := range []int{3, 5, 10} {
		setupDKG(t, nodes)
	}
}

func setupDKG(t *testing.T, nbrNodes int) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, _, tree := local.GenBigTree(nbrNodes, nbrNodes, nbrNodes, true)

	pi, err := local.CreateProtocol(Name, tree)
	protocol := pi.(*Protocol)
	protocol.Wait = true
	if err != nil {
		t.Fatal("Couldn't start protocol:", err)
	}
	log.ErrFatal(pi.Start())

	timeout := network.WaitRetry * 2 * time.Second
	select {
	case <-protocol.Done:
		require.NotNil(t, protocol.DKG)
	case <-time.After(timeout):
		t.Fatal("Didn't finish in time")
	}
}
