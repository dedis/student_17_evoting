package service

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
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
