package app

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/dedis/kyber/abstract"
)

// Entity is the current depiction of a node in the mixnet. This is
// still very experimental until a more stable way of parsing a roster
// has been found.
type Entity struct {
	address string
	public  abstract.Point
	secret  abstract.Scalar
}

// Roster is a grouping of entities (nodes) of the mixnet.
type Roster []Entity

// Create a new roster from a given TOML file.
func roster(file string, suite abstract.Suite) (Roster, error) {
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

	roster := make([]Entity, len(parsed.Servers))

	for index, server := range parsed.Servers {
		entity := Entity{server.Address, suite.Point(), suite.Scalar()}

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

func (roster Roster) keys() []abstract.Point {
	keys := make([]abstract.Point, 0, len(roster))
	for _, key := range roster {
		keys = append(keys, key.public)
	}

	return keys
}

func (roster Roster) send(index int, message Message) error {
	address := roster[index].address
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	channel := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	encoder := gob.NewEncoder(channel)
	decoder := gob.NewDecoder(channel)

	err = encoder.Encode(message)
	if err != nil {
		return err
	}
	_ = channel.Flush()

	var response Message
	err = decoder.Decode(&response)
	if err != nil {
		return err
	} else if response.Kind != ack {
		return errors.New(string(response.Encoding))
	}

	return nil
}

func (roster Roster) broadcast(host int, message Message) {
	for index := range roster {
		if index != host {
			_ = roster.send(index, message)
		}
	}
}

func (roster Roster) broadcastTo(host int, indexes []int, message Message) {
	for _, index := range indexes {
		if index != host {
			_ = roster.send(index, message)
		}
	}
}
