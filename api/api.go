package api

import (
	"strconv"

	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/kyber"
	"github.com/dedis/kyber/sign/schnorr"
	"github.com/dedis/onet"
	"github.com/dedis/onet/network"

	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/crypto"
)

func init() {
	network.RegisterMessages(
		Link{}, LinkReply{}, Open{}, OpenReply{}, Cast{}, CastReply{},
		Shuffle{}, ShuffleReply{}, Decrypt{}, DecryptReply{}, GetBox{},
		GetBoxReply{}, GetMixes{}, GetMixesReply{}, GetPartials{},
		GetPartialsReply{}, Ping{},
	)
}

type Login struct {
	ID        skipchain.SkipBlockID // ID of the master skipchain.
	User      uint32                // User identifier.
	Signature []byte                // Signature from the front-end.
}

// Digest appends the digits of the user identifier to the skipblock ID.
func (l *Login) Digest() []byte {
	message := l.ID
	for _, c := range strconv.Itoa(int(l.User)) {
		d, _ := strconv.Atoi(string(c))
		message = append(message, byte(d))
	}
	return message
}

// Sign creates a Schnorr signature of the login digest.
func (l *Login) Sign(secret kyber.Scalar) error {
	sig, err := schnorr.Sign(crypto.Suite, secret, l.Digest())
	l.Signature = sig
	return err
}

// Verify checks the Schnorr signature.
func (l *Login) Verify(public kyber.Point) error {
	return schnorr.Verify(crypto.Suite, public, l.Digest(), l.Signature)
}

type LoginReply struct {
	Token     string             // Token (time-limited) for further calls.
	Admin     bool               // Admin indicates if user has admin rights.
	Elections []*chains.Election // Elections the user participates in.
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

type Ping struct {
	Nonce uint32 // Nonce can be any integer.
}
