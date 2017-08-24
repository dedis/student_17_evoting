package service

import (
	"errors"
	"sync"
	"time"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/protocol"
	"github.com/qantik/nevv/shuffle"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

var templateID onet.ServiceID

func init() {
	templateID, _ = onet.RegisterNewService(api.ServiceName, newService)
	network.RegisterMessage(&storage{})
	network.RegisterMessage(&Config{})
}

type Service struct {
	*onet.ServiceProcessor

	storage *storage
}

// storageID reflects the data we're storing - we could store more
// than one structure.
const storageID = "main"

type storage struct {
	sync.Mutex

	Elections map[string]*Election
}

type Election struct {
	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock

	*protocol.SharedSecret
}

type Config struct {
	Name    string
	Genesis *skipchain.SkipBlock
}

// GenerateRequest ...
func (service *Service) GenerateRequest(request *api.GenerateRequest) (
	*api.GenerateResponse, onet.ClientError) {

	length := len(request.Roster.List)
	tree := request.Roster.GenerateNaryTreeWithRoot(length, service.ServerIdentity())
	dkg, err := service.CreateProtocol(protocol.NameDKG, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	client := skipchain.NewClient()
	genesis, err := client.CreateGenesis(request.Roster, 1, 1,
		skipchain.VerificationNone, nil, nil)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	config, _ := network.Marshal(&Config{Name: request.Name, Genesis: genesis})
	setupDKG := dkg.(*protocol.SetupDKG)
	setupDKG.Wait = true
	if err = setupDKG.SetConfig(&onet.GenericConfig{Data: config}); err != nil {
		return nil, onet.NewClientError(err)
	}

	if err := dkg.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	select {
	case <-setupDKG.Done:
		shared, _ := setupDKG.SharedSecret()
		service.storage.Lock()
		service.storage.Elections[request.Name] = &Election{genesis, genesis, shared}
		service.storage.Unlock()
		service.save()

		return &api.GenerateResponse{Key: shared.X, Hash: genesis.Hash}, nil
	case <-time.After(2000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("dkg didn't finish in time"))
	}
}

func (service *Service) NewProtocol(node *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {
	switch node.ProtocolName() {
	case protocol.NameDKG:
		dkg, err := protocol.NewSetupDKG(node)
		if err != nil {
			return nil, err
		}

		setupDKG := dkg.(*protocol.SetupDKG)
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

			config := data.(*Config)

			service.storage.Lock()
			election := &Election{config.Genesis, config.Genesis, shared}
			service.storage.Elections[config.Name] = election
			service.storage.Unlock()
			service.save()
		}(conf)

		return dkg, nil
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

	election, found := service.storage.Elections[request.Election]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
	}

	client := skipchain.NewClient()
	response, err := client.StoreSkipBlock(election.Latest, nil, request.Ballot)
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	log.Lvl3(service.ServerIdentity(), "Stored ballot at", response.Latest.Index)

	service.storage.Lock()
	election.Latest = response.Latest
	service.storage.Unlock()
	service.save()

	return &api.CastResponse{}, nil
}

func (service *Service) ShuffleRequest(request *api.ShuffleRequest) (
	*api.ShuffleResponse, onet.ClientError) {

	election, found := service.storage.Elections[request.Election]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
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
		service.storage.Lock()
		election.Latest = shuffle.Latest
		service.storage.Unlock()
		service.save()

		return &api.ShuffleResponse{}, nil
	case <-time.After(5000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("Shuffle timeout"))
	}
}

func (service *Service) FetchRequest(request *api.FetchRequest) (
	*api.FetchResponse, onet.ClientError) {

	election, found := service.storage.Elections[request.Election]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
	}

	client := skipchain.NewClient()
	block, err := client.GetSingleBlockByIndex(election.Genesis.Roster,
		election.Genesis.Hash, int(request.Block))
	if err != nil {
		return nil, err
	}

	_, blob, _ := network.Unmarshal(block.Data)
	if err != nil {
		return nil, err
	}

	collection := blob.(*api.Collection)

	return &api.FetchResponse{Ballots: collection.Ballots}, nil
}

func (s *Service) save() {
	s.storage.Lock()
	defer s.storage.Unlock()
	err := s.Save(storageID, s.storage)
	if err != nil {
		log.Error("Couldn't save file:", err)
	}
}

// Tries to load the configuration and updates the data in the service
// if it finds a valid config-file.
func (s *Service) tryLoad() error {
	s.storage = &storage{Elections: make(map[string]*Election)}
	if !s.DataAvailable(storageID) {
		return nil
	}

	msg, err := s.Load(storageID)
	if err != nil {
		return err
	}
	var ok bool
	s.storage, ok = msg.(*storage)
	if !ok {
		return errors.New("Data of wrong type")
	}
	return nil
}

func newService(c *onet.Context) onet.Service {
	s := &Service{ServiceProcessor: onet.NewServiceProcessor(c)}

	if err := s.RegisterHandlers(s.GenerateRequest, s.CastRequest,
		s.ShuffleRequest, s.FetchRequest); err != nil {
		log.ErrFatal(err, "Couldn't register messages")
	}

	if err := s.tryLoad(); err != nil {
		log.Error(err)
	}

	return s
}
