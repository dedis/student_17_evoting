package app

import (
	"bufio"
	"crypto/cipher"
	"encoding/gob"
	"log"
	"net"
	"reflect"

	"github.com/dedis/kyber/abstract"
	"github.com/dedis/kyber/ed25519"
	"github.com/dedis/kyber/share/dkg"
	"github.com/dedis/protobuf"
)

// Server is a mixnet node concerned with the distributed key generation,
// and the shuffling of incoming objects. It sends and receives raw TCP messages
// that are multiplexed and taken care of by the corresponding handler function.
type Server struct {
	// Address of the node
	host string

	// Server socket
	listener net.Listener

	// Cryptographical constants
	suite  abstract.Suite
	stream cipher.Stream

	// Multiplexer of handler functions
	muxes map[string]func(*bufio.ReadWriter, Message)

	// Mapping of all open sessions
	sessions map[string]*Session

	// Default DEDIS protobuf constructor
	constructors protobuf.Constructors
}

// CreateServer returns a new server instance for a given host address.
func CreateServer(host string) (*Server, error) {
	server := new(Server)

	listener, err := net.Listen("tcp", host)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	server.host = host
	server.listener = listener
	server.suite = ed25519.NewAES128SHA256Ed25519(false)
	server.stream = server.suite.Cipher(abstract.RandomKey)
	server.sessions = make(map[string]*Session)

	// Create multiplexer and register handler functions.
	server.muxes = make(map[string]func(*bufio.ReadWriter, Message))
	server.muxes[MsgStartDkg] = server.startDkg
	server.muxes[MsgStartDeal] = server.startDeal
	server.muxes[MsgStartResponse] = server.startResponse
	server.muxes[MsgStartCommit] = server.startCommit
	server.muxes[MsgSharedKey] = server.sharedKey
	server.muxes[MsgStartShuffle] = server.startShuffle
	server.muxes[MsgShuffle] = server.shuffle
	server.muxes[MsgDeal] = server.deal
	server.muxes[MsgResponse] = server.response
	server.muxes[MsgCommit] = server.commit

	// See https://github.com/dedis/onet/blob/master/network/encoding.go
	// for a sample initialization of a protobuf constructor.
	server.constructors = make(protobuf.Constructors)
	var public abstract.Point
	var secret abstract.Scalar
	server.constructors[reflect.TypeOf(&public).Elem()] =
		func() interface{} { return server.suite.Point() }
	server.constructors[reflect.TypeOf(&secret).Elem()] =
		func() interface{} { return server.suite.Scalar() }

	return server, nil
}

// Listen starts the listening routine of the server. Accepted connections are
// handled concurrently.
func (server *Server) Listen() {
	for {
		connection, err := server.listener.Accept()
		if err != nil {
			log.Println("Incoming connection rejected")
			continue
		}

		log.Println(connection.RemoteAddr(), " has connected")
		go server.dispatch(connection)
	}
}

// Parse incoming messages and dispatch the corresponing handler function.
func (server *Server) dispatch(conn net.Conn) {
	channel := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer func() { _ = conn.Close() }()

	var message Message
	decoder := gob.NewDecoder(channel)
	if decoder.Decode(&message) != nil {
		log.Println("Could not decode message")
		return
	}

	log.Println("Received", message.Kind, "from", conn.RemoteAddr())

	_, found := server.sessions[message.Session]
	if !found && message.Kind != MsgStartDkg {
		terminate(channel, fail, message.Session, "Session not found")
		return
	} else if found && message.Kind == MsgStartDkg {
		terminate(channel, fail, message.Session, "Session already exists")
		return
	}

	mux, _ := server.muxes[message.Kind]
	mux(channel, message)
}

// End an open connection with a message.
func terminate(channel *bufio.ReadWriter, kind, session, text string) {
	encoding := []byte(text)
	message := Message{kind, session, len(encoding), encoding}

	encoder := gob.NewEncoder(channel)
	if encoder.Encode(message) != nil {
		return
	}
	_ = channel.Flush()

	log.Println("Sent", message.Kind)
}

func term(channel *bufio.ReadWriter, kind, session string, encoding []byte) {
	message := Message{kind, session, len(encoding), encoding}

	encoder := gob.NewEncoder(channel)
	if encoder.Encode(message) != nil {
		return
	}
	_ = channel.Flush()

	log.Println("Sent", message.Kind)
}

// Start distributed key generation handler function.
func (server *Server) startDkg(channel *bufio.ReadWriter, message Message) {
	name := string(message.Session)
	pool := "resources/pool.toml"

	session, err := session(name, server.suite, server.stream, pool, server.host)
	if err != nil {
		terminate(channel, fail, name, err.Error())
		return
	}

	server.sessions[name] = session
	terminate(channel, ack, name, "")
}

// Start deal distribution handler function.
func (server *Server) startDeal(channel *bufio.ReadWriter, message Message) {
	if err := server.sessions[message.Session].startDeal(); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}

// Start response distribution handler function.
func (server *Server) startResponse(channel *bufio.ReadWriter, message Message) {
	if err := server.sessions[message.Session].startResponse(); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}

// Start commit distribution handler function.
func (server *Server) startCommit(channel *bufio.ReadWriter, message Message) {
	if err := server.sessions[message.Session].startCommit(); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}

// Shared key retrieval handler function.
func (server *Server) sharedKey(channel *bufio.ReadWriter, message Message) {
	key, err := server.sessions[message.Session].sharedKey()
	if err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, key.String())
}

// Start shuffle handler function.
// TODO: Replace with storage on SkipChain.
func (server *Server) startShuffle(channel *bufio.ReadWriter, message Message) {
	s := Shuffle{}
	if err := protobuf.DecodeWithConstructors(message.Encoding, &s,
		server.constructors); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	_ = server.sessions[message.Session].startShuffle(&s, server.suite, server.stream)
	terminate(channel, ack, message.Session, "")
}

// Shuffle handler functions. Returns output shuffle pairs to requester.
// TODO: Replace with storage on SkipChain.
func (server *Server) shuffle(channel *bufio.ReadWriter, message Message) {
	shuffle := Shuffle{}
	if err := protobuf.DecodeWithConstructors(message.Encoding, &shuffle,
		server.constructors); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	pairs := server.sessions[message.Session].output

	response := Shuffle{0, nil, pairs}
	encoding, err := protobuf.Encode(&response)
	if err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	term(channel, ack, message.Session, encoding)
}

// Incoming deal handler function.
func (server *Server) deal(channel *bufio.ReadWriter, message Message) {
	deal := dkg.Deal{}
	if err := protobuf.DecodeWithConstructors(message.Encoding, &deal,
		server.constructors); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	if err := server.sessions[message.Session].deal(&deal); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}

// Incoming response handler function.
func (server *Server) response(channel *bufio.ReadWriter, message Message) {
	response := dkg.Response{}
	if err := protobuf.DecodeWithConstructors(message.Encoding, &response,
		server.constructors); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	if err := server.sessions[message.Session].response(&response); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}

// Incoming justification handler function.
func (server *Server) justification(channel *bufio.ReadWriter, message Message) {
	justification := dkg.Justification{}
	if err := protobuf.DecodeWithConstructors(message.Encoding, &justification,
		server.constructors); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	if err := server.sessions[message.Session].justification(&justification); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}

// Incoming commit handler function.
func (server *Server) commit(channel *bufio.ReadWriter, message Message) {
	commits := dkg.SecretCommits{}
	if err := protobuf.DecodeWithConstructors(message.Encoding, &commits,
		server.constructors); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	if err := server.sessions[message.Session].commit(&commits); err != nil {
		terminate(channel, fail, message.Session, err.Error())
		return
	}

	terminate(channel, ack, message.Session, "")
}
