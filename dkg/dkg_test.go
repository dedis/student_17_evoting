package dkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
)

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
	pi.Start()

	select {
	case <-protocol.Done:
		_, err := protocol.SharedSecret()
		assert.Nil(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Didn't finish in time")
	}
}
