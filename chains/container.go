package chains

import "gopkg.in/dedis/crypto.v0/abstract"

// Ballot represents a vote and is created by the frontend when a
// user casts his decision.
type Ballot struct {
	// User identifier.
	User User `protobuf:"1,req,user"`

	// Alpha is the first element in the ElGamal ciphertext.
	Alpha abstract.Point `protobuf:"2,req,alpha"`
	// Beta is the second element in the ElGamal ciphertext.
	Beta abstract.Point `protobuf:"3,req,beta"`
}

type Box struct {
	Ballots []*Ballot `protobuf:"1,opt,ballots"`
}

type Mix struct {
	Ballots []*Ballot `protobuf:"1,req,ballots"`
	Proof   []byte    `protobuf:"2,req,proof"`

	Node string `protobuf:"3,req,node"`
}

type Partial struct {
	Points []*abstract.Point `protobuf:"1,req,points"`
	Proof  []byte            `protobuf:"2,req,proof"`

	Node string `protobuf:3,req,node`
}
