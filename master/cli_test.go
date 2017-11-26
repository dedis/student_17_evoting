package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRoster(t *testing.T) {
	defer os.Remove("./public.toml")

	group := `[[servers]]
                    Address = "tcp://127.0.0.1:7002"
                    Public = "UCWCfyVUtrawdKZxGX8VRKlmBi9B64tOPQCE0wPpcDA="
                    Description = "Conode_1"
                  [[servers]]
                    Address = "tcp://127.0.0.1:7004"
                    Public = "gQF/C/DbwU3ren2T/BNIA3MDT9hao11AU2b/mV5rf+g="
                    Description = "Conode_2"`
	_ = ioutil.WriteFile("./public.toml", []byte(group), 0777)

	roster, _ := parseRoster("./public.toml")
	assert.Equal(t, 2, len(roster.List))

	_, err := parseRoster("./foo.toml")
	assert.NotNil(t, err)

	group = "foo"
	_ = ioutil.WriteFile("./public.toml", []byte(group), 0777)

	_, err = parseRoster("./public.toml")
	assert.NotNil(t, err)
}

func TestParseAdmins(t *testing.T) {
	scipers := "123456,654321"
	admins, err := parseAdmins(scipers)
	assert.Equal(t, []uint32{123456, 654321}, admins)

	scipers = "123456,654a21"
	_, err = parseAdmins(scipers)
	assert.NotNil(t, err)
}
