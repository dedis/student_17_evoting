package api

import (
	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/kyber"
	"github.com/dedis/onet"
	"github.com/dedis/onet/network"

	"github.com/qantik/nevv/chains"
)

func init() {
	network.RegisterMessages(
		Ping{}, Link{}, LinkReply{}, Open{}, OpenReply{}, Cast{}, CastReply{},
		Shuffle{}, ShuffleReply{}, Decrypt{}, DecryptReply{}, GetBox{},
		GetBoxReply{}, GetMixes{}, GetMixesReply{}, GetPartials{},
		GetPartialsReply{},
	)
}

type Ping struct {
	Nonce uint32 // Nonce can be any integer.
}

type Link struct {
	Pin    string       // Pin of the running service.
	Roster *onet.Roster // Roster that handles elections.
	Key    kyber.Point  // Key is a front-end public key.
	Admins []uint32     // Admins is a list of election administrators.
}

type LinkReply struct {
	ID skipchain.SkipBlockID // ID of the master skipchain.
}

type Login struct {
	ID        skipchain.SkipBlockID // ID of the master skipchain.
	User      uint32                // User identifier.
	Signature []byte                // Signature from the front-end.
}

type LoginReply struct {
	Token     string             // Token (time-limited) for further calls.
	Admin     bool               // Admin indicates if user has admin rights.
	Elections []*chains.Election // Elections the user participates in.
}

type Open struct {
	Token    string                // Token for authentication.
	ID       skipchain.SkipBlockID // ID of the master skipchain.
	Election *chains.Election      // Election object.
}

type OpenReply struct {
	ID  skipchain.SkipBlockID // ID of the election skipchain.
	Key kyber.Point           // Key assigned by the DKG.
}

type Cast struct {
	Token  string                // Token for authentication.
	ID     skipchain.SkipBlockID // ID of the election skipchain.
	Ballot *chains.Ballot        // Ballot to be casted.
}

type CastReply struct{}

type Shuffle struct {
	Token string                // Token for authentication.
	ID    skipchain.SkipBlockID // ID of the election skipchain.
}

type ShuffleReply struct{}

type Decrypt struct {
	Token string                // Token for authentication.
	ID    skipchain.SkipBlockID // ID of the election skipchain.
}

type DecryptReply struct{}

type GetBox struct {
	Token string                // Token for authentication.
	ID    skipchain.SkipBlockID // ID of the election skipchain.
}

type GetBoxReply struct {
	Box *chains.Box // Box of encrypted ballots.
}

type GetMixes struct {
	Token string                // Token for authentication.
	ID    skipchain.SkipBlockID // ID of the election skipchain.
}

type GetMixesReply struct {
	Mixes []*chains.Mix // Mixes from all conodes.
}

type GetPartials struct {
	Token string                // Token for authentication.
	ID    skipchain.SkipBlockID // ID of the election skipchain.
}

type GetPartialsReply struct {
	Partials []*chains.Partial // Partials from all conodes.
}
