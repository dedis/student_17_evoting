package main

import (
	"bufio"
	"crypto/cipher"
	"errors"
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
	muxes map[string]func(*bufio.ReadWriter, string)

	// Mapping of all open sessions
	sessions map[string]*Session

	// Default DEDIS protobuf constructor
	constructors protobuf.Constructors
}

// Create a new server instance for a given host address.
func server(host string) (*Server, error) {
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
	server.muxes = make(map[string]func(*bufio.ReadWriter, string))
	server.muxes[msgStartDkg] = server.startDkg
	server.muxes[msgStartDeal] = server.startDeal
	server.muxes[msgStartResponse] = server.startResponse
	server.muxes[msgStartCommit] = server.startCommit
	server.muxes[msgSharedKey] = server.sharedKey
	server.muxes[msgDeal] = server.deal
	server.muxes[msgResponse] = server.response
	server.muxes[msgCommit] = server.commit

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

// Start the listening routine of the server. Accepted connections are
// handled concurrently.
func (server *Server) open() {
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

// Dispatch reads the first line of an incoming message and calls the corresponing
// handler function.
// TODO: Unify dispatcher with message struct such that only whole messages are sent.
func (server *Server) dispatch(conn net.Conn) {
	channel := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer func() { _ = conn.Close() }()

	kind, err := line(channel)
	if err != nil {
		return
	}

	log.Println("Received", kind, "from", conn.RemoteAddr())

	mux, _ := server.muxes[kind]
	mux(channel, conn.RemoteAddr().String())
}

// Terminate closes an exisiting TCP connection with a short message.
// TODO: Unify with message struct.
func terminate(channel *bufio.ReadWriter, message, kind, address string) {
	if _, err := channel.WriteString(message + "\n"); err != nil {
		log.Println("Could not write", message)
		return
	}

	if channel.Flush() != nil {
		log.Println("Could not flush", message)
		return
	}

	log.Println("Sent", message, "for", kind, "to", address)
}

// Lookup extracts the session from a message checks for its existance
// TODO: Remove this function after messsage unification.
func (server *Server) lookup(channel *bufio.ReadWriter) (*Session, error) {
	name, err := line(channel)
	if err != nil {
		return nil, err
	}

	session, found := server.sessions[name]
	if !found {
		return nil, errors.New("Session " + name + " not found")
	}

	return session, nil
}

// Start distributed key generation handler function.
func (server *Server) startDkg(channel *bufio.ReadWriter, address string) {
	name, err := line(channel)
	if err != nil {
		terminate(channel, fail, msgStartDkg, address)
		return
	}

	session, err := session(name, server.suite, server.stream, "pool.toml", server.host)
	if err != nil {
		terminate(channel, fail, msgStartDkg, address)
		return
	}

	server.sessions[name] = session

	terminate(channel, ack, msgStartDkg, address)
}

// Start deal distribution handler function.
func (server *Server) startDeal(channel *bufio.ReadWriter, address string) {
	session, err := server.lookup(channel)
	if err != nil {
		terminate(channel, fail, msgStartDeal, address)
		return
	}

	if err = session.startDeal(); err != nil {
		terminate(channel, fail, msgStartDeal, address)
		return
	}

	terminate(channel, ack, msgStartDeal, address)
}

// Start response distribution handler function.
func (server *Server) startResponse(channel *bufio.ReadWriter, address string) {
	session, err := server.lookup(channel)
	if err != nil {
		terminate(channel, fail, msgStartResponse, address)
		return
	}

	if session.startResponse() != nil {
		terminate(channel, fail, msgStartResponse, address)
		return
	}

	terminate(channel, ack, msgStartResponse, address)
}

// Start commit distribution handler function.
func (server *Server) startCommit(channel *bufio.ReadWriter, address string) {
	session, err := server.lookup(channel)
	if err != nil {
		terminate(channel, fail, msgStartCommit, address)
		return
	}

	if session.startCommit() != nil {
		terminate(channel, fail, msgStartCommit, address)
		return
	}

	terminate(channel, ack, msgStartCommit, address)
}

// Shared key retrieval handler function.
func (server *Server) sharedKey(channel *bufio.ReadWriter, address string) {
	session, err := server.lookup(channel)
	if err != nil {
		terminate(channel, fail, msgSharedKey, address)
		return
	}

	key, err := session.sharedKey()
	if err != nil {
		terminate(channel, fail, msgSharedKey, address)
		return
	}

	terminate(channel, key.String(), msgSharedKey, address)
}

// Incoming deal handler function.
func (server *Server) deal(channel *bufio.ReadWriter, address string) {
	message, err := unpack(channel)
	if err != nil {
		terminate(channel, fail, msgDeal, address)
		return
	}

	deal := dkg.Deal{}
	if protobuf.DecodeWithConstructors(message.encoding, &deal,
		server.constructors) != nil {
		terminate(channel, fail, msgDeal, address)
		return
	}

	if server.sessions[message.session].deal(&deal) != nil {
		terminate(channel, fail, msgDeal, address)
		return
	}

	terminate(channel, ack, msgDeal, address)
}

// Incoming response handler function.
func (server *Server) response(channel *bufio.ReadWriter, address string) {
	message, err := unpack(channel)
	if err != nil {
		terminate(channel, fail, msgResponse, address)
		return
	}

	response := dkg.Response{}
	if protobuf.DecodeWithConstructors(message.encoding, &response,
		server.constructors) != nil {
		terminate(channel, fail, msgResponse, address)
		return
	}

	if server.sessions[message.session].response(&response) != nil {
		terminate(channel, fail, msgResponse, address)
		return
	}

	terminate(channel, ack, msgResponse, address)
}

// Incoming justification handler function.
func (server *Server) justification(channel *bufio.ReadWriter, address string) {
	message, err := unpack(channel)
	if err != nil {
		terminate(channel, fail, msgJustification, address)
		return
	}

	justification := dkg.Justification{}
	if protobuf.DecodeWithConstructors(message.encoding, &justification,
		server.constructors) != nil {
		terminate(channel, fail, msgJustification, address)
		return
	}

	if server.sessions[message.session].justification(&justification) != nil {
		terminate(channel, fail, msgJustification, address)
		return
	}

	terminate(channel, ack, msgJustification, address)
}

// Incoming commit handler function.
func (server *Server) commit(channel *bufio.ReadWriter, address string) {
	message, err := unpack(channel)
	if err != nil {
		terminate(channel, fail, msgCommit, address)
		return
	}

	commits := dkg.SecretCommits{}
	if protobuf.DecodeWithConstructors(message.encoding, &commits,
		server.constructors) != nil {
		terminate(channel, fail, msgCommit, address)
		return
	}

	if server.sessions[message.session].commit(&commits) != nil {
		terminate(channel, fail, msgCommit, address)
		return
	}

	terminate(channel, ack, msgCommit, address)
}
