package service

import (
	"math/rand"
	"time"
)

type user struct {
	sciper int
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

// nonce returns a random string for a given length n.
func nonce(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"

	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = chars[rand.Intn(len(chars))]
	}
	return string(bytes)
}
