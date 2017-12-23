package dkg

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/share/dkg"

	"github.com/qantik/nevv/crypto"
)

// Simulate executes the distributed key generation protocol offline for a
// given number of nodes. It returns a list of key generators out of which
// the shared public key and the private secrets can be extracted. One key
// generator for per node.
func Simulate(nodes, threshold int) []*dkg.DistKeyGenerator {
	dkgs := make([]*dkg.DistKeyGenerator, nodes)
	scalars := make([]abstract.Scalar, nodes)
	points := make([]abstract.Point, nodes)

	// Initialisation
	for i := range scalars {
		scalars[i] = crypto.Suite.Scalar().Pick(crypto.Stream)
		points[i] = crypto.Suite.Point().Mul(nil, scalars[i])
	}

	// Key-sharing
	for i := range dkgs {
		dkgs[i], _ = dkg.NewDistKeyGenerator(crypto.Suite, scalars[i], points,
			crypto.Stream, threshold)
	}
	// Exchange of Deals
	responses := make([][]*dkg.Response, nodes)
	for i, p := range dkgs {
		responses[i] = make([]*dkg.Response, nodes)
		deals, _ := p.Deals()
		for j, d := range deals {
			responses[i][j], _ = dkgs[j].ProcessDeal(d)
		}
	}
	// Process responses
	for _, resp := range responses {
		for j, r := range resp {
			for k, p := range dkgs {
				if r != nil && j != k {
					p.ProcessResponse(r)
				}
			}
		}
	}

	// Secret commits
	for _, p := range dkgs {
		commit, _ := p.SecretCommits()
		for _, p2 := range dkgs {
			p2.ProcessSecretCommits(commit)
		}
	}

	return dkgs
}
