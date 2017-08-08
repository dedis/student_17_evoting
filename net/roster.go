package net

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/dedis/kyber/abstract"
)

type entity struct {
	address string
	public  abstract.Point
	secret  abstract.Scalar
}

type Roster []entity

func NewRoster(file string, suite abstract.Suite) (Roster, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Servers []struct {
			Address string
			Public  string
			Secret  string
		}
	}

	if _, err = toml.Decode(string(data), &parsed); err != nil {
		return nil, err
	}

	roster := make([]entity, len(parsed.Servers))

	for index, server := range parsed.Servers {
		entity := entity{server.Address, suite.Point(), suite.Scalar()}

		decoded, err := hex.DecodeString(server.Public)
		if err != nil {
			return nil, err
		}
		buffer := bytes.NewBuffer(decoded)
		if _, err = entity.public.UnmarshalFrom(buffer); err != nil {
			return nil, err
		}

		decoded, err = hex.DecodeString(server.Secret)
		if err != nil {
			return nil, err
		}
		buffer = bytes.NewBuffer(decoded)
		if _, err = entity.secret.UnmarshalFrom(buffer); err != nil {
			return nil, err
		}

		roster[index] = entity

	}

	return roster, nil
}

func (roster Roster) publicKeys() []abstract.Point {
	keys := make([]abstract.Point, 0, len(roster))
	for _, key := range roster {
		keys = append(keys, key.public)
	}

	return keys
}

func (roster Roster) Broadcast(message Message, host int, queue *queue.Queue) {
	for index := range roster {
		if index != host {
			message.index = index
			if roster.Send(message) != nil {
				_ = queue.Put(message)
			}
		}
	}
}

func (roster Roster) BroadcastTo(message Message, host int, qual []int, queue *queue.Queue) {
	for _, index := range qual {
		if index != host {
			message.index = index
			if roster.Send(message) != nil {
				_ = queue.Put(message)
			}
		}
	}
}

func (roster Roster) Send(message Message) error {
	address := roster[message.index].address
	connection, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("Could not establish connection to", address)
		return err
	}
	defer shut(connection)

	io, err := message.pack(connection)
	if err != nil {
		return err
	}

	if err = io.Flush(); err != nil {
		log.Println("Could not flush channel to", address)
		return err
	}

	log.Println("Sent", message.kind, "to", address)

	response, err := readLine(io)
	if response != ACK || err != nil {
		log.Println("No ack for", message.kind, "from", address)
		return errors.New(ERROR)
	}

	log.Println("Ack for", message.kind, "from", address)

	return nil
}
