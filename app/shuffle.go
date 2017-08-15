package app

import "github.com/dedis/kyber/abstract"

// Pair designates an ElGamal encryption pair.
type Pair struct {
	Alpha abstract.Point
	Beta  abstract.Point
}

// Shuffle is used for communcation between mixnet nodes and third parties.
type Shuffle struct {
	// Index of the current node within the order list.
	Index int

	// Order specifies the ordering of mixnet nodes during shuffling.
	Order []uint32

	// ElGamal pairs to be shuffled.
	Pairs []Pair
}

func createPairs(alpha, beta []abstract.Point) []Pair {
	pairs := make([]Pair, len(alpha))
	for index := range pairs {
		pairs[index] = Pair{alpha[index], beta[index]}
	}
	return pairs
}

func (shuffle Shuffle) alpha() []abstract.Point {
	points := make([]abstract.Point, len(shuffle.Pairs))
	for index, pair := range shuffle.Pairs {
		points[index] = pair.Alpha
	}
	return points
}

func (shuffle Shuffle) beta() []abstract.Point {
	points := make([]abstract.Point, len(shuffle.Pairs))
	for index, pair := range shuffle.Pairs {
		points[index] = pair.Beta
	}
	return points
}

func (shuffle Shuffle) strings() ([]string, []string) {
	alpha := make([]string, len(shuffle.Pairs))
	beta := make([]string, len(shuffle.Pairs))

	for index := range shuffle.Pairs {
		alpha[index] = shuffle.Pairs[index].Alpha.String()
		beta[index] = shuffle.Pairs[index].Beta.String()
	}

	return alpha, beta
}
