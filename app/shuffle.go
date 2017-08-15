package app

import "github.com/dedis/kyber/abstract"

type Pair struct {
	Alpha abstract.Point
	Beta  abstract.Point
}

type Shuffle struct {
	Index int
	Order []uint32
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
