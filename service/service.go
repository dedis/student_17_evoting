package service

/*
The service.go defines what to do for each API-call. This part of the service
runs on the node.
*/

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
}

// Service is our template-service
type Service struct {
	// We need to embed the ServiceProcessor, so that incoming messages
	// are correctly handled.
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

	Chains map[string]*skipchain.SkipBlock
}

type Base struct {
	Key abstract.Point
}

// ClockRequest starts a template-protocol and returns the run-time.
func (s *Service) ClockRequest(req *api.ClockRequest) (
	*api.ClockResponse, onet.ClientError) {

	s.storage.Lock()
	s.storage.Count++
	s.storage.Unlock()
	s.save()

	tree := req.Roster.GenerateNaryTreeWithRoot(2, s.ServerIdentity())
	pi, err := s.CreateProtocol(protocol.Name, tree)
	if err != nil {
		return nil, onet.NewClientError(err)
	}
	start := time.Now()
	_ = pi.Start()
	resp := &api.ClockResponse{
		Children: <-pi.(*protocol.Template).ChildCount,
	}
	resp.Time = time.Now().Sub(start).Seconds()
	return resp, nil
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

		s.storage.Chains[req.Name] = genesis
		return &api.GenerateResponse{Key: shared.X, Hash: genesis.Hash}, nil
		//s.saveMutex.Lock()
		//s.Storage.Shared[string(reply.OCS.Hash)] = shared
		//s.saveMutex.Unlock()
		//reply.X = shared.X
	case <-time.After(2000 * time.Millisecond):
		return nil, onet.NewClientError(errors.New("dkg didn't finish in time"))
	}
}

// CountRequest returns the number of instantiations of the protocol.
func (s *Service) CountRequest(req *api.CountRequest) (
	*api.CountResponse, onet.ClientError) {

	s.storage.Lock()
	defer s.storage.Unlock()
	return &api.CountResponse{Count: s.storage.Count}, nil
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
	s.storage = &storage{Chains: make(map[string]*skipchain.SkipBlock)}
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
	if err := s.RegisterHandlers(s.ClockRequest, s.CountRequest,
		s.GenerateRequest); err != nil {
		log.ErrFatal(err, "Couldn't register messages")
	}
	if err := s.tryLoad(); err != nil {
		log.Error(err)
	}
	return s
}
