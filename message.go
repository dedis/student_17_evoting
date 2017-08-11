package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
)

const (
	// Short messages to terminate transmission
	ack  = "ack"
	fail = "fail"

	// Messages from outside the mixnet
	msgStartDkg      = "start_dkg"
	msgStartDeal     = "start_deal"
	msgStartResponse = "start_response"
	msgStartCommit   = "start_commit"
	msgSharedKey     = "shared_key"

	// Messages between nodes of the mixnet
	msgDeal          = "deal"
	msgResponse      = "reponse"
	msgJustification = "justification"
	msgCommit        = "commit"
)

type Message struct {
	kind     string
	session  string
	size     int
	encoding []byte
}

func (message Message) pack(conn net.Conn) *bufio.ReadWriter {
	channel := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	_, _ = channel.WriteString(message.kind + "\n")
	_, _ = channel.WriteString(message.session + "\n")
	_, _ = channel.WriteString(strconv.FormatInt(int64(message.size), 10) + "\n")
	_, _ = channel.Write(message.encoding)

	return channel
}

func unpack(channel *bufio.ReadWriter) (*Message, error) {
	session, err := line(channel)
	if err != nil {
		return nil, err
	}

	line, err := line(channel)
	if err != nil {
		return nil, err
	}

	n, _ := strconv.Atoi(line)
	bytes := make([]byte, n)

	for i := 0; i < n; i++ {
		byte, err := channel.ReadByte()
		if err != nil {
			log.Println("Could not read byte from io")
			return nil, err
		}
		bytes[i] = byte
	}

	return &Message{ack, session, len(bytes), bytes}, nil
}
