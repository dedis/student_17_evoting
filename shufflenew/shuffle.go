package shufflenew

import (
	"errors"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/storage"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/proof"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/crypto.v0/shuffle"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

type Protocol struct {
	*onet.TreeNodeInstance

	Chain *storage.Chain

	Index    int
	Finished chan bool
}

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	_, _ = onet.GlobalProtocolRegister(Name, New)
}

func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}

	for _, handler := range []interface{}{protocol.HandlePrompt, protocol.HandleTerminate} {
		if err := protocol.RegisterHandler(handler); err != nil {
			return nil, err
		}
	}

	return protocol, nil
}

func (p *Protocol) shuffle(key abstract.Point, X, Y []abstract.Point) (
	XX, YY []abstract.Point, pi []int, P proof.Prover) {

	k := len(X)

	ps := shuffle.PairShuffle{}
	ps.Init(api.Suite, k)

	pi = make([]int, k)
	for i := 0; i < k; i++ {
		pi[i] = i
	}

	for i := k - 1; i > 0; i-- {
		j := int(random.Uint64(api.Stream) % uint64(i+1))
		if j != i {
			t := pi[j]
			pi[j] = pi[i]
			pi[i] = t
		}
	}

	beta := make([]abstract.Scalar, k)
	for i := 0; i < k; i++ {
		beta[i] = api.Suite.Scalar().Pick(api.Stream)
	}

	Xbar, Ybar := make([]abstract.Point, k), make([]abstract.Point, k)
	for i := 0; i < k; i++ {
		Xbar[i] = api.Suite.Point().Mul(nil, beta[pi[i]])
		Xbar[i].Add(Xbar[i], X[pi[i]])
		Ybar[i] = api.Suite.Point().Mul(key, beta[pi[i]])
		Ybar[i].Add(Ybar[i], Y[pi[i]])
	}

	prover := func(ctx proof.ProverContext) error {
		return ps.Prove(pi, nil, key, beta, X, Y, api.Stream, ctx)
	}

	return Xbar, Ybar, pi, prover
}

func (p *Protocol) Start() error {
	ballots, err := p.Chain.Ballots()
	if err != nil {
		return err
	}

	if len(ballots) < 2 {
		return errors.New("Not enough (> 1) ballots to shuffle")
	}

	msg := MessagePrompt{p.TreeNode(), Prompt{p.Chain.Election().Key, ballots}}
	if err := p.HandlePrompt(msg); err != nil {
		return err
	}

	return nil
}

func (p *Protocol) HandlePrompt(prompt MessagePrompt) error {
	k := len(prompt.Ballots)
	alpha, beta := make([]abstract.Point, k), make([]abstract.Point, k)

	for i, ballot := range prompt.Ballots {
		alpha[i] = ballot.Alpha
		beta[i] = ballot.Beta
	}

	gamma, delta, pi, _ := p.shuffle(prompt.Key, alpha, beta)

	// Reconstruct ballot list with shuffle permutation
	shuffled := make([]*api.BallotNew, k)
	for i := range shuffled {
		shuffled[i] = &api.BallotNew{
			User:  prompt.Ballots[pi[i]].User,
			Alpha: gamma[i],
			Beta:  delta[i],
		}
	}

	if p.IsLeaf() {
		return p.SendTo(p.Root(), &Terminate{shuffled})
	}

	return p.SendToChildren(&Prompt{prompt.Key, shuffled})
}

func (p *Protocol) HandleTerminate(terminate MessageTerminate) error {
	index, err := p.Chain.Store(&api.BoxNew{terminate.Ballots})
	if err != nil {
		return err
	}

	p.Index = index
	p.Finished <- true

	return nil
}
