package service

import (
	"errors"
	"time"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/dkg"
	"github.com/qantik/nevv/shuffle"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

func init() {
	_, _ = onet.RegisterNewService(api.ID, new)
	for _, message := range []interface{}{
		&Storage{}, &dkg.Config{}, &shuffle.Config{},
	} {
		network.RegisterMessage(message)
	}
}

// Service is the principal application structure holding the onet service processor
// as well as the storage facility for the conode.
type Service struct {
	*onet.ServiceProcessor

	Storage *Storage
}

// GenerateRequest ...
func (service *Service) GenerateRequest(request *api.GenerateRequest) (
	*api.GenerateResponse, onet.ClientError) {

	length := len(request.Roster.List)
	tree := request.Roster.GenerateNaryTreeWithRoot(length, service.ServerIdentity())
	protocol, err := service.CreateProtocol(dkg.NameDKG, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	client := skipchain.NewClient()
	genesis, err := client.CreateGenesis(request.Roster, 1, 1,
		skipchain.VerificationNone, nil, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	config, _ := network.Marshal(&dkg.Config{Name: request.Name, Genesis: genesis})
	setupDKG := protocol.(*dkg.SetupDKG)
	setupDKG.Wait = true
	if err = setupDKG.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
		return nil, onet.NewClientError(err)
	}

	if err := protocol.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-setupDKG.Done:
		shared, _ := setupDKG.SharedSecret()
		service.Storage.createElection(request.Name, genesis, nil, shared)
		service.save()

		return &api.GenerateResponse{Key: shared.X, Hash: genesis.Hash}, nil
	case <-time.After(2000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("DKG timeout"))
	}
}

func (service *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {
	switch node.ProtocolName() {
	case dkg.NameDKG:
		protocol, err := dkg.NewSetupDKG(node)
		if err != nil {
			return nil, err
		}

		setupDKG := protocol.(*dkg.SetupDKG)
		go func(conf *onet.GenericConfig) {
			<-setupDKG.Done
			shared, err := setupDKG.SharedSecret()
			if err != nil {
				return
			}

			_, data, err := network.Unmarshal(conf.Data)
			if err != nil {
				return
			}

			config := data.(*dkg.Config)
			service.Storage.createElection(config.Name, config.Genesis, nil, shared)
			service.save()
		}(conf)

		return protocol, nil
	case shuffle.Name:
		protocol, err := shuffle.New(node)
		if err != nil {
			return nil, err
		}

		shuffle := protocol.(*shuffle.Protocol)
		go func(conf *onet.GenericConfig) {
			<-shuffle.Done
		}(conf)

		return shuffle, nil
	default:
		return nil, errors.New("Unknown protocol")
	}
}

func (service *Service) CastRequest(request *api.CastRequest) (
	*api.CastResponse, onet.ClientError) {

	election, err := service.Storage.get(request.Election)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	client := skipchain.NewClient()
	response, err := client.StoreSkipBlock(election.Latest, nil, request.Ballot)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	log.Lvl3(service.ServerIdentity(), "Stored ballot at", response.Latest.Index)

	service.Storage.updateLatest(request.Election, response.Latest)
	service.save()

	return &api.CastResponse{}, nil
}

func (service *Service) ShuffleRequest(request *api.ShuffleRequest) (
	*api.ShuffleResponse, onet.ClientError) {

	election, err := service.Storage.get(request.Election)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	tree := election.Genesis.Roster.GenerateNaryTreeWithRoot(1, service.ServerIdentity())
	protocol, err := service.CreateProtocol(shuffle.Name, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	shuffle := protocol.(*shuffle.Protocol)
	shuffle.Genesis = election.Genesis
	shuffle.Latest = election.Latest
	if err = shuffle.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-shuffle.Done:
		log.Lvl3("Shuffle done")
		service.Storage.updateLatest(request.Election, shuffle.Latest)
		service.save()

		return &api.ShuffleResponse{}, nil
	case <-time.After(5000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
	}
}

func (service *Service) FetchRequest(request *api.FetchRequest) (
	*api.FetchResponse, onet.ClientError) {

	election, err := service.Storage.get(request.Election)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	client := skipchain.NewClient()
	block, err := client.GetSingleBlockByIndex(election.Genesis.Roster,
		election.Genesis.Hash, int(request.Block))
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	_, blob, err := network.Unmarshal(block.Data)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	box := blob.(*api.Box)

	return &api.FetchResponse{Ballots: box.Ballots}, nil
}

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
	service.Storage = &Storage{Elections: make(map[string]*Election)}
	if !service.DataAvailable(api.ID) {
		return nil
	}

	msg, err := service.Load(api.ID)
	if err != nil {
		return err
	}
	service.Storage = msg.(*Storage)

	return nil
}

// New hooks into the onet registrator to initialize a new service loading
// potential data saved on the disk by an earlier run.
func new(context *onet.Context) onet.Service {
	service := &Service{ServiceProcessor: onet.NewServiceProcessor(context)}

	if err := service.RegisterHandlers(service.GenerateRequest, service.CastRequest,
		service.ShuffleRequest, service.FetchRequest); err != nil {
		log.ErrFatal(err)
	}

	if err := service.load(); err != nil {
		log.Error(err)
	}

	return service
}
