package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/crypto"
	"github.com/qantik/nevv/service"

	"github.com/dedis/kyber"
	"github.com/dedis/onet"
	"github.com/dedis/onet/app"
)

func main() {
	argRoster := flag.String("roster", "", "path to group toml file")
	_ = flag.String("key", "", "client-side public key")
	argAdmins := flag.String("admins", "", "list of admin scipers")
	argPin := flag.String("pin", "", "service pin")
	flag.Parse()

	roster, err := parseRoster(*argRoster)
	if err != nil {
		panic(err)
	}

	admins, err := parseAdmins(*argAdmins)
	if err != nil {
		panic(err)
	}

	var client struct {
		*onet.Client
	}

	request := &api.Link{Pin: *argPin, Roster: roster, Admins: admins}
	reply := &api.LinkReply{}
	client.Client = onet.NewClient(crypto.Suite, service.Name)
	if err = client.SendProtobuf(roster.List[0], request, reply); err != nil {
		panic(err)
	}

	fmt.Println("Master ID:", reply.ID)
}

// parseRoster reads a Dedis group toml file a converts it to a cothority roster.
func parseRoster(path string) (*onet.Roster, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	group, err := app.ReadGroupDescToml(file)
	if err != nil {
		return nil, err
	}
	return group.Roster, nil
}

func parseKey(key string) (kyber.Point, error) {
	return nil, nil
}

// parseAdmins converts a string of comma-separated sciper numbers in
// the format sciper1,sciper2,sciper3 to a list of integers.
func parseAdmins(scipers string) ([]uint32, error) {
	if scipers == "" {
		return nil, nil
	}

	admins := make([]uint32, 0)
	for _, admin := range strings.Split(scipers, ",") {
		sciper, err := strconv.Atoi(admin)
		if err != nil {
			return nil, err
		}
		admins = append(admins, uint32(sciper))
	}
	return admins, nil
}
