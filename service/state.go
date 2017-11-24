package service

import (
	"math/rand"
	"time"
)

type user struct {
	sciper uint32
	admin  bool
	time   int
}

type state struct {
	log map[string]*user
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
				for nonce, user := range s.log {
					if user.time == 5 {
						delete(s.log, nonce)
					} else {
						user.time++
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
func (s *state) register(sciper uint32, admin bool) string {
	token := nonce(32)
	s.log[token] = &user{sciper, admin, 0}
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
