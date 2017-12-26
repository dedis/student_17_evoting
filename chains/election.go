package chains

import (
	"errors"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	network.RegisterMessage(Election{})
	network.RegisterMessage(Ballot{})
	network.RegisterMessage(Box{})
	network.RegisterMessage(Text{})
}

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
}

// Text holds the decrypted plaintext of a user's ballot.
type Text struct {
	// User identifier.
	User User `protobuf:"1,req,user"`
	// Data is the extracted data from ciphertext.
	Data []byte `protobuf:"2,opt,data"`
}

// Box wraps a list of ballots or texts. This is mainly for storage on the Skipchain
// purposes since pure lists cannot be marshalled.
type Box struct {
	// Ballots is a list of (encrypted, shuffled, decrypted) ballots.
	Ballots []*Ballot `protobuf:"1,opt,ballots"`
	// Texts is a list of decrypted plaintexts from ballots.
	Texts []*Text `protobuf:"2,opt,texts"`
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

	// ID is the election's Skipchain identifier.
	ID skipchain.SkipBlockID `protobuf:"4,opt,id"`
	// Roster is the list of conodes responsible for the election.
	Roster *onet.Roster `protobuf:"5,opt,roster"`
	// Key is the public key from the DKG protocol.
	Key abstract.Point `protobuf:"6,opt,key"`
	// Data can hold any marshallable object (e.g. questions).
	Data []byte `protobuf:"7,opt,data"`
	// Stage indicates the phase of an election. 0 running, 1 shuffled, 2 decrypted.
	Stage uint32 `protobuf:"8,opt,stage"`

	// Description details further information about the election.
	Description string `protobuf:"9,opt,description"`
	// End date of the election.
	End string `protobuf:"10,opt,end"`
}

func FetchElection(roster *onet.Roster, id skipchain.SkipBlockID) (*Election, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	// By definition the master object is stored right after the genesis Skipblock.
	_, blob, _ := network.Unmarshal(chain[1].Data)
	election := blob.(*Election)

	// Set stage.
	for i := 2; i < len(chain); i++ {
		_, blob, _ := network.Unmarshal(chain[i].Data)
		if _, ok := blob.(*Box); ok {
			election.Stage++
		}
	}
	return election, nil
}

func (e *Election) Ballots() (*Box, error) {
	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	// Use map to only included a user's last ballot.
	mapping := make(map[User]*Ballot)
	for i := 2; i < len(chain); i++ {
		_, blob, _ := network.Unmarshal(chain[i].Data)
		ballot, ok := blob.(*Ballot)
		if !ok {
			break
		}
		mapping[ballot.User] = ballot
	}

	ballots := make([]*Ballot, 0)
	for _, ballot := range mapping {
		ballots = append(ballots, ballot)
	}

	return &Box{Ballots: ballots}, nil
}

func (e *Election) Shuffle() (*Box, error) {
	if e.Stage < 1 {
		return nil, errors.New("Election not shuffled yet")
	}

	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	var blob network.Message
	if e.Stage == 1 {
		_, blob, _ = network.Unmarshal(chain[len(chain)-1].Data)
	} else {
		_, blob, _ = network.Unmarshal(chain[len(chain)-2].Data)
	}
	return blob.(*Box), nil
}

func (e *Election) Decryption() (*Box, error) {
	if e.Stage < 2 {
		return nil, errors.New("Election not decrypted yet")
	}

	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	_, blob, _ := network.Unmarshal(chain[len(chain)-1].Data)
	return blob.(*Box), nil
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
