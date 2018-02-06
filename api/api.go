package api

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/chains"
)

func init() {
	network.RegisterMessage(Ping{})
	network.RegisterMessage(Link{})
	network.RegisterMessage(LinkReply{})
	network.RegisterMessage(Open{})
	network.RegisterMessage(OpenReply{})
	network.RegisterMessage(Cast{})
	network.RegisterMessage(CastReply{})
	network.RegisterMessage(Shuffle{})
	network.RegisterMessage(ShuffleReply{})
	network.RegisterMessage(Decrypt{})
	network.RegisterMessage(DecryptReply{})
	network.RegisterMessage(Aggregate{})
	network.RegisterMessage(AggregateReply{})
}

// Ping is the network probing message to check whether the service
// is contactable and running. In the successful case, another ping
// message is returned to the client.
type Ping struct {
	// Nonce is a random integer chosen by the client.
	Nonce uint32 `protobuf:"1,req,nonce"`
}

// Link is sent to the service whenever a new master skipchain has to be
// created. The message is only processed if the pin is corresponds to the
// pin of the running service. Since it is not part of the official API, its
// origin should be nevv's command line tool.
type Link struct {
	// Pin identifier of the service.
	Pin string `protobuf:"1,req,pin"`
	// Roster specifies the nodes handling the master skipchain.
	Roster *onet.Roster `protobuf:"2,req,roster"`
	// Key is the frontend public key.
	Key abstract.Point `protobuf:"3,req,key"`
	// Admins is a list of responsible admin (sciper numbers) users.
	Admins []chains.User `protobuf:"4,opt,admins"`
}

// LinkReply is returned when a master skipchain has been successfully created.
// It is only supposed to be sent to the command line tool since it is not part
// of the official API.
type LinkReply struct {
	// Master is the id of the genesis block of the master Skipchain.
	Master skipchain.SkipBlockID `protobuf:"1,opt,master"`
}

// Login is sent whenever a user wants to register himself to the service.
// To evaluate his privilege level the master Skipchain ID must be included.
// Furthermore to unambiguously authenticate the user a Tequila signature
// has to be sent as well.
type Login struct {
	// Master is the ID of the master Skipchain.
	Master skipchain.SkipBlockID `protobuf:"1,req,master"`
	// User is a Sciper six digit identifier.
	User chains.User `protobuf:"2,req,sciper"`
	// Signature from the Tequila service.
	Signature []byte `protobuf:"3,req,signature"`
}

// LoginReply marks a successful registration to the service. It contains
// the user's time-limited token as well as all election the user is part of.
type LoginReply struct {
	// Token is for authenticating an already registered user.
	Token string `protobuf:"1,req,token"`
	// Admin indicates if the user has admin priviledges.
	Admin bool `protobuf:"2,req,admin"`
	// Elections contains all elections in which the user participates.
	Elections []*chains.Election `protobuf:"3,rep,elections"`
}

// Open is sent when an administrator creates a new election. The admin
// has to be logged in and be marked as such in the master Skipchain.
type Open struct {
	// Token to check if admin is logged in.
	Token string `protobuf:"1,req,token"`
	// Master is the ID of the master skipchain.
	Master skipchain.SkipBlockID `protobuf:"2,req,master"`
	// Election is the skeleton of the to-be created election.
	Election *chains.Election `protobuf:"3,req,election"`
}

// OpenReply marks the successful creation of a new election. It contains
// the ID of the election Skipchain as well as the public key from the
// distributed key generation protocol.
type OpenReply struct {
	// Genesis is the ID of the election Skipchain.
	Genesis skipchain.SkipBlockID `protobuf:"1,req,genesis"`
	// Key is the election public key from the DKG.
	Key abstract.Point `protobuf:"2,req,key"`
}

// Cast is sent when a user wishes to cast a ballot in an election. Said
// user has to be logged in and part of the election's user list.
type Cast struct {
	// Token to check if user is logged in.
	Token string `protobuf:"1,req,token"`
	// Genesis is the ID of the election Skipchain.
	Genesis skipchain.SkipBlockID `protobuf:"2,req,genesis"`
	// Ballot is the user's actual vote.
	Ballot *chains.Ballot `protobuf:"3,req,ballot"`
}

// CastReply is returned when a ballot has been successfully casted. It
// included the index of the SkipBlock containing the ballot.
type CastReply struct {
	// Index is the index of the Skipblock containing the casted ballot.
	Index uint32 `protobuf:"1,req,index"`
}

// Aggregate is sent to retrieve a box of either encrypted, shuffled or
// decrypted ballots. The sender has to be logged in and has to be either
// the election's creator or part of its user list.
type Aggregate struct {
	// Token to check if the sender is logged
	Token string `protobuf:"1,req,token"`
	// Genesis is the ID of the election Skipchain.
	Genesis skipchain.SkipBlockID `protobuf:"2,req,genesis"`
	// Type of the box {0: Encrypted Ballots, 1: Shuffled, 2: Decryption}.
	Type uint32 `protobuf:"3,req,type"`
}

// AggregateReply contains the requested box of ballots requested
// by the sender.
type AggregateReply struct {
	// Box of either encrypted, shuffled or decrypted ballots.
	Box *chains.Box `protobuf:"1,req,box"`
}

type Shuffle struct {
	Token   string                `protobuf:"1,req,token"`
	Genesis skipchain.SkipBlockID `protobuf:"2,req,token"`
}

type ShuffleReply struct {
	Shuffled *chains.Box `protobuf:"1,req.shuffled"`
}

type Decrypt struct {
	Token   string                `protobuf:"1,req,token"`
	Genesis skipchain.SkipBlockID `protobuf:"2,req,token"`
}

type DecryptReply struct {
	Decrypted *chains.Box `protobuf:"1,req.shuffled"`
}

type GetBox struct {
	Token   string                `protobuf:"1,req,token"`
	Genesis skipchain.SkipBlockID `protobuf:"2,req,genesis"`
}

type GetBoxReply struct {
	Box *chains.Box `protobuf:"1,req,box"`
}

type GetMixes struct {
	Token   string                `protobuf:"1,req,token"`
	Genesis skipchain.SkipBlockID `protobuf:"2,req,genesis"`
}

type GetMixesReply struct {
	Mixes []*chains.Mix `protobuf:"1,req,mixes"`
}

type GetPartials struct {
	Token   string                `protobuf:"1,req,token"`
	Genesis skipchain.SkipBlockID `protobuf:"2,req,genesis"`
}

type GetPartialsReply struct {
	Partials []*chains.Partial `protobuf:"1,req,partials"`
}
