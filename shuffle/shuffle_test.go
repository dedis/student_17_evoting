package shuffle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func TestProtocol(t *testing.T) {
	box := func(n int) *chains.Box {
		ballots := make([]*chains.Ballot, 0)
		for i := 0; i < n; i++ {
			ballots = append(ballots, &chains.Ballot{
				chains.User(i),
				crypto.Suite.Point(),
				crypto.Suite.Point(),
			})
		}
		return &chains.Box{Ballots: ballots}

	}

	for _, nodes := range []int{3, 5, 7, 10} {
		local := onet.NewLocalTest()
		_, _, tree := local.GenBigTree(nodes, nodes, 1, true)
		instance, err := local.CreateProtocol(Name, tree)
		assert.Nil(t, err)

		protocol := instance.(*Protocol)
		protocol.Key = crypto.Suite.Point()
		protocol.Box = box(nodes)
		assert.Nil(t, protocol.Start())

		select {
		case <-protocol.Finished:
			assert.Equal(t, nodes, len(protocol.Shuffle.Ballots))
		case <-time.After(2 * time.Second):
			assert.Fail(t, "Protocol timeout")
		}

		local.CloseAll()
	}
}
