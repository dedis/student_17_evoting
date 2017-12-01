package service

import (
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

const Name = "nevv"

var serviceID onet.ServiceID

type Service struct {
	*onet.ServiceProcessor

	secrets map[string]*dkg.SharedSecret

	state *state
	node  *onet.Roster
	pin   string
}

type synchronizer struct {
	Genesis skipchain.SkipBlockID
}

// Ping is the handler through which the service can be probed. It returns
// the same message with the nonce incremented by one.
func (s *Service) Ping(req *api.Ping) (*api.Ping, onet.ClientError) {
	return &api.Ping{req.Nonce + 1}, nil
}

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

	return &api.LinkReply{genesis.Hash}, nil
}

func (s *Service) Open(req *api.Open) (*api.OpenReply, onet.ClientError) {
	if _, err := s.assertLevel(req.Token, true); err != nil {
		return nil, onet.NewClientError(err)
	}

	master, err := chains.GetMaster(s.node, req.Master)
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
		req.Election.Key = secret.X
		s.secrets[string(genesis.Hash)] = secret

		if _, err := chains.Store(roster, genesis.Hash, req.Election); err != nil {
			return nil, onet.NewClientError(err)
		}
		link := &chains.Link{genesis.Hash}
		if _, err = chains.Store(roster, req.Master, link); err != nil {
			return nil, onet.NewClientError(err)
		}

		return &api.OpenReply{genesis.Hash, secret.X}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("DKG timeout"))
	}
}

func (s *Service) Login(req *api.Login) (*api.LoginReply, onet.ClientError) {
	master, err := chains.GetMaster(s.node, req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	links, err := chains.GetLinks(s.node, req.Master)
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

	token := s.state.register(req.User, master.IsAdmin(req.User))
	return &api.LoginReply{token, elections}, nil
}

func (s *Service) Cast(req *api.Cast) (*api.CastReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.GetElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsUser(user) {
		return nil, onet.NewClientError(errors.New("Invalid user"))
	}

	index, err := chains.Store(election.Roster, req.Genesis, req.Ballot)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.CastReply{uint32(index)}, nil
}

func (s *Service) Aggregate(req *api.Aggregate) (*api.AggregateReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, false)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.GetElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsUser(user) {
		return nil, onet.NewClientError(errors.New("Invalid user"))
	}

	box, err := chains.GetBox(s.node, req.Genesis, req.Type)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.AggregateReply{box}, nil
}

func (s *Service) Finalize(req *api.Finalize) (*api.FinalizeReply, onet.ClientError) {
	user, err := s.assertLevel(req.Token, true)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election, err := chains.GetElection(s.node, req.Genesis)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	if !election.IsCreator(user) {
		return nil, onet.NewClientError(errors.New("Not a creator"))
	}

	decryption, _ := chains.GetBox(s.node, req.Genesis, chains.DECRYPTION)
	if decryption != nil {
		return nil, onet.NewClientError(errors.New("Election already finalized"))
	}

	ballots, _ := chains.GetBallots(s.node, req.Genesis)
	box := &chains.Box{ballots}

	if _, err = chains.Store(election.Roster, req.Genesis, box); err != nil {
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

		instance, _ := s.CreateProtocol(decrypt.Name, tree)
		protocol := instance.(*decrypt.Protocol)
		protocol.Secret = s.secrets[string(req.Genesis)]
		protocol.Shuffle = shuffled

		config, _ := network.Marshal(&synchronizer{req.Genesis})
		if err = protocol.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
			return nil, onet.NewClientError(err)
		}

		if err = protocol.Start(); err != nil {
			return nil, onet.NewClientError(err)
		}

		select {
		case <-protocol.Finished:
			return &api.FinalizeReply{shuffled, protocol.Decryption}, nil
		case <-time.After(2 * time.Second):
			return nil, onet.NewClientError(errors.New("Decrypt timeout"))
		}

		return &api.FinalizeReply{}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
	}

	return &api.FinalizeReply{}, nil
}

func (s *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	// Unmarshal synchronizer structure.
	unmarshal := func(data []byte) *synchronizer {
		_, blob, _ := network.Unmarshal(conf.Data)
		return blob.(*synchronizer)
	}

	switch node.ProtocolName() {
	case dkg.Name:
		instance, _ := dkg.New(node)
		protocol := instance.(*dkg.Protocol)
		go func() {
			<-protocol.Done
			secret, _ := protocol.SharedSecret()
			s.secrets[string(unmarshal(conf.Data).Genesis)] = secret
		}()
		return protocol, nil
	case shuffle.Name:
		instance, err := shuffle.New(node)
		if err != nil {
			return nil, err
		}
		return instance.(*shuffle.Protocol), nil
	case decrypt.Name:
		instance, err := decrypt.New(node)
		if err != nil {
			return nil, err
		}

		protocol := instance.(*decrypt.Protocol)
		// _, blob, _ := network.Unmarshal(config.Data)
		// sync := blob.(*synchronizer)
		// protocol.Chain = service.Storage.Chains[sync.ElectionName]
		protocol.Secret = s.secrets[string(unmarshal(conf.Data).Genesis)]
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
	)
	service.node = onet.NewRoster([]*network.ServerIdentity{service.ServerIdentity()})

	return service
}
