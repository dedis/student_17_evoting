package service

import (
	"time"

	"errors"
	"sync"

	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/protocol"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

// Used for tests
var templateID onet.ServiceID

func init() {
	var err error
	templateID, err = onet.RegisterNewService(api.ServiceName, newService)
	log.ErrFatal(err)
	network.RegisterMessage(&storage{})
	network.RegisterMessage(&Base{})
	network.RegisterMessage(&Ballot{})
}

// Service is our template-service
type Service struct {
	*onet.ServiceProcessor

	storage *storage
}

// storageID reflects the data we're storing - we could store more
// than one structure.
const storageID = "main"

// storage is used to save our data.
type storage struct {
	Count int
	sync.Mutex

	Elections map[string]*Election
}

type Base struct {
	Key abstract.Point
}

type Election struct {
	Genesis *skipchain.SkipBlock
	Last    *skipchain.SkipBlock
	Indexes []int
}

type Ballot struct {
	Data string
}

// GenerateRequest ...
func (s *Service) GenerateRequest(req *api.GenerateRequest) (
	*api.GenerateResponse, onet.ClientError) {

	tree := req.Roster.GenerateNaryTreeWithRoot(len(req.Roster.List), s.ServerIdentity())
	pi, err := s.CreateProtocol(protocol.NameDKG, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}
	setupDKG := pi.(*protocol.SetupDKG)
	setupDKG.Wait = true
	//setupDKG.SetConfig(&onet.GenericConfig{Data: reply.OCS.Hash}) ???
	if err := pi.Start(); err != nil {
		return nil, onet.NewClientError(err)
	}

	log.Lvl3("Started DKG-protocol - waiting for done")

	select {
	case <-setupDKG.Done:
		shared, err := setupDKG.SharedSecret()
		if err != nil {
			return nil, onet.NewClientError(err)
		}

		base := &Base{Key: shared.X}
		client := skipchain.NewClient()
		genesis, err := client.CreateGenesis(req.Roster, 1, 1,
			skipchain.VerificationNone, base, nil)
		if err != nil {
			return nil, onet.NewClientError(err)
		}

		election := &Election{Genesis: genesis, Last: genesis, Indexes: make([]int, 0)}
		s.storage.Elections[req.Name] = election
		return &api.GenerateResponse{Key: shared.X, Hash: genesis.Hash}, nil
		//s.saveMutex.Lock()
		//s.Storage.Shared[string(reply.OCS.Hash)] = shared
		//s.saveMutex.Unlock()
		//reply.X = shared.X
	case <-time.After(2000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("dkg didn't finish in time"))
	}
}

func (service *Service) CastRequest(request *api.CastRequest) (
	*api.CastResponse, onet.ClientError) {

	election, found := service.storage.Elections[request.Name]
	if !found {
		return nil, onet.NewClientError(errors.New("Election not found"))
	}

	client := skipchain.NewClient()
	response, err := client.StoreSkipBlock(election.Last, nil, []byte(request.Ballot))
	if err != nil {
		return nil, onet.NewClientError(err)
	}

	election.Last = response.Latest
	election.Indexes = append(election.Indexes, response.Latest.Index)
	log.Lvl3("Indexes:", election.Indexes)

	return &api.CastResponse{}, nil
}

func (s *Service) NewProtocol(tn *onet.TreeNodeInstance, conf *onet.GenericConfig) (
	onet.ProtocolInstance, error) {

	log.Lvl3("Not templated yet")
	return nil, nil
}

// saves all skipblocks.
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

// newService receives the context that holds information about the node it's
// running on. Saving and loading can be done using the context. The data will
// be stored in memory for tests and simulations, and on disk for real deployments.
func newService(c *onet.Context) onet.Service {
	s := &Service{
		ServiceProcessor: onet.NewServiceProcessor(c),
	}

	if err := s.RegisterHandlers(s.GenerateRequest, s.CastRequest); err != nil {
		log.ErrFatal(err, "Couldn't register messages")
	}

	if err := s.tryLoad(); err != nil {
		log.Error(err)
	}

	return s
}
