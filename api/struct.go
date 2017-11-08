package api

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(Ping{})
	network.RegisterMessage(GenerateElection{})
	network.RegisterMessage(GenerateElectionResponse{})
	network.RegisterMessage(GetElections{})
	network.RegisterMessage(GetElectionsReply{})
	network.RegisterMessage(GetBallots{})
	network.RegisterMessage(GetBallotsResponse{})
	network.RegisterMessage(CastBallot{})
	network.RegisterMessage(CastBallotResponse{})
	network.RegisterMessage(Shuffle{})
	network.RegisterMessage(ShuffleReply{})
	network.RegisterMessage(GetShuffle{})
	network.RegisterMessage(GetShuffleReply{})
	network.RegisterMessage(Decrypt{})
	network.RegisterMessage(DecryptReply{})
}

// Ping is the network probing message to check whether the service
// is contactable and running. In the successful case, another ping
// message is returned to the client.
type Ping struct {
	// Nonce is a random integer chosen by the client.
	Nonce uint32 `protobuf:"1,req,nonce"`
}

type GenerateElection struct {
	Token    string   `protobuf:"1,req,token"`
	Election Election `protobuf:"2,req,election"`
}

type GenerateElectionResponse struct {
	Key abstract.Point `protobuf:"1,req,key"`
}

type GetElections struct {
	Token string `protobuf:"1,req,token"`
	User  string `protobuf:"2,req,user"`
}

type GetElectionsReply struct {
	Elections []*Election `protobuf:"1,rep,elections"`
}

type CastBallot struct {
	Token string `protobuf:"1,req,token"`

	ID     string `protobuf:"2,req,id"`
	Ballot Ballot `protobuf:"2,req,ballot"`
}

type CastBallotResponse struct {
	Block uint32 `protobuf:"1,req,block"`
}

type GetBallots struct {
	Token string `protobuf:"1,req,token"`

	ID string `protobuf:"2,req,id"`
}

// TODO: Change ballot list to box.
type GetBallotsResponse struct {
	Ballots []*Ballot `protobuf:"1,req,ballots"`
}

type Shuffle struct {
	Token string `protobuf:"1,req,token"`

	ID string `protobuf:"2,req,id"`
}

type ShuffleReply struct {
	Block uint32 `protobuf:"1,req,block"`
}

type GetShuffle struct {
	Token string `protobuf:"1,req,token"`

	ID string `protobuf:"2,req,id"`
}

type GetShuffleReply struct {
	Box *Box `protobuf:"1,req,box"`
}

type Decrypt struct {
	Token string `protobuf:"1,req,token"`

	ID string `protobuf:"2:req, id"`
}

type DecryptReply struct {
	Block uint32 `protobuf:"1,req,block"`
}
