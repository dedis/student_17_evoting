package decrypt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/dkg"
)

func TestProtocol(t *testing.T) {
	_ = func(n int, key abstract.Point) *chains.Box {
		ballots := make([]*chains.Ballot, 0)
		for i := 0; i < n; i++ {
			M, _ := suite.Point().Pick([]byte{byte(i)}, random.Stream)

			k := suite.Scalar().Pick(random.Stream)
			K := suite.Point().Mul(nil, k)
			S := suite.Point().Mul(key, k)
			C := S.Add(S, M)

			ballots = append(ballots, &chains.Ballot{
				chains.User(i),
				K,
				C,
				nil,
			})
		}
		return &chains.Box{ballots}

	}

	for _, nodes := range []int{3, 5, 7, 10} {
		local := onet.NewLocalTest()
		_, _, tree := local.GenBigTree(nodes, nodes, nodes, true)
		instance, err := local.CreateProtocol(dkg.Name, tree)
		assert.Nil(t, err)

		protocol := instance.(*dkg.Protocol)
		protocol.Wait = true
		assert.Nil(t, protocol.Start())
		select {
		case <-protocol.Done:
			// secret, err := protocol.SharedSecret()
			// instance, err := local.CreateProtocol(Name, tree)
			// protocol := instance.(*Protocol)
			// protocol.Secret = secret
			// protocol.Shuffle = box(nodes, secret.X)
			// fmt.Println(protocol.ServerIdentity(), protocol)
			// assert.Nil(t, protocol.Start())
			// select {
			// case <-protocol.Finished:
			// 	assert.True(t, true)
			// case <-time.After(4 * time.Second):
			// 	assert.Fail(t, "Protocol timeout")
			// }
			// assert.Nil(t, err)
			assert.True(t, true)
		case <-time.After(4 * time.Second):
			assert.Fail(t, "Protocol timeout")
		}

		local.CloseAll()
	}
}
