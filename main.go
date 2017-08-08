package main

import (
	"log"
	"os"

	"github.com/qantik/mikser/net"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("Specify host address")
	}

	server, err := net.NewServer(os.Args[1], "pool.toml")
	if err != nil {
		os.Exit(1)
	}

	server.Open()
}
