package service

import (
	"testing"
	"time"

	"github.com/qantik/nevv/api"
	"github.com/stretchr/testify/assert"

	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
)

func castServices(services []onet.Service) []*Service {
	cast := make([]*Service, len(services))
	for i, service := range services {
		cast[i] = service.(*Service)
	}

	return cast
}

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestGenerateElection(t *testing.T) {
	local := onet.NewTCPTest()

	hosts, roster, _ := local.GenTree(3, true)
	defer local.CloseAll()

	services := castServices(local.GetServices(hosts, serviceID))

	election := api.Election{"test", "", "", "", []byte{}, roster, []string{}, nil, ""}
	msg := &api.GenerateElection{Token: "", Election: election}

	response, err := services[0].GenerateElection(msg)
	if err != nil {
		log.ErrFatal(err)
	}

	<-time.After(time.Second)

	key1 := services[0].Storage.GetElection("test").Key
	key2 := services[1].Storage.GetElection("test").Key
	key3 := services[2].Storage.GetElection("test").Key

	assert.Equal(t, key1, key2, key3, response.Key)
}
