package net

import (
	"bufio"
	"log"
	"net"
	"strconv"

	"github.com/dedis/kyber/share/dkg"
	"github.com/dedis/protobuf"
)

const (
	ACK   = "ack"
	ERROR = "error"

	GENERATE = "generate"
	DEAL     = "deal"
	RESPONSE = "response"
	CERT     = "cert"
	COMMIT   = "commit"
	KEY      = "key"
)

type Message struct {
	index   int
	kind    string
	session string
	body    interface{}
}

func (message Message) pack(connection net.Conn) (*bufio.ReadWriter, error) {
	var encoding []byte
	var err error

	switch message.kind {
	case DEAL:
		body := message.body.(dkg.Deal)
		encoding, err = protobuf.Encode(&body)
		break
	case RESPONSE:
		body := message.body.(dkg.Response)
		encoding, err = protobuf.Encode(&body)
		break
	case COMMIT:
		body := message.body.(dkg.SecretCommits)
		encoding, err = protobuf.Encode(&body)
		break
	default:
		panic("Unknown message kind")
	}

	if err != nil {
		log.Println("Could not encode", message.kind)
		log.Println(err)
		return nil, err
	}

	io := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))
	_, _ = io.WriteString(message.kind + "\n")
	_, _ = io.WriteString(message.session + "\n")
	_, _ = io.WriteString(strconv.FormatInt(int64(len(encoding)), 10) + "\n")
	_, _ = io.Write(encoding)

	return io, nil
}
