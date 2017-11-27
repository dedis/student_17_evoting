package service

import (
	"errors"
	"time"

	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/dkg"
	"github.com/qantik/nevv/storage"
)

const Name = "nevv"

type Service struct {
	*onet.ServiceProcessor

	Storage *storage.Storage

	secrets map[string]*dkg.SharedSecret

	state *state
	Pin   string
}

type synchronizer struct {
	Genesis skipchain.SkipBlockID
}

var serviceID onet.ServiceID

func init() {
	network.RegisterMessage(&synchronizer{})
	serviceID, _ = onet.RegisterNewService(Name, new)
}

func (s *Service) NewProtocol(node *onet.TreeNodeInstance, config *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	switch node.ProtocolName() {
	case dkg.Name:
		instance, _ := dkg.New(node)
		protocol := instance.(*dkg.Protocol)
		go func() {
			<-protocol.Done

			secret, _ := protocol.SharedSecret()
			_, blob, _ := network.Unmarshal(config.Data)
			sync := blob.(*synchronizer)
			s.secrets[string(sync.Genesis)] = secret
		}()
		return protocol, nil
	// case shuffle.Name:
	// 	instance, err := shuffle.New(node)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return instance.(*shuffle.Protocol), nil
	// case decrypt.Name:
	// 	instance, err := decrypt.New(node)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	protocol := instance.(*decrypt.Protocol)
	// 	_, blob, _ := network.Unmarshal(config.Data)
	// 	sync := blob.(*synchronizer)
	// 	protocol.Chain = service.Storage.Chains[sync.ElectionName]
	// 	return protocol, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

func (s *Service) Ping(req *api.Ping) (*api.Ping, onet.ClientError) {
	return &api.Ping{req.Nonce + 1}, nil
}

// func (s *Service) GenerateElection(req *api.GenerateElection) (
// 	*api.GenerateElectionResponse, onet.ClientError) {

// 	election := req.Election

// 	client := skipchain.NewClient()
// 	genesis, _ := client.CreateGenesis(election.Roster, 1, 1,
// 		skipchain.VerificationNone, nil, nil)

// 	size := len(election.Roster.List)
// 	tree := election.Roster.GenerateNaryTreeWithRoot(size, s.ServerIdentity())
// 	instance, err := s.CreateProtocol(dkg.Name, tree)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	protocol := instance.(*dkg.Protocol)
// 	protocol.Wait = true

// 	config, _ := network.Marshal(&synchronizer{election.ID, genesis})
// 	if err = protocol.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if err = protocol.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-protocol.Done:
// 		shared, _ := protocol.SharedSecret()
// 		election.Key = shared.X

// 		chain := &storage.Chain{Genesis: genesis, SharedSecret: shared}
// 		_, _ = chain.Store(&election)
// 		s.Storage.Chains[election.ID] = chain
// 		s.save()

// 		return &api.GenerateElectionResponse{shared.X}, nil
// 	case <-time.After(2 * time.Second):
// 		return nil, onet.NewClientError(errors.New("DKG timeout"))
// 	}
// }

// func (s *Service) GetElections(req *api.GetElections) (
// 	*api.GetElectionsReply, onet.ClientError) {

// 	elections := s.Storage.ElectionsForUser(req.User)

// 	return &api.GetElectionsReply{elections}, nil
// }

// func (s *Service) CastBallot(req *api.CastBallot) (*api.CastBallotResponse, onet.ClientError) {
// 	chain, found := s.Storage.Chains[req.ID]
// 	if !found {
// 		return nil, onet.NewClientError(errors.New("Election not found"))
// 	}

// 	index, err := chain.Store(&req.Ballot)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	return &api.CastBallotResponse{uint32(index)}, nil
// }

// func (s *Service) GetBallots(req *api.GetBallots) (*api.GetBallotsResponse, onet.ClientError) {
// 	chain, found := s.Storage.Chains[req.ID]
// 	if !found {
// 		return nil, onet.NewClientError(errors.New("Election not found"))
// 	}

// 	ballots, err := chain.Ballots()
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	return &api.GetBallotsResponse{ballots}, nil
// }

// func (s *Service) Shuffle(req *api.Shuffle) (*api.ShuffleReply, onet.ClientError) {
// 	chain, found := s.Storage.Chains[req.ID]
// 	if !found {
// 		return nil, onet.NewClientError(errors.New("Election not found"))
// 	}

// 	tree := chain.Genesis.Roster.GenerateNaryTreeWithRoot(1, s.ServerIdentity())
// 	instance, err := s.CreateProtocol(shuffle.Name, tree)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	protocol := instance.(*shuffle.Protocol)
// 	protocol.Chain = chain

// 	if err = protocol.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-protocol.Finished:
// 		return &api.ShuffleReply{uint32(protocol.Index)}, nil
// 	case <-time.After(2 * time.Second):
// 		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
// 	}
// }

// func (s *Service) GetShuffle(req *api.GetShuffle) (*api.GetShuffleReply, onet.ClientError) {
// 	chain, found := s.Storage.Chains[req.ID]
// 	if !found {
// 		return nil, onet.NewClientError(errors.New("Election not found"))
// 	}

// 	if !chain.IsShuffled() {
// 		return nil, onet.NewClientError(errors.New("No shuffle available"))
// 	}

// 	boxes, err := chain.Boxes()
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	return &api.GetShuffleReply{boxes[0]}, nil
// }

// func (s *Service) Decrypt(req *api.Decrypt) (*api.DecryptReply, onet.ClientError) {
// 	chain, found := s.Storage.Chains[req.ID]
// 	if !found {
// 		return nil, onet.NewClientError(errors.New("Election not found"))
// 	}

// 	if !chain.IsShuffled() || chain.IsDecrypted() {
// 		return nil, onet.NewClientError(errors.New("Decryption not possible"))
// 	}

// 	tree := chain.Genesis.Roster.GenerateNaryTreeWithRoot(2, s.ServerIdentity())
// 	if tree == nil {
// 		return nil, onet.NewClientError(errors.New("Could not generate tree"))
// 	}

// 	instance, err := s.CreateProtocol(decrypt.Name, tree)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	protocol := instance.(*decrypt.Protocol)
// 	protocol.Chain = chain

// 	config, _ := network.Marshal(&synchronizer{req.ID})
// 	if err = protocol.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if err := protocol.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-protocol.Finished:
// 		return &api.DecryptReply{protocol.Index}, nil
// 	case <-time.After(2000 * time.Millisecond):
// 		return nil, onet.NewClientError(errors.New("Decryption timeout"))
// 	}
// }

func (s *Service) Link(req *api.Link) (*api.LinkReply, onet.ClientError) {
	if req.Pin == "" {
		log.Lvl3("Current session ping:", s.Pin)
		return &api.LinkReply{}, nil
	} else if req.Pin != s.Pin {
		return nil, onet.NewClientError(errors.New("Wrong ping"))
	}

	master := &master{req.Key, req.Admins}

	client := skipchain.NewClient()
	genesis, _ := client.CreateGenesis(req.Roster, 1, 1,
		skipchain.VerificationStandard, master, nil)

	return &api.LinkReply{genesis.Hash}, nil
}

func (s *Service) Open(req *api.Open) (*api.OpenReply, onet.ClientError) {
	user, found := s.state.log[req.Token]
	if !found {
		return nil, onet.NewClientError(errors.New("Not logged in"))
	} else if !user.admin {
		return nil, onet.NewClientError(errors.New("Need admin privilege"))
	}

	roster := onet.NewRoster([]*network.ServerIdentity{s.ServerIdentity()})

	client := skipchain.NewClient()
	master, err := client.GetUpdateChain(roster, req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	roster = master.Update[0].Roster
	genesis, _ := client.CreateGenesis(roster, 1, 1,
		skipchain.VerificationStandard, req.Election, nil)

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

		client.StoreSkipBlock(genesis, roster, req.Election)

		latest := master.Update[len(master.Update)-1]
		client.StoreSkipBlock(latest, roster, &link{genesis.Hash})

		return &api.OpenReply{genesis.Hash, secret.X}, nil
	case <-time.After(time.Second):
		return nil, onet.NewClientError(errors.New("DKG timeout"))
	}
}

func (s *Service) Login(req *api.Login) (*api.LoginReply, onet.ClientError) {
	roster := onet.NewRoster([]*network.ServerIdentity{s.ServerIdentity()})

	client := skipchain.NewClient()
	chain, err := client.GetUpdateChain(roster, req.Master)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	_, blob, _ := network.Unmarshal(chain.Update[0].Data)
	master := blob.(*master)

	token := s.state.register(req.Sciper, master.admin(req.Sciper))

	elections := make([]*api.EElection, 0)
	for i := 1; i < len(chain.Update); i++ {
		_, blob, _ := network.Unmarshal(chain.Update[i].Data)
		link := blob.(*link)

		electionChain, _ := client.GetUpdateChain(roster, link.Genesis)
		_, blob, _ = network.Unmarshal(electionChain.Update[1].Data)
		election := blob.(*api.EElection)

		for _, user := range election.Users {
			if user == req.Sciper {
				elections = append(elections, election)
			}
		}

	}
	return &api.LoginReply{token, elections}, nil
}

func (service *Service) save() {
	service.Storage.Lock()
	defer service.Storage.Unlock()

	err := service.Save(Name, service.Storage)
	if err != nil {
		log.Error(err)
	}
}

func (service *Service) load() error {
	service.Storage = &storage.Storage{Chains: make(map[string]*storage.Chain)}
	if !service.DataAvailable(Name) {
		return nil
	}

	msg, err := service.Load(Name)
	if err != nil {
		return err
	}
	service.Storage = msg.(*storage.Storage)
	// service.Pin = nonce(6)

	return nil
}

func new(context *onet.Context) onet.Service {
	service := &Service{
		ServiceProcessor: onet.NewServiceProcessor(context),
		secrets:          make(map[string]*dkg.SharedSecret),
		state:            &state{make(map[string]*user)},
		Pin:              nonce(6),
	}

	if err := service.RegisterHandlers(
		service.Ping, service.Link, service.Open); err != nil {
		log.ErrFatal(err)
	}

	if err := service.load(); err != nil {
		log.Error(err)
	}

	return service
}
