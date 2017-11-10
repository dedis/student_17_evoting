package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNonce(t *testing.T) {
	n1, n2, n3 := nonce(10), nonce(10), nonce(10)
	assert.Equal(t, 10, len(n1), len(n2), len(n3))
	assert.NotEqual(t, n1, n2, n3)
}

func TestSchedule(t *testing.T) {
	s := state{make(map[string]*user)}
	s.log["u1"] = &user{0, false, 0}
	s.log["u2"] = &user{1, false, 2}
	s.log["u3"] = &user{2, false, 4}

	stop := s.schedule(time.Second)
	<-time.After(1000 * time.Millisecond)
	assert.Equal(t, 3, len(s.log))
	<-time.After(1000 * time.Millisecond)
	assert.Equal(t, 2, len(s.log))
	<-time.After(2000 * time.Millisecond)
	assert.Equal(t, 1, len(s.log))
	<-time.After(2000 * time.Millisecond)
	assert.Equal(t, 0, len(s.log))
	stop <- true
}
