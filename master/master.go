package main

import "gopkg.in/dedis/crypto.v0/abstract"

type Master struct {
	Key    abstract.Point
	Admins []int
}
