package net

import (
	"bufio"
	"crypto/cipher"
	"errors"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/dedis/kyber/abstract"
	"github.com/dedis/kyber/ed25519"
	"github.com/dedis/kyber/share/dkg"
	"github.com/dedis/protobuf"

	test "github.com/qantik/mikser/dkg"
)

type mux func(*bufio.ReadWriter, string)

type Server struct {
	index    int
	listener net.Listener
	queue    *queue.Queue
	suite    abstract.Suite
	stream   cipher.Stream
	pool     Roster
	muxes    map[string]mux
	sessions map[string]*dkg.DistKeyGenerator

	beta map[string]test.Session

	comm bool

	constructors protobuf.Constructors
}

func readBlock(io *bufio.ReadWriter) ([]byte, error) {
	line, err := readLine(io)
	if err != nil {
		return nil, err
	}

	n, _ := strconv.Atoi(line)
	bytes, err := readBytes(io, n)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (server *Server) generate(io *bufio.ReadWriter, remote string) {
	id, err := readLine(io)
	if err != nil {
		return
	}

	secret := server.pool[server.index].secret
	generator, err := dkg.NewDistKeyGenerator(server.suite, secret,
		server.pool.publicKeys(), server.stream, 3)
	if err != nil {
		log.Println(err)
		return
	}

	server.sessions[id] = generator
	deals, err := server.sessions[id].Deals()
	if err != nil {
		log.Println(err)
		return
	}

	for index, deal := range deals {
		message := Message{index, DEAL, id, *deal}
		if server.pool.Send(message) != nil {
			_ = server.queue.Put(message)
		}
	}
}

func (server *Server) deal(io *bufio.ReadWriter, remote string) {
	session, err := readLine(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	if _, found := server.sessions[session]; !found {
		//log.Println("Session", session, "not found")
		sendShort(io, ERROR, remote)
		return
	}

	bytes, err := readBlock(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	deal := dkg.Deal{}
	if protobuf.DecodeWithConstructors(bytes, &deal, server.constructors) != nil {
		sendShort(io, ERROR, remote)
		return
	}

	response, err := server.sessions[session].ProcessDeal(&deal)
	if err != nil {
		//log.Println(err)
		sendShort(io, ERROR, remote)
		return
	}

	message := Message{-1, RESPONSE, session, *response}
	server.pool.Broadcast(message, server.index, server.queue)

	sendShort(io, ACK, remote)
}

func (server *Server) response(io *bufio.ReadWriter, remote string) {
	session, err := readLine(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	if _, found := server.sessions[session]; !found {
		sendShort(io, ERROR, remote)
		return
	}

	bytes, err := readBlock(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	response := dkg.Response{}
	if protobuf.DecodeWithConstructors(bytes, &response, server.constructors) != nil {
		sendShort(io, ERROR, remote)
		return
	}

	_, err = server.sessions[session].ProcessResponse(&response)
	if err != nil {
		//log.Println(err)
		sendShort(io, ERROR, remote)
		return
	}

	sendShort(io, ACK, remote)
}

func (server *Server) cert(io *bufio.ReadWriter, remote string) {
	session, err := readLine(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	if _, found := server.sessions[session]; !found {
		sendShort(io, ERROR, remote)
		return
	}

	log.Println("CERTIFIED", server.sessions[session].Certified())
	log.Println("QUAL", server.sessions[session].QUAL())

	commits, err := server.sessions[session].SecretCommits()
	if err != nil {
		log.Println("Certification not ready")
		sendShort(io, ERROR, remote)
		return
	}

	server.comm = true
	qual := server.sessions[session].QUAL()
	message := Message{-1, COMMIT, session, *commits}
	server.pool.BroadcastTo(message, server.index, qual, server.queue)
}

func (server *Server) commit(io *bufio.ReadWriter, remote string) {
	if !server.comm {
		sendShort(io, ERROR, remote)
		return
	}
	session, err := readLine(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	if _, found := server.sessions[session]; !found {
		sendShort(io, ERROR, remote)
		return
	}

	bytes, err := readBlock(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	commits := dkg.SecretCommits{}
	if protobuf.DecodeWithConstructors(bytes, &commits, server.constructors) != nil {
		sendShort(io, ERROR, remote)
		return
	}

	complaint, err := server.sessions[session].ProcessSecretCommits(&commits)
	log.Println(complaint, err)

	sendShort(io, ACK, remote)
}

func (server *Server) key(io *bufio.ReadWriter, remote string) {
	session, err := readLine(io)
	if err != nil {
		sendShort(io, ERROR, remote)
		return
	}

	if _, found := server.sessions[session]; !found {
		sendShort(io, ERROR, remote)
		return
	}

	k, err := server.sessions[session].DistKeyShare()
	if err != nil {
		log.Println(err)
		sendShort(io, ERROR, remote)
		return
	}

	sendShort(io, k.Public().String(), remote)
}

func (server *Server) dispatch(connection net.Conn) {
	io := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))
	defer shut(connection)

	kind, err := readLine(io)
	if err != nil {
		return
	}

	log.Println("Received", kind, "from", connection.RemoteAddr())

	mux, _ := server.muxes[kind]
	mux(io, connection.RemoteAddr().String())
}

// Resend head of message queue, enqueue again if transmission failed.
func (server *Server) reduce() {
	if server.queue.Len() > 0 {
		messages, _ := server.queue.Get(1)
		head := messages[0].(Message)
		if server.pool.Send(head) != nil {
			_ = server.queue.Put(head)
		}
	}
}

// Open starts the listener loop of the TCP server.
func (server *Server) Open() {
	for {
		go server.reduce()
		connection, err := server.listener.Accept()
		if err != nil {
			log.Println("Incoming connection rejected")
			continue
		}

		log.Println(connection.RemoteAddr(), " has connected")
		go server.dispatch(connection)
	}
}

// NewServer creates and returns a new TCP server instance or an error
// if the requirements are not met.
func NewServer(host, file string) (*Server, error) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
	log.SetPrefix(host + " > ")

	server := new(Server)
	server.index = -1

	listener, err := net.Listen("tcp", host)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	server.listener = listener
	server.suite = ed25519.NewAES128SHA256Ed25519(false)
	server.stream = server.suite.Cipher(abstract.RandomKey)
	server.sessions = make(map[string]*dkg.DistKeyGenerator)
	server.queue = queue.New(128)

	if server.pool, err = NewRoster(file, server.suite); err != nil {
		return nil, err
	}

	for index, entity := range server.pool {
		if entity.address == host {
			server.index = index
		}
	}
	if server.index == -1 {
		log.Println("Host not found in pool")
		return nil, errors.New("*")
	}

	server.muxes = make(map[string]mux)
	server.muxes[GENERATE] = server.generate
	server.muxes[DEAL] = server.deal
	server.muxes[RESPONSE] = server.response
	server.muxes[CERT] = server.cert
	server.muxes[COMMIT] = server.commit
	server.muxes[KEY] = server.key
	server.comm = false

	server.constructors = make(protobuf.Constructors)
	var public abstract.Point
	var secret abstract.Scalar
	server.constructors[reflect.TypeOf(&public).Elem()] =
		func() interface{} { return server.suite.Point() }
	server.constructors[reflect.TypeOf(&secret).Elem()] =
		func() interface{} { return server.suite.Scalar() }

	return server, nil
}
