package service

// import (
// 	"github.com/dedis/cothority/skipchain"

// 	"gopkg.in/dedis/crypto.v0/abstract"
// 	"gopkg.in/dedis/onet.v1/network"

// 	"github.com/qantik/nevv/election"
// )

// func init() {
// 	network.RegisterMessage(&master{})
// 	network.RegisterMessage(&link{})
// }

// type master struct {
// 	Key    abstract.Point
// 	Admins []election.User
// }

// type link struct {
// 	Genesis skipchain.SkipBlockID
// }

// func (m *master) isAdmin(user election.User) bool {
// 	for _, admin := range m.Admins {
// 		if admin == user {
// 			return true
// 		}
// 	}
// 	return false
// }
