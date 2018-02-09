package dkg

import (
	"errors"

	"github.com/dedis/kyber"
	dkg "github.com/dedis/kyber/share/dkg/rabin"

	"github.com/qantik/nevv/crypto"
)

func Simulate(nbrNodes, threshold int) (dkgs []*dkg.DistKeyGenerator, err error) {
	// 1 - share generation
	dkgs = make([]*dkg.DistKeyGenerator, nbrNodes)
	scalars := make([]kyber.Scalar, nbrNodes)
	points := make([]kyber.Point, nbrNodes)
	// 1a - initialisation
	for i := range scalars {
		scalars[i] = crypto.Suite.Scalar().Pick(crypto.Suite.RandomStream())
		points[i] = crypto.Suite.Point().Mul(scalars[i], nil)
	}

	// 1b - key-sharing
	for i := range dkgs {
		dkgs[i], err = dkg.NewDistKeyGenerator(crypto.Suite,
			scalars[i], points, threshold)
		if err != nil {
			return
		}
	}
	// Exchange of Deals
	responses := make([][]*dkg.Response, nbrNodes)
	for i, p := range dkgs {
		responses[i] = make([]*dkg.Response, nbrNodes)
		deals, err := p.Deals()
		if err != nil {
			return nil, err
		}
		for j, d := range deals {
			responses[i][j], err = dkgs[j].ProcessDeal(d)
			if err != nil {
				return nil, err
			}
		}
	}
	// ProcessResponses
	for _, resp := range responses {
		for j, r := range resp {
			for k, p := range dkgs {
				if r != nil && j != k {
					justification, err := p.ProcessResponse(r)
					if err != nil {
						return nil, err
					}
					if justification != nil {
						return nil, errors.New("there should be no justification")
					}
				}
			}
		}
	}

	// Secret commits
	for _, p := range dkgs {
		commit, err := p.SecretCommits()
		if err != nil {
			return nil, err
		}
		for _, p2 := range dkgs {
			compl, err := p2.ProcessSecretCommits(commit)
			if err != nil {
				return nil, err
			}
			if compl != nil {
				return nil, errors.New("there should be no complaint")
			}
		}
	}

	// Verify if all is OK
	for _, p := range dkgs {
		if !p.Finished() {
			return nil, errors.New("one of the dkgs is not finished yet")
		}
	}
	return
}
