package app

const (
	// Short messages to terminate transmission
	ack  = "ack"
	fail = "fail"

	// Messages from outside the mixnet
	MsgStartDkg      = "start_dkg"
	MsgStartDeal     = "start_deal"
	MsgStartResponse = "start_response"
	MsgStartCommit   = "start_commit"
	MsgSharedKey     = "shared_key"

	// Messages between nodes of the mixnet
	MsgDeal          = "deal"
	MsgResponse      = "reponse"
	MsgJustification = "justification"
	MsgCommit        = "commit"
)

// Message is the basic mean of communication from and to a mixnet node.
// Since the application serves raw TCP data this structure has to be encoded
// through the golang gob encoder https://golang.org/pkg/encoding/gob.
// TODO: Add sender address field.
type Message struct {
	// Kind depicts the message type, must be one of the exported types.
	Kind string

	// Name of the session for which the message is meant.
	Session string

	// Number of bytes of the encoding field. TODO: Remove.
	Size int

	// Message body. Format is specified by the message kind.
	Encoding []byte
}
