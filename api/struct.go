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
		CastRequest{}, CastResponse{}, Ballot{},
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

// CastRequest ...
type CastRequest struct {
	Election string
	Ballot   *Ballot
}

// CastResponse ...
type CastResponse struct {
}
