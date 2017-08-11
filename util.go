package main

import (
	"bufio"
	"log"
	"strings"
)

func line(io *bufio.ReadWriter) (string, error) {
	data, err := io.ReadString('\n')
	if err != nil {
		log.Println("Could not read from io")
		return "", err
	}

	return strings.Trim(data, "\n"), nil
}
