package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/app"
)

func main() {
	argRoster := flag.String("roster", "", "path to group toml file")
	_ = flag.String("key", "", "client-side public key")
	argAdmins := flag.String("admins", "", "list of admin scipers")
	_ = flag.Int("pin", 0, "service pin number")

	flag.Parse()

	file, err := os.Open(*argRoster)
	if err != nil {
		panic(err)
	}
	group, err := app.ReadGroupDescToml(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(group.Roster)

	admins := make([]int, 0)
	for _, admin := range strings.Split(*argAdmins, ",") {
		sciper, _ := strconv.Atoi(admin)
		admins = append(admins, sciper)
	}

}

// parseRoster reads a Dedis group toml file a converts it to
// cothority roster.
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

func parseKey(key string) (abstract.Point, error) {
	return nil, nil
}

// parseAdmins converts a string of comma-separated sciper numbers in
// the format sciper1,sciper2,sciper3 to a list of integers.
func parseAdmins(scipers string) ([]int, error) {
	admins := make([]int, 0)
	for _, admin := range strings.Split(scipers, ",") {
		sciper, err := strconv.Atoi(admin)
		if err != nil {
			return nil, err
		}
		admins = append(admins, sciper)
	}
	return admins, nil
}
