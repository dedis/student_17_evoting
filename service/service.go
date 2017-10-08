package service

import (
	"errors"
	"time"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/decrypt"
	"github.com/qantik/nevv/dkg"
	"github.com/qantik/nevv/shuffle"
	"github.com/qantik/nevv/storage"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

var serviceID onet.ServiceID

func init() {
	network.RegisterMessage(&synchronizer{})
	network.RegisterMessage(&Storage{})
	serviceID, _ = onet.RegisterNewService(api.ID, new)
}

type Service struct {
	*onet.ServiceProcessor

	Storage *storage.Storage
}

type synchronizer struct {
	ElectionName string
	Block        *skipchain.SkipBlock
}

// func (service *Service) DecryptionRequest(request *api.DecryptionRequest) (
// 	*api.DecryptionResponse, onet.ClientError) {

// 	election, err := service.Storage.Get(request.Election)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	tree := election.Genesis.Roster.GenerateNaryTreeWithRoot(2, service.ServerIdentity())
// 	if tree == nil {
// 		return nil, onet.NewClientError(errors.New("Could not generate tree"))
// 	}

// 	pi, err := service.CreateProtocol(decrypt.Name, tree)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	protocol := pi.(*decrypt.Protocol)
// 	protocol.Genesis = election.Genesis
// 	protocol.SharedSecret = election.SharedSecret

// 	config, _ := network.Marshal(&synchronizer{request.Election, nil})
// 	if err = protocol.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if err := protocol.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-protocol.Done:
// 		return &api.DecryptionResponse{}, nil
// 	case <-time.After(2000 * time.Millisecond):
// 		return nil, onet.NewClientError(errors.New("Decryption timeout"))
// 	}
// }

// func (service *Service) GenerateRequest(request *api.GenerateRequest) (
// 	*api.GenerateResponse, onet.ClientError) {

// 	length := len(request.Roster.List)
// 	tree := request.Roster.GenerateNaryTreeWithRoot(length, service.ServerIdentity())
// 	protocol, err := service.CreateProtocol(dkg.NameDKG, tree)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	client := skipchain.NewClient()
// 	genesis, err := client.CreateGenesis(request.Roster, 1, 1,
// 		skipchain.VerificationNone, nil, nil)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	setupDKG := protocol.(*dkg.SetupDKG)
// 	setupDKG.Wait = true

// 	config, _ := network.Marshal(&synchronizer{request.Name, genesis})
// 	if err = setupDKG.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if err := protocol.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-setupDKG.Done:
// 		shared, _ := setupDKG.SharedSecret()
// 		election := NewElection(request.Name, genesis, shared)
// 		service.update(election)

// 		fmt.Println(shared.X)

// 		return &api.GenerateResponse{Key: shared.X, Hash: genesis.Hash}, nil
// 	case <-time.After(2000 * time.Millisecond):
// 		return nil, onet.NewClientError(errors.New("DKG timeout"))
// 	}
// }

func (service *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	// _, blob, err := network.Unmarshal(conf.Data)
	// if err != nil {
	// 	return nil, err
	// }
	// sync := blob.(*synchronizer)

	switch node.ProtocolName() {
	case dkg.NameDKG:
		protocol, err := dkg.NewSetupDKG(node)
		if err != nil {
			return nil, err
		}

		// setupDKG := protocol.(*dkg.SetupDKG)
		// go func(conf *onet.GenericConfig) {
		// 	<-setupDKG.Done
		// 	shared, err := setupDKG.SharedSecret()
		// 	if err != nil {
		// 		return
		// 	}

		// 	election := NewElection(sync.ElectionName, sync.Block, shared)
		// 	service.update(election)
		// }(conf)

		return protocol, nil
	case shuffle.Name:
		protocol, err := shuffle.New(node)
		if err != nil {
			return nil, err
		}

		// election, err := service.Storage.Get(sync.ElectionName)
		// if err != nil {
		// 	return nil, err
		// }

		shuffle := protocol.(*shuffle.Protocol)
		// shuffle.Genesis = election.Genesis
		// shuffle.Latest = election.Latest
		// shuffle.Key = election.SharedSecret.X

		// if err = shuffle.SetConfig(&onet.GenericConfig{Data: conf.Data}); err != nil {
		// 	return nil, onet.NewClientError(err)
		// }

		return shuffle, nil
	case decrypt.Name:
		protocol, err := decrypt.New(node)
		if err != nil {
			return nil, err
		}

		// election, _ := service.Storage.Get(sync.ElectionName)

		decr := protocol.(*decrypt.Protocol)
		// decr.Genesis = election.Genesis
		// decr.SharedSecret = election.SharedSecret

		return decr, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

// func (service *Service) CastRequest(request *api.CastRequest) (
// 	*api.CastResponse, onet.ClientError) {

// 	election, err := service.Storage.Get(request.Election)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	client := skipchain.NewClient()
// 	response, err := client.StoreSkipBlock(election.Latest, nil, request.Ballot)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	service.propagate(election.Genesis.Roster.List,
// 		&synchronizer{request.Election, response.Latest})

// 	return &api.CastResponse{}, nil
// }

// func (service *Service) ShuffleRequest(request *api.ShuffleRequest) (
// 	*api.ShuffleResponse, onet.ClientError) {

// 	election, err := service.Storage.Get(request.Election)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	tree := election.Genesis.Roster.GenerateNaryTreeWithRoot(1, service.ServerIdentity())
// 	protocol, err := service.CreateProtocol(shuffle.Name, tree)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	shuffle := protocol.(*shuffle.Protocol)
// 	shuffle.Genesis = election.Genesis
// 	shuffle.Latest = election.Latest
// 	shuffle.Key = election.SharedSecret.X

// 	config, _ := network.Marshal(&synchronizer{request.Election, nil})
// 	if err = shuffle.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	if err = shuffle.Start(); err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	select {
// 	case <-shuffle.Done:
// 		service.propagate(election.Genesis.Roster.List,
// 			&synchronizer{request.Election, shuffle.Latest})
// 		return &api.ShuffleResponse{}, nil
// 	case <-time.After(5000 * time.Millisecond):
// 		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
// 	}
// }

// func (service *Service) FetchRequest(request *api.FetchRequest) (
// 	*api.FetchResponse, onet.ClientError) {

// 	election, err := service.Storage.Get(request.Election)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	client := skipchain.NewClient()
// 	block, err := client.GetSingleBlockByIndex(election.Genesis.Roster,
// 		election.Genesis.Hash, int(request.Block))
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	_, blob, err := network.Unmarshal(block.Data)
// 	if err != nil {
// 		return nil, onet.NewClientError(err)
// 	}

// 	box := blob.(*api.Box)

// 	return &api.FetchResponse{Ballots: box.Ballots}, nil
// }

// Save stores the storage structure on the disk. Has to be called explicitly
// by the service in order to take action.
func (service *Service) save() {
	service.Storage.Lock()
	defer service.Storage.Unlock()

	err := service.Save(api.ID, service.Storage)
	if err != nil {
		log.Error(err)
	}
}

// Load retrieves the the storage structure from the disk and assigns it to
// newly created service.
func (service *Service) load() error {
	// service.Storage = &Storage{
	// 	Elections: make(map[string]*Election),
	// 	Chains:    make(map[string]*skipchain.SkipBlock),
	// }

	service.Storage = &storage.Storage{Chains: make(map[string]*storage.Chain)}
	if !service.DataAvailable(api.ID) {
		return nil
	}

	msg, err := service.Load(api.ID)
	if err != nil {
		return err
	}
	service.Storage = msg.(*storage.Storage)

	return nil
}

// func (service *Service) update(election *Election) {
// 	service.Storage.SetElection(election)
// 	service.save()
// }

func (service *Service) synchronize(envelope *network.Envelope) {
	// sync := envelope.Msg.(*synchronizer)
	// service.Storage.SetLatest(sync.ElectionName, sync.Block)
	// service.save()

	sync := envelope.Msg.(*synchronizer)
	service.Storage.Chains[sync.ElectionName] = &storage.Chain{Genesis: sync.Block}
	service.save()
}

func (service *Service) propagate(list []*network.ServerIdentity, sync *synchronizer) {
	for _, node := range list {
		_ = service.SendRaw(node, sync)
	}
}

// New hooks into the onet registrator to initialize a new service loading
// potential data saved on the disk by an earlier run.
func new(context *onet.Context) onet.Service {
	service := &Service{ServiceProcessor: onet.NewServiceProcessor(context)}

	if err := service.RegisterHandlers(
		// service.GenerateRequest, service.CastRequest,
		// service.ShuffleRequest, service.FetchRequest,
		// service.DecryptionRequest, service.GenerateElection,
		service.CastBallot, service.GetBallots); err != nil {
		log.ErrFatal(err)
	}
	service.RegisterProcessorFunc(network.MessageType(synchronizer{}), service.synchronize)
	if err := service.load(); err != nil {
		log.Error(err)
	}

	return service
}

///////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////

func (s *Service) GenerateElection(req *api.GenerateElection) (
	*api.GenerateElectionResponse, onet.ClientError) {

	election := req.Election

	size := len(election.Roster.List)
	tree := election.Roster.GenerateNaryTreeWithRoot(size, s.ServerIdentity())
	instance, err := s.CreateProtocol(dkg.NameDKG, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	protocol := instance.(*dkg.SetupDKG)
	protocol.Wait = true
	_ = protocol.Start()

	select {
	case <-protocol.Done:
		shared, _ := protocol.SharedSecret()
		election.Key = shared.X

		client := skipchain.NewClient()
		genesis, err := client.CreateGenesis(election.Roster, 1, 1,
			skipchain.VerificationNone, &election, nil)
		if err != nil {
			return nil, onet.NewClientError(err)
		}

		s.Storage.Chains[election.ID] = &storage.Chain{Genesis: genesis}
		s.save()

		s.propagate(election.Roster.List, &synchronizer{election.ID, genesis})

		return &api.GenerateElectionResponse{shared.X}, nil
	case <-time.After(2 * time.Second):
		return nil, onet.NewClientError(errors.New("DKG timeout"))
	}
}

func (s *Service) CastBallot(req *api.CastBallot) (*api.CastBallotResponse, onet.ClientError) {
	chain, found := s.Storage.Chains[req.ID]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
	}

	index, err := chain.Store(&req.Ballot)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.CastBallotResponse{uint32(index)}, nil
}

func (s *Service) GetBallots(req *api.GetBallots) (*api.GetBallotsResponse, onet.ClientError) {
	chain, found := s.Storage.Chains[req.ID]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
	}

	ballots, err := chain.Ballots()
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	return &api.GetBallotsResponse{ballots}, nil
}

func (s *Service) Shuffle(req *api.Shuffle) (*api.ShuffleReply, onet.ClientError) {
	return nil, nil
}
