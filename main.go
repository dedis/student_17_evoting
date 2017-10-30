package main

import (
	_ "github.com/dedis/cothority/cosi/service"
	_ "github.com/qantik/nevv/service"
	"gopkg.in/dedis/onet.v1/app"
)

func main() {
	app.Server()
}
