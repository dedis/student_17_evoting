package service

import (
	"encoding/base64"
	"errors"
	"time"

	"gopkg.in/dedis/cothority.v1/skipchain"
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
	ServiceID, _ = onet.RegisterNewService(Name, new)
}

// Name is the identifier of the service (application name).
const Name = "nevv"

// serviceID is the onet services identifier. Only used for testing.
var ServiceID onet.ServiceID

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
	Pin string
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
		log.Lvl3("Current session ping:", s.Pin)
		return &api.LinkReply{}, nil
	}
	if req.Pin != s.Pin {
		return nil, onet.NewClientError(errors.New("Wrong ping"))
	}

	genesis, err := chains.New(req.Roster, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	master := &chains.Master{req.Key, genesis.Hash, req.Roster, req.Admins}
	if _, err = master.Append(master); err != nil {
		return nil, onet.NewClientError(err)
	}
	return &api.LinkReply{base64.StdEncoding.EncodeToString(genesis.Hash)}, nil
}

// Open is the handler through which a new election can be created by an
// administrator. It performs the distributed key generation protocol to
// establish a shared public key for the election. This key as well as the
// ID of the newly created election Skipchain are returned.
func (s *Service) Open(req *api.Open) (*api.OpenReply, onet.ClientError) {
	if _, err := s.assertLevel(req.Token, true); err != nil {
		return nil, onet.NewClientError(err)
	}

	master, err := chains.FetchMaster(s.node, req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}
	roster := master.Roster

	genesis, err := chains.New(roster, nil)
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
		req.Election.Stage = 0
		s.secrets[string(genesis.Hash)] = secret

		// Store election on its Skipchain and add link to master Skipchain.
		if _, err := req.Election.Append(req.Election); err != nil {
			return nil, onet.NewClientError(err)
		}
		link := &chains.Link{genesis.Hash}
		if _, err = master.Append(link); err != nil {
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
	master, err := chains.FetchMaster(s.node, req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}
	links, err := master.Links()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	elections := make([]*chains.Election, 0)
	for _, link := range links {
		id := base64.StdEncoding.EncodeToString(link.Genesis)
		election, err := chains.FetchElection(s.node, id)
		if err != nil {
			return nil, onet.NewClientError(err)
		}

		if election.IsUser(req.User) || election.IsCreator(req.User) {
			elections = append(elections, election)
		}
	}

	admin := master.IsAdmin(req.User)
	token := s.state.register(req.User, admin)
	return &api.LoginReply{token, admin, elections}, nil
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

	election, err := chains.FetchElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsUser(user) {
		return nil, onet.NewClientError(errors.New("User not part of election"))
	}

	if election.Stage > 0 {
		return nil, onet.NewClientError(errors.New("Election already closed"))
	}

	index, err := election.Append(req.Ballot)
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
	if req.Type > 2 {
		return nil, onet.NewClientError(errors.New("Invalid aggregation type"))
	}

	election, err := chains.FetchElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsUser(user) {
		return nil, onet.NewClientError(errors.New("User not part of election"))
	}

	var box *chains.Box

	switch req.Type {
	case chains.BALLOTS:
		box, err = election.Ballots()
	case chains.SHUFFLE:
		box, err = election.Shuffle()
	case chains.DECRYPTION:
		box, err = election.Decryption()
	}

	if err != nil {
		return nil, onet.NewClientError(err)
	}
	return &api.AggregateReply{box}, nil
}

func (s *Service) Shuffle(req *api.Shuffle) (*api.ShuffleReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, true)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.FetchElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsCreator(user) {
		return nil, onet.NewClientError(errors.New("Only creators can shuffle"))
	}

	if election.Stage >= 1 {
		return nil, onet.NewClientError(errors.New("Election already shuffled"))
	}

	ballots, err := election.Ballots()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	tree := election.Roster.GenerateNaryTreeWithRoot(1, s.ServerIdentity())
	instance, _ := s.CreateProtocol(shuffle.Name, tree)
	protocol := instance.(*shuffle.Protocol)
	protocol.Key = election.Key
	protocol.Box = ballots

	if err = protocol.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-protocol.Finished:
		if _, err = election.Append(protocol.Shuffle); err != nil {
			return nil, onet.NewClientError(err)
		}
		return &api.ShuffleReply{protocol.Shuffle}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
	}
}

func (s *Service) Decrypt(req *api.Decrypt) (*api.DecryptReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, true)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.FetchElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsCreator(user) {
		return nil, onet.NewClientError(errors.New("Only creators can shuffle"))
	}

	if election.Stage >= 2 {
		return nil, onet.NewClientError(errors.New("Election already decrypted"))
	}

	id, _ := base64.StdEncoding.DecodeString(election.ID)

	shuffled, err := election.Shuffle()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	tree := election.Roster.GenerateNaryTreeWithRoot(1, s.ServerIdentity())
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
		if _, err = election.Append(protocol.Decryption); err != nil {
			return nil, onet.NewClientError(err)
		}
		return &api.DecryptReply{protocol.Decryption}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("Decrypt timeout"))
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
	stamp, found := s.state.log[token]
	if !found {
		return 0, errors.New("Not logged in")
	}

	if admin && !stamp.admin {
		return 0, errors.New("Need admin level")
	}

	return stamp.user, nil
}

// new initializes the service and registers all the message handlers.
func new(context *onet.Context) onet.Service {
	service := &Service{
		ServiceProcessor: onet.NewServiceProcessor(context),
		secrets:          make(map[string]*dkg.SharedSecret),
		state:            &state{make(map[string]*stamp)},
		Pin:              nonce(6),
	}

	service.RegisterHandlers(
		service.Ping,
		service.Link,
		service.Open,
		service.Login,
		service.Cast,
		service.Aggregate,
		service.Shuffle,
		service.Decrypt,
	)
	service.node = onet.NewRoster([]*network.ServerIdentity{service.ServerIdentity()})

	return service
}
