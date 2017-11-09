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
		protocol(t, nodes)
	}
}

func protocol(t *testing.T, nodes int) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, _, tree := local.GenBigTree(nodes, nodes, nodes, true)

	pi, err := local.CreateProtocol(Name, tree)
	if err != nil {
		t.Fatal("Couldn't start protocol:", err)
	}

	protocol := pi.(*Protocol)
	protocol.Wait = true
	log.ErrFatal(pi.Start())

	timeout := network.WaitRetry * 2 * time.Second
	select {
	case <-protocol.Done:
		require.NotNil(t, protocol.DKG)
	case <-time.After(timeout):
		t.Fatal("Didn't finish in time")
	}
}
