package service

import (
	"errors"
	"time"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
	"github.com/qantik/nevv/dkg"
	"github.com/qantik/nevv/shuffle"
)

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
	pin string
}

// synchronizer is sent before the start of a protocol to make sure all
// nodes of the roster have to ID of the involved election Skipchain.
type synchronizer struct {
	// Genesis is the ID of an election Skipchain.
	ID skipchain.SkipBlockID
}

func init() {
	network.RegisterMessage(synchronizer{})
	ServiceID, _ = onet.RegisterNewService(Name, new)
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

	genesis, err := chains.New(req.Roster, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	master := &chains.Master{req.Key, genesis.Hash, req.Roster, req.Admins}
	if _, err := chains.Store(req.Roster, genesis.Hash, master); err != nil {
		return nil, onet.NewClientError(err)
	}
	return &api.LinkReply{genesis.Hash}, nil
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

	genesis, err := chains.New(master.Roster, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	size := len(master.Roster.List)
	tree := master.Roster.GenerateNaryTreeWithRoot(size, s.ServerIdentity())
	instance, _ := s.CreateProtocol(dkg.Name, tree)
	protocol := instance.(*dkg.Protocol)
	protocol.Wait = true

	config, _ := network.Marshal(&synchronizer{genesis.Hash})
	protocol.SetConfig(&onet.GenericConfig{Data: config})

	if err = protocol.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-protocol.Done:
		secret, _ := protocol.SharedSecret()
		req.Election.ID = genesis.Hash
		req.Election.Roster = master.Roster
		req.Election.Key = secret.X
		s.secrets[genesis.Short()] = secret

		// Store election on its Skipchain and add link to master Skipchain.
		_, err := chains.Store(master.Roster, genesis.Hash, req.Election)
		if err != nil {
			return nil, onet.NewClientError(err)
		}
		_, err = chains.Store(master.Roster, master.ID, &chains.Link{genesis.Hash})
		if err != nil {
			return nil, onet.NewClientError(err)
		}

		return &api.OpenReply{genesis.Hash, secret.X}, nil
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
		election, err := chains.FetchElection(s.node, link.Genesis)
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

	if user != req.Ballot.User {
		return nil, onet.NewClientError(errors.New("User != Ballot.User"))
	} else if !election.IsUser(user) && !election.IsCreator(user) {
		return nil, onet.NewClientError(errors.New("User not part of election"))
	} else if election.Stage > 0 {
		return nil, onet.NewClientError(errors.New("Election already closed"))
	}

	index, err := chains.Store(election.Roster, election.ID, req.Ballot)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.CastReply{uint32(index)}, nil
}

func (s *Service) GetBox(req *api.GetBox) (*api.GetBoxReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.FetchElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsUser(user) && !election.IsCreator(user) {
		return nil, onet.NewClientError(errors.New("User not part of election"))
	}

	box, err := election.Box()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.GetBoxReply{Box: box}, nil
}

func (s *Service) GetMixes(req *api.GetMixes) (
	*api.GetMixesReply, onet.ClientError) {

	user, err := s.assertLevel(req.Token, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.FetchElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsUser(user) && !election.IsCreator(user) {
		return nil, onet.NewClientError(errors.New("User not part of election"))
	}

	mixes, err := election.Mixes()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.GetMixesReply{Mixes: mixes}, nil
}

// // Aggregate is the handler through which a box of decrypted, shuffled or
// // decrypted ballots of an election can be retrieved.
// func (s *Service) Aggregate(req *api.Aggregate) (*api.AggregateReply, onet.ClientError) {
// 	user, err := s.assertLevel(req.Token, false)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	election, err := chains.FetchElection(s.node, req.Genesis)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if !election.IsUser(user) && !election.IsCreator(user) {
// 		return nil, onet.NewClientError(errors.New("User not part of election"))
// 	}

// 	var box *chains.Box
// 	switch req.Type {
// 	case 0:
// 		box, err = election.Ballots()
// 	case 1:
// 		box, err = election.Shuffle()
// 	case 2:
// 		box, err = election.Decryption()
// 	default:
// 		return nil, onet.NewClientError(errors.New("Invalid aggregation type"))
// 	}
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	return &api.AggregateReply{box}, nil
// }

// Shuffle is the handler through which the shuffle protcol is initiated for an
// election. The shuffle can only be started by the creator and for elections in
// stage 0, the shuffled ballots are then returned.
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
	} else if election.Stage > 0 {
		return nil, onet.NewClientError(errors.New("Election already shuffled"))
	}

	box, err := election.Ballots()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	// Aggregate ballots and store in single block.
	if _, err = chains.Store(election.Roster, election.ID, box); err != nil {
		return nil, onet.NewClientError(err)
	}

	tree := election.Roster.GenerateNaryTreeWithRoot(1, s.ServerIdentity())
	instance, _ := s.CreateProtocol(shuffle.Name, tree)
	protocol := instance.(*shuffle.Protocol)
	protocol.Election = election

	config, _ := network.Marshal(&synchronizer{election.ID})
	protocol.SetConfig(&onet.GenericConfig{Data: config})

	if err = protocol.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-protocol.Finished:
		return &api.ShuffleReply{}, nil
	case <-time.After(5 * time.Second):
		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
	}
}

// Decrypt is the handler through which the decryption protocol is initiated for an
// election. The decryption can only be started by the creator and for elections in stage
// 1, the decrypted ballots are then returned.
// func (s *Service) Decrypt(req *api.Decrypt) (*api.DecryptReply, onet.ClientError) {
// 	user, err := s.assertLevel(req.Token, true)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	election, err := chains.FetchElection(s.node, req.Genesis)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if !election.IsCreator(user) {
// 		return nil, onet.NewClientError(errors.New("Only creators can shuffle"))
// 	} else if election.Stage > 1 {
// 		return nil, onet.NewClientError(errors.New("Election already decrypted"))
// 	}

// 	shuffled, err := election.Shuffle()
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	tree := election.Roster.GenerateNaryTreeWithRoot(1, s.ServerIdentity())
// 	instance, _ := s.CreateProtocol(decrypt.Name, tree)
// 	protocol := instance.(*decrypt.Protocol)
// 	protocol.Secret = s.secrets[skipchain.SkipBlockID(election.ID).Short()]
// 	protocol.Shuffle = shuffled

// 	config, _ := network.Marshal(&synchronizer{election.ID})
// 	protocol.SetConfig(&onet.GenericConfig{Data: config})

// 	if err = protocol.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-protocol.Finished:
// 		_, err = chains.Store(election.Roster, election.ID, protocol.Decryption)
// 		if err != nil {
// 			return nil, onet.NewClientError(err)
// 		}
// 		return &api.DecryptReply{protocol.Decryption}, nil
// 	case <-time.After(2 * time.Second):
// 		return nil, onet.NewClientError(errors.New("Decrypt timeout"))
// 	}
// }

// NewProtocol is called by the onet processor on non-root nodes to signal
// the initialization of a new protocol. Here, the synchronizer message is
// received and processed by the non-root nodes before the protocol starts.
func (s *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	// Unmarshal synchronizer structure.
	unmarshal := func() skipchain.SkipBlockID {
		_, blob, _ := network.Unmarshal(conf.Data)
		return blob.(*synchronizer).ID
	}

	switch node.ProtocolName() {
	// Retrieve and store shared secret after DKG has finished.
	case dkg.Name:
		instance, _ := dkg.New(node)
		protocol := instance.(*dkg.Protocol)
		go func() {
			<-protocol.Done
			secret, _ := protocol.SharedSecret()
			s.secrets[unmarshal().Short()] = secret
		}()
		return protocol, nil
	// Only initialize the shuffle protocol.
	case shuffle.Name:
		election, err := chains.FetchElection(s.node, unmarshal())
		if err != nil {
			return nil, err
		}

		instance, _ := shuffle.New(node)
		protocol := instance.(*shuffle.Protocol)
		protocol.Election = election

		config, _ := network.Marshal(&synchronizer{election.ID})
		protocol.SetConfig(&onet.GenericConfig{Data: config})

		return protocol, nil
	// Pass conode's shared secret to the decrypt protocol.
	// case decrypt.Name:
	// 	instance, _ := decrypt.New(node)
	// 	protocol := instance.(*decrypt.Protocol)
	// 	protocol.Secret = s.secrets[unmarshal().Short()]
	// 	return protocol, nil
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
		pin:              nonce(6),
	}

	service.RegisterHandlers(
		service.Ping,
		service.Link,
		service.Open,
		service.Login,
		service.Cast,
		service.GetBox,
		service.GetMixes,
		// service.Aggregate,
		service.Shuffle,
		// service.Decrypt,
	)
	service.state.schedule(3 * time.Minute)
	service.node = onet.NewRoster([]*network.ServerIdentity{service.ServerIdentity()})
	return service
}
