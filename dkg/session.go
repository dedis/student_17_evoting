package dkg

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/dedis/kyber/share/dkg"
)

type Session struct {
	generator    *dkg.DistKeyGenerator
	participants []int
	responses    []*dkg.Response
	queue        *queue.Queue
}
