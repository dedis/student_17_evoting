package service

import (
	"math/rand"
	"time"

	"github.com/qantik/nevv/chains"
)

type stamp struct {
	user  chains.User
	admin bool
	time  int
}

type state struct {
	log map[string]*stamp
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// schedule periodically increments the time counter for each user in the
// state log and removes him if the time limit has been reached.
func (s *state) schedule(interval time.Duration) chan bool {
	ticker := time.NewTicker(interval)
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				for nonce, stamp := range s.log {
					if stamp.time == 5 {
						delete(s.log, nonce)
					} else {
						stamp.time++
					}
				}
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()

	return stop
}

// register a new user in the log and return 32 character nonce as a token.
func (s *state) register(user chains.User, admin bool) string {
	token := nonce(32)
	s.log[token] = &stamp{user, admin, 0}
	return token
}

// nonce returns a random string for a given length n.
func nonce(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"

	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = chars[rand.Intn(len(chars))]
	}
	return string(bytes)
}
