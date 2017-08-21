package api

import (
	"github.com/dedis/cothority/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

// We need to register all messages so the network knows how to handle them.
func init() {
	for _, msg := range []interface{}{
		CountRequest{}, CountResponse{},
		ClockRequest{}, ClockResponse{},
		GenerateRequest{}, GenerateResponse{},
		CastRequest{}, CastResponse{},
	} {
		network.RegisterMessage(msg)
	}
}

// ClockRequest will run the tepmlate-protocol on the roster and return
// the time spent doing so.
type ClockRequest struct {
	Roster *onet.Roster
}

// ClockResponse returns the time spent for the protocol-run.
type ClockResponse struct {
	Time     float64
	Children int
}

// CountRequest will return how many times the protocol has been run.
type CountRequest struct {
}

// CountResponse returns the number of protocol-runs
type CountResponse struct {
	Count int
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

// BallotRequest ...
type CastRequest struct {
	Name   string
	Ballot string
}

// BallotResponse ...
type CastResponse struct {
}
