package main

import (
	"crypto/cipher"
	"errors"
	"log"

	"github.com/dedis/kyber/abstract"
	"github.com/dedis/kyber/share/dkg"
	"github.com/dedis/protobuf"
)

// Session designates a running process that hosts distributed key generations
// and shuffling procedures.
type Session struct {
	// Identifier of this session
	name string

	// Kyber DKG object
	generator *dkg.DistKeyGenerator

	// Roster of assigned nodes
	roster Roster

	// Indexes of certified nodes in roster
	qual []int

	// Index of current host in roster
	index int

	// Collection of generated responses before distribution
	responses []*dkg.Response
}

// Create a new session object for a given name and roster.
func session(name string, suite abstract.Suite, stream cipher.Stream,
	file, host string) (*Session, error) {

	session := Session{}
	session.name = name

	roster, err := roster(file, suite)
	if err != nil {
		return nil, err
	}

	// Find host in roster.
	var secret abstract.Scalar
	for i, entity := range roster {
		if entity.address == host {
			session.index = i
			secret = entity.secret
		}
	}

	keys := roster.keys()
	generator, err := dkg.NewDistKeyGenerator(suite, secret, keys, stream, len(keys)-1)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	session.roster = roster
	session.generator = generator
	session.qual = make([]int, 0)
	session.responses = make([]*dkg.Response, 0)

	return &session, nil
}

// startDeal generates the deals from the DKG generator and distributes
// them to the corresponding nodes of the mixnet.
func (session *Session) startDeal() error {
	deals, err := session.generator.Deals()
	if err != nil {
		return err
	}

	for index, deal := range deals {
		encoding, err := protobuf.Encode(deal)
		if err != nil {
			log.Println(err)
			return err
		}

		message := Message{msgDeal, session.name, len(encoding), encoding}
		if err = session.roster.send(index, message); err != nil {
			return err
		}
	}

	return nil
}

// deal processes an incoming deal object and saves the created response objects.
func (session *Session) deal(deal *dkg.Deal) error {
	response, err := session.generator.ProcessDeal(deal)
	if err != nil {
		log.Println(err)
		return err
	}

	session.responses = append(session.responses, response)

	return nil
}

// startResponse takes the saved response objects and distributes them amongst
// the nodes of the mixnet.
func (session *Session) startResponse() error {
	if len(session.responses) != len(session.roster)-1 {
		log.Println("Not all responses available")
		return errors.New("")
	}

	for _, response := range session.responses {
		encoding, err := protobuf.Encode(response)
		if err != nil {
			log.Println(err)
			return err
		}

		message := Message{msgResponse, session.name, len(encoding), encoding}
		session.roster.broadcast(session.index, message)
	}

	return nil
}

// response processes an incoming response object.
func (session *Session) response(response *dkg.Response) error {
	justification, err := session.generator.ProcessResponse(response)
	if err != nil {
		log.Println(err)
		return err
	} else if justification != nil {
		encoding, err := protobuf.Encode(justification)
		if err != nil {
			log.Println(err)
			return err
		}

		message := Message{msgResponse, session.name, len(encoding), encoding}
		session.roster.broadcast(session.index, message)
	}

	return nil
}

// justification processes an incoming justification object.
func (session *Session) justification(justification *dkg.Justification) error {
	if err := session.generator.ProcessJustification(justification); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// startCommit generates the shared commit from the DKG generator and broadcasts it
// to the set of certified nodes.
func (session *Session) startCommit() error {
	if !session.generator.Certified() {
		log.Println("Certification not established")
		return errors.New("")
	}

	session.qual = session.generator.QUAL()

	commits, _ := session.generator.SecretCommits()

	encoding, err := protobuf.Encode(commits)
	if err != nil {
		log.Println(err)
		return err
	}

	message := Message{msgCommit, session.name, len(encoding), encoding}
	session.roster.broadcastTo(session.index, session.qual, message)

	return nil
}

// commit processes an incoming commit objectj.
func (session *Session) commit(commits *dkg.SecretCommits) error {
	// TODO: complaints
	_, err := session.generator.ProcessSecretCommits(commits)
	if err != nil {
		return err
	}

	return nil
}

// sharedKey retrieves the shared public key from the after the successful
// termination of the protocol.
func (session *Session) sharedKey() (abstract.Point, error) {
	share, err := session.generator.DistKeyShare()
	if err != nil {
		return nil, err
	}

	return share.Public(), nil
}
