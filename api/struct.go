package api

import (
	"github.com/dedis/cothority/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	for _, msg := range []interface{}{
		GenerateRequest{}, GenerateResponse{},
		CastRequest{}, CastResponse{},
		ShuffleRequest{}, ShuffleResponse{},
		FetchRequest{}, FetchResponse{},
		Ballot{}, Collection{},
	} {
		network.RegisterMessage(msg)
	}
}

// GenerateRequest ...
type GenerateRequest struct {
	Name   string
	Roster *onet.Roster
}

// GenerateResponse ...
type GenerateResponse struct {
	Key  abstract.Point
	Hash skipchain.SkipBlockID
}

// Ballot ...
type Ballot struct {
	Alpha abstract.Point
	Beta  abstract.Point
}

type Collection struct {
	Ballots []Ballot
}

func (collection *Collection) Join(alpha []abstract.Point, beta []abstract.Point) {
	collection.Ballots = make([]Ballot, len(alpha))
	for index := 0; index < len(alpha); index++ {
		collection.Ballots[index] = Ballot{alpha[index], beta[index]}
	}
}

func (collection *Collection) Split() ([]abstract.Point, []abstract.Point) {
	alpha := make([]abstract.Point, len(collection.Ballots))
	beta := make([]abstract.Point, len(collection.Ballots))
	for index := 0; index < len(collection.Ballots); index++ {
		alpha[index] = collection.Ballots[index].Alpha
		beta[index] = collection.Ballots[index].Beta
	}

	return alpha, beta
}

// CastRequest ...
type CastRequest struct {
	Election string
	Ballot   *Ballot
}

// CastResponse ...
type CastResponse struct {
}

// ShuffleRequest ...
type ShuffleRequest struct {
	Election string
}

// ShuffleResponse ...
type ShuffleResponse struct {
}

type FetchRequest struct {
	Block uint32
}

type FetchResponse struct {
	Collection *Collection
}
