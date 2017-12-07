package service

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/decrypt"
	"github.com/qantik/nevv/dkg"
	"github.com/qantik/nevv/shuffle"
)

func init() {
	network.RegisterMessage(&synchronizer{})
	serviceID, _ = onet.RegisterNewService(Name, new)
}

// Name is the identifier of the service (application name).
const Name = "nevv"

// serviceID is the onet services identifier. Only used for testing.
var serviceID onet.ServiceID

// Service is the application's core structure. It is the first object that
// is created upon startup, registering all the message handlers. All in all
// the nevv service tries to be as stateless as possible (REST interface) apart
// from the map of registered users and the shared secrets stored after every
// execution of the distributed key generation protocol.
type Service struct {
	// onet processor. All handler functions are attached to it.
	*onet.ServiceProcessor

	// secrets stores the shared secrets for each election. This is
	// different for each node participating in the DKG.
	secrets map[string]*dkg.SharedSecret

	// state is the log of currently logged in users.
	state *state
	// node is a unitary roster only consisting of this conode.
	node *onet.Roster
	// pin is the current service number. Used to authenticate link messages.
	pin string
}

// synchronizer is sent before the start of a protocol to make sure all
// nodes of the roster have to ID of the involved election Skipchain.
type synchronizer struct {
	// Genesis is the ID of an election Skipchain.
	ID skipchain.SkipBlockID
}

// Ping is the handler through which the service can be probed. It returns
// the same message with the nonce incremented by one.
func (s *Service) Ping(req *api.Ping) (*api.Ping, onet.ClientError) {
	return &api.Ping{req.Nonce + 1}, nil
}

// Link is the handler through which a new master Skipchain can be registered
// at the service. It will print the session pin if it is not specified in the
// request. It returns the ID of the newly created master Skipchain.
func (s *Service) Link(req *api.Link) (*api.LinkReply, onet.ClientError) {
	if req.Pin == "" {
		log.Lvl3("Current session ping:", s.pin)
		return &api.LinkReply{}, nil
	} else if req.Pin != s.pin {
		return nil, onet.NewClientError(errors.New("Wrong ping"))
	}

	master := &chains.Master{req.Key, req.Roster, req.Admins}
	genesis, err := chains.Create(req.Roster, master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.LinkReply{base64.StdEncoding.EncodeToString(genesis.Hash)}, nil
}

// Open is the handler through which a new election can be created by an
// administrator. It performs the distributed key generation protocol to
// establish a shared public key for the election. This key as well as the
// ID of the newly created election Skipchain are returned.
func (s *Service) Open(req *api.Open) (*api.OpenReply, onet.ClientError) {
	log.Lvl3("OPEN", s.state.log)
	if _, err := s.assertLevel(req.Token, true); err != nil {
		return nil, onet.NewClientError(err)
	}

	master, masterID, err := s.fetchMaster(req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}
	roster := master.Roster

	genesis, err := chains.Create(roster, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	tree := roster.GenerateNaryTreeWithRoot(len(roster.List), s.ServerIdentity())
	instance, _ := s.CreateProtocol(dkg.Name, tree)
	protocol := instance.(*dkg.Protocol)
	protocol.Wait = true

	config, _ := network.Marshal(&synchronizer{genesis.Hash})
	protocol.SetConfig(&onet.GenericConfig{Data: config})
	protocol.Start()
	select {
	case <-protocol.Done:
		secret, _ := protocol.SharedSecret()
		req.Election.ID = base64.StdEncoding.EncodeToString(genesis.Hash)
		req.Election.Roster = roster
		req.Election.Key = secret.X
		s.secrets[string(genesis.Hash)] = secret

		// Store election on its Skipchain and add link to master Skipchain.
		if _, err := chains.Store(roster, genesis.Hash, req.Election); err != nil {
			return nil, onet.NewClientError(err)
		}
		link := &chains.Link{genesis.Hash}
		if _, err = chains.Store(roster, masterID, link); err != nil {
			return nil, onet.NewClientError(err)
		}

		return &api.OpenReply{
			base64.StdEncoding.EncodeToString(genesis.Hash),
			secret.X,
		}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("DKG timeout"))
	}
}

// Login enables a user to register himself at the services. It checks the
// user's permission level in the master Skipchain and creates a new entry
// in the log. It returns a list of all elections said user is participating in.
func (s *Service) Login(req *api.Login) (*api.LoginReply, onet.ClientError) {
	master, id, err := s.fetchMaster(req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}
	links, err := chains.GetLinks(s.node, id)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	elections := make([]*chains.Election, 0)
	for _, link := range links {
		election, err := chains.GetElection(s.node, link.Genesis)
		if err != nil {
			return nil, onet.NewClientError(err)
		}

		if election.IsUser(req.User) {
			elections = append(elections, election)
		}
	}

	log.Lvl3(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", req)
	log.Lvl3(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", master)
	log.Lvl3(master.IsAdmin(req.User))
	token := s.state.register(req.User, master.IsAdmin(req.User))

	log.Lvl3(s.assertLevel(token, true))

	return &api.LoginReply{token, elections}, nil
}

// Cast is the handler through which a user can cast a ballot in an election.
// If the user is actually a participator in the election then his ballot
// is appended to the election Skipchain in a separate block. The function
// returns the index of the said block containing the ballot.
func (s *Service) Cast(req *api.Cast) (*api.CastReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, id, err := s.fetchElection(req.Genesis, user, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	// If there exist boxes the election is already finalized.
	box, _ := chains.GetBox(s.node, id, chains.SHUFFLE)
	if box != nil {
		return nil, onet.NewClientError(errors.New("Election already finalized"))
	}

	index, err := chains.Store(election.Roster, id, req.Ballot)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.CastReply{uint32(index)}, nil
}

// Aggregate is the handler through which a box of decrypted, shuffled or
// decrypted ballots of an election can be retrieved.
func (s *Service) Aggregate(req *api.Aggregate) (*api.AggregateReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	_, id, err := s.fetchElection(req.Genesis, user, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	box, err := chains.GetBox(s.node, id, req.Type)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.AggregateReply{box}, nil
}

// Finalize is called by an election creator to terminate the poll. It runs the
// shuffle and decryption protocol one after another before returning the both
// boxes (shuffle, decryption). To simplify further access to the encrypted
// ballots all ballots are firsts collected and then stored in a separate box
// inside a new Skipblock.
func (s *Service) Finalize(req *api.Finalize) (*api.FinalizeReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, true)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, id, err := s.fetchElection(req.Genesis, user, true)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	// If there exist boxes, election is already finalized.
	decryption, _ := chains.GetBox(s.node, id, chains.DECRYPTION)
	if decryption != nil {
		return nil, onet.NewClientError(errors.New("Election already finalized"))
	}

	// Store ballots in a box on the Skipchain.
	ballots, _ := chains.GetBallots(s.node, id)
	box := &chains.Box{ballots}
	if _, err = chains.Store(election.Roster, id, box); err != nil {
		return nil, onet.NewClientError(err)
	}

	tree := election.Roster.GenerateNaryTreeWithRoot(1, s.ServerIdentity())
	instance, _ := s.CreateProtocol(shuffle.Name, tree)
	protocol := instance.(*shuffle.Protocol)
	protocol.Key = election.Key
	protocol.Box = box
	if err = protocol.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}
	select {
	case <-protocol.Finished:
		shuffled := protocol.Shuffle
		if _, err = chains.Store(election.Roster, id, shuffled); err != nil {
			return nil, onet.NewClientError(err)
		}

		instance, _ := s.CreateProtocol(decrypt.Name, tree)
		protocol := instance.(*decrypt.Protocol)
		protocol.Secret = s.secrets[string(id)]
		protocol.Shuffle = shuffled

		config, _ := network.Marshal(&synchronizer{id})
		if err = protocol.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
			return nil, onet.NewClientError(err)
		}

		if err = protocol.Start(); err != nil {
			return nil, onet.NewClientError(err)
		}
		select {
		case <-protocol.Finished:
			_, err = chains.Store(election.Roster, id, protocol.Decryption)
			if err != nil {
				return nil, onet.NewClientError(err)
			}
			return &api.FinalizeReply{shuffled, protocol.Decryption}, nil
		case <-time.After(2 * time.Second):
			return nil, onet.NewClientError(errors.New("Decrypt timeout"))
		}

		return &api.FinalizeReply{}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
	}
}

// NewProtocol is called by the onet processor on non-root nodes to signal
// the initialization of a new protocol. Here, the synchronizer message is
// received and processed by the non-root nodes before the protocol starts.
func (s *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	// Unmarshal synchronizer structure.
	unmarshal := func(data []byte) *synchronizer {
		_, blob, _ := network.Unmarshal(conf.Data)
		return blob.(*synchronizer)
	}

	switch node.ProtocolName() {
	// Retrieve and store shared secret after DKG has finished.
	case dkg.Name:
		instance, _ := dkg.New(node)
		protocol := instance.(*dkg.Protocol)
		go func() {
			<-protocol.Done
			secret, _ := protocol.SharedSecret()
			s.secrets[string(unmarshal(conf.Data).ID)] = secret
		}()
		return protocol, nil
	// Only initialize the shuffle protocol.
	case shuffle.Name:
		instance, err := shuffle.New(node)
		if err != nil {
			return nil, err
		}
		return instance.(*shuffle.Protocol), nil
	// Pass conode's shared secret to the decrypt protocol.
	case decrypt.Name:
		instance, err := decrypt.New(node)
		if err != nil {
			return nil, err
		}
		protocol := instance.(*decrypt.Protocol)
		protocol.Secret = s.secrets[string(unmarshal(conf.Data).ID)]
		return protocol, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

// assertLevel is a helper function that verifies in the log if a given user is
// registered in the service and has admin level if required and then returns said user.
func (s *Service) assertLevel(token string, admin bool) (chains.User, error) {
	log.Lvl3("STAMP", s.state.log)
	stamp, found := s.state.log[token]
	if !found {
		return 0, errors.New("Not logged in")
	}

	if admin && !stamp.admin {
		return 0, errors.New("Need admin level")
	}

	return stamp.user, nil
}

// fetchElection is a helper function that retrieves an election from a Skipchain
// and verifies the given user's priviledge. It returns an election object and its
// Skipchain identifier.
func (s *Service) fetchElection(id string, user chains.User, creator bool) (
	*chains.Election, []byte, error) {

	electionID, _ := base64.StdEncoding.DecodeString(id)
	election, err := chains.GetElection(s.node, electionID)
	if err != nil {
		return nil, nil, err
	}

	if (!creator && election.IsUser(user)) || (creator && election.IsCreator(user)) {
		return election, electionID, nil
	}

	return nil, nil, errors.New("User is neither creator nor registered in election")
}

// fetchMaster is a helper function that retrieves an master object from the master
// Skipchain. Returning the master object and its Skipchain identifier.
func (s *Service) fetchMaster(id string) (*chains.Master, []byte, error) {
	masterID, _ := base64.StdEncoding.DecodeString(id)
	master, err := chains.GetMaster(s.node, masterID)
	if err != nil {
		return nil, nil, onet.NewClientError(err)
	}

	return master, masterID, nil
}

// new initializes the service and registers all the message handlers.
func new(context *onet.Context) onet.Service {
	service := &Service{
		ServiceProcessor: onet.NewServiceProcessor(context),
		secrets:          make(map[string]*dkg.SharedSecret),
		state:            &state{make(map[string]*stamp)},
		pin:              nonce(6),
	}

	service.RegisterHandlers(
		service.Ping,
		service.Link,
		service.Open,
		service.Login,
		service.Aggregate,
		service.Finalize,
	)
	service.node = onet.NewRoster([]*network.ServerIdentity{service.ServerIdentity()})

	return service
}
