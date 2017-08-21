package api_test

import (
	"testing"

	_ "github.com/dedis/cothority_template/service"
	"gopkg.in/dedis/onet.v1/log"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}
