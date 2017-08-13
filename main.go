package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/qantik/mikser/app"
)

func main() {
	host := flag.String("host", "", "host [address:port]")
	silent := flag.Bool("silent", false, "disable log output")
	flag.Parse()

	log.SetFlags(log.Lshortfile)
	log.SetPrefix(*host + " > ")
	if *silent {
		log.SetOutput(ioutil.Discard)
	}

	server, err := app.CreateServer(*host)
	if err != nil {
		os.Exit(1)
	}

	server.Listen()
}
