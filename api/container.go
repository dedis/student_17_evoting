package api

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

type Election struct {
	ID     string       `protobuf:"1,req,id"`
	Admin  string       `protobuf:"2,req,admin"`
	Start  string       `protobuf:"3,req,start"`
	End    string       `protobuf:"4,req,end"`
	Data   []byte       `protobuf:"5,req,data"`
	Roster *onet.Roster `protobuf:"6,req,roster"`

	Users []string `protobuf:"7,rep,users"`

	Key         abstract.Point `protobuf:"8,opt,key"`
	Description string         `protobuf:"9,opt,description"`
}

// Ballot consists of an ElGamal key pair that is created by the frontend
// and and stored in an individual block on the SkipChain.
type Ballot struct {
	Alpha abstract.Point
	Beta  abstract.Point
}

// Box wraps a list of Ballots for easier bulk storage on the SkipChain after
// for example a shuffle procedure.
type Box struct {
	Ballots []Ballot
}

// Join is a helper method for joining two list of ElGamal pairs to a single
// list of ballots.
func (box *Box) Join(alpha []abstract.Point, beta []abstract.Point) {
	box.Ballots = make([]Ballot, len(alpha))
	for index := 0; index < len(alpha); index++ {
		box.Ballots[index] = Ballot{alpha[index], beta[index]}
	}
}

// Split is a helper method to create two separate lists of ElGamal pairs from
// a single list.
func (box Box) Split() (alpha, beta []abstract.Point) {
	length := len(box.Ballots)
	alpha = make([]abstract.Point, length)
	beta = make([]abstract.Point, length)

	for index, ballot := range box.Ballots {
		alpha[index] = ballot.Alpha
		beta[index] = ballot.Beta
	}

	return
}

type BallotNew struct {
	User string `protobuf:"1,req,user"`

	Alpha abstract.Point `protobuf:"2,req,alpha"`
	Beta  abstract.Point `protobuf:"3,req,beta"`

	Clear []byte `protobuf:"4,opt,clear"`
}

type BoxNew struct {
	Ballots []*BallotNew
}
