package chains

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(Election{})
	network.RegisterMessage(Ballot{})
	network.RegisterMessage(Box{})
}

const (
	BALLOTS = iota
	SHUFFLE
	DECRYPTION
)

// User is the unique (injective) identifier for a voter. It
// corresponds to EPFL's Tequila Sciper six digit number.
type User uint32

// Ballot represents a vote and is created by the frontend when a
// user casts his decision.
type Ballot struct {
	// User identifier.
	User User `protobuf:"1,req,user"`

	// Alpha is the first element in the ElGamal ciphertext.
	Alpha abstract.Point `protobuf:"2,req,alpha"`
	// Beta is the second element in the ElGamal ciphertext.
	Beta abstract.Point `protobuf:"3,req,beta"`

	// Text is created upon decryption of the above ciphertext.
	Text []byte `protobuf:"4,opt,text"`
}

// Box wraps a list of ballots. This is mainly for storage on the Skipchain
// purposes since pure lists cannot be marshalled.
type Box struct {
	// Ballots is a list of (encrypted, shuffled, decrypted) ballots.
	Ballots []*Ballot
}

// Election is the base object for a voting procedure. It is stored
// in the second SkipBlock right after the (empty) genesis block. A reference
// to the election Skipchain is appended to the master Skipchain upon opening.
type Election struct {
	// Name is a string identifier.
	Name string `protobuf:"1,req,name"`
	// Creator is the user who opened the election.
	Creator User `protobuf:"2,req,creator"`
	// Users is a list of voters who are allowed to participate.
	Users []User `protobuf:"3,rep,users"`

	// Roster is the list of conodes responsible for the election.
	Roster *onet.Roster `protobuf:"4,opt,roster"`
	// Key is the public key from the DKG protocol.
	Key abstract.Point `protobuf:"5,opt,key"`
	// Data can hold any marshallable object (e.g. questions).
	Data []byte `protobuf:"6,opt,data"`

	// Description details further information about the election.
	Description string `protobuf:"7,opt,description"`
	// End date of the election.
	End string `protobuf:"8,opt,end"`
}

// IsUser checks if a given user is a registered voter for the election.
func (e *Election) IsUser(user User) bool {
	for _, u := range e.Users {
		if u == user {
			return true
		}
	}
	return false
}

// IsUser checks if a given user is the creator of the election.
func (e *Election) IsCreator(user User) bool {
	return user == e.Creator
}
