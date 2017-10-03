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
		GenerateElection{}, GenerateElectionResponse{},
		CastBallot{}, CastBallotResponse{},
		Election{}, Ballot{}, BallotNew{}, Box{},
	} {
		network.RegisterMessage(msg)
	}
}

// GenerateRequest initiates the creation of a new election and the
// corresponding SkipChain given a name and a roster of conodes.
type GenerateRequest struct {
	Name   string
	Roster *onet.Roster
}

// GenerateResponse is returned to the frontend after a successful creation
// of an election. It contains the public key from the distributed key generation
// protocol as well as the hash of the genesis SkipBlock.
type GenerateResponse struct {
	Key  abstract.Point
	Hash skipchain.SkipBlockID
}

// CastRequest prompts the addition of a ballot to an election's SkipChain.
type CastRequest struct {
	Election string
	Ballot   *Ballot
}

// CastResponse is returned to the frontend after the ballot has been
// successfully casted.
type CastResponse struct {
}

// ShuffleRequest initiates the shuffle protocol for a given election.
type ShuffleRequest struct {
	Election string
}

// ShuffleResponse is returned to the frontend when the shuffle procedure has
// been completed.
type ShuffleResponse struct {
}

// FetchRequest requests the ballots for a given election stored in a specific
// block.
type FetchRequest struct {
	Election string
	Block    uint32
}

// FetchResponse with a list of ballots is returned to the frontend when the particular
// block of the election has been found.
type FetchResponse struct {
	Ballots []Ballot
}

type DecryptionRequest struct {
	Election string
	Ballot   *Ballot
}

type DecryptionResponse struct {
}

/////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

type GenerateElection struct {
	Token    string   `protobuf:"1,req,token"`
	Election Election `protobuf:"2,req,election"`
}

type GenerateElectionResponse struct {
	Key abstract.Point `protobuf:"1,req,key"`
}

type CastBallot struct {
	Token string `protobuf:"1,req,token"`

	ID     string    `protobuf:"2,req,id"`
	Ballot BallotNew `protobuf:"2,req,ballot"`
}

type CastBallotResponse struct {
	Block uint32 `protobuf:"1,req,block"`
}
