package service

import (
	"testing"

	"github.com/qantik/nevv/api"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestService_ClockRequest(t *testing.T) {
	local := onet.NewTCPTest()
	// generate 5 hosts, they don't connect, they process messages, and they
	// don't register the tree or entitylist
	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	gr := &api.GenerateRequest{Name: "test", Roster: roster}

	services := local.GetServices(hosts, serviceID)
	response, err := services[0].(*Service).GenerateRequest(gr)
	if err != nil {
		log.ErrFatal(err)
	}

	log.Lvl2(response)

	b := &api.Ballot{Alpha: api.Suite.Point().Base(), Beta: api.Suite.Point().Base()}
	cr := &api.CastRequest{Election: "test", Ballot: b}

	re, err := services[1].(*Service).CastRequest(cr)
	if err != nil {
		log.ErrFatal(err)
	}

	log.Lvl2(re, services[2].(*Service).Storage)
}
