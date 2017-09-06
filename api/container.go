package api

import (
	"encoding/hex"
	"fmt"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/ed25519"
)

// Point designates the ed25519 curve point with its coordinates
// represented in raw bytes.
type Point struct {
	X, Y, Z []byte
}

// Pack converts raw byte coordinates to ed25519 curve point.
func (point *Point) Pack() abstract.Point {
	element := Suite.Point().(ed25519.Internal)
	element.Place(point.X, point.Y, point.Z)

	return element
}

// Unpack splits a ed25519 curve point into its raw byte coordinates.
func (point *Point) Unpack(element abstract.Point) {
	convert := element.(ed25519.Internal)
	point.X = convert.GetX()
	point.Y = convert.GetY()
	point.Z = convert.GetZ()
}

func (point *Point) UnpackNorm(element abstract.Point) {
	convert := element.(ed25519.Internal)
	point.X = convert.GetNormX()
	point.Y = convert.GetNormY()
	point.Z = convert.GetNormZ()
}

func (point *Point) Out() {
	fmt.Println("x", hex.EncodeToString(point.X))
	fmt.Println("y", hex.EncodeToString(point.Y))
	fmt.Println("z", hex.EncodeToString(point.Z))
}

// Ballot consists of an ElGamal key pair that is created by the frontend
// and and stored in an individual block on the SkipChain.
type Ballot struct {
	Alpha Point
	Beta  Point
}

// Box wraps a list of Ballots for easier bulk storage on the SkipChain after
// for example a shuffle procedure.
type Box struct {
	Ballots []Ballot
}

// Join is a helper method for joining two list of ElGamal pairs to a single
// list of ballots.
func (box *Box) Join(alpha []abstract.Point, beta []abstract.Point) {
	box.Ballots = make([]Ballot, len(alpha))
	for index := 0; index < len(alpha); index++ {
		gamma, delta := Point{}, Point{}
		gamma.Unpack(alpha[index])
		delta.Unpack(beta[index])
		box.Ballots[index] = Ballot{gamma, delta}
	}
}

// Split is a helper method to create two separate lists of ElGamal pairs from
// a single list.
func (box Box) Split() (alpha, beta []abstract.Point) {
	length := len(box.Ballots)
	alpha = make([]abstract.Point, length)
	beta = make([]abstract.Point, length)

	for index, ballot := range box.Ballots {
		alpha[index] = ballot.Alpha.Pack()
		beta[index] = ballot.Beta.Pack()
	}

	return
}
