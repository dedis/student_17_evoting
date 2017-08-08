package net

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func readLine(io *bufio.ReadWriter) (string, error) {
	data, err := io.ReadString('\n')
	if err != nil {
		log.Println("Could not read from io")
		return "", err
	}

	return strings.Trim(data, "\n"), nil
}

func readBytes(io *bufio.ReadWriter, n int) ([]byte, error) {
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		byte, err := io.ReadByte()
		if err != nil {
			log.Println("Could not read byte from io")
			return nil, err
		}
		bytes[i] = byte
	}

	return bytes, nil
}

func sendShort(io *bufio.ReadWriter, message, remote string) {
	if _, err := io.WriteString(message + "\n"); err != nil {
		log.Println("Could not write", message, "to", remote)
		return
	}

	if io.Flush() != nil {
		log.Println("Could not flush", message, "to", remote)
		return
	}

	log.Println("Sent", message, "to", remote)
}

func shut(connection net.Conn) {
	if err := connection.Close(); err != nil {
		log.Println(err)
	}
}
