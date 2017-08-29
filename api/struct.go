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
		Ballot{}, Box{},
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

// FetchRequest ...
type FetchRequest struct {
	Election string
	Block    uint32
}

// FetchResponse ...
type FetchResponse struct {
	Ballots []Ballot
}
