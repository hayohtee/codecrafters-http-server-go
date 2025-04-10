package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	log.Println("starting server on", l.Addr().String())

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go func() {
			defer conn.Close()
			if err := handleConn(conn); err != nil {
				log.Println("error handling connection:", err)
			}
		}()
	}
}

func handleConn(conn net.Conn)(conn net.Conn) error {
	scanner := bufio.NewScanner(conn)

	// Read the request line
	if !scanner.Scan() {
		return errors.New("failed to read request line")
	}
	header := scanner.Text()

	// Check if the request line is empty
	if header == "" {
		return errors.New("empty request header")
	}
	requestLine := strings.Split(header, " ")

	// Check if the request line is valid
	if len(requestLine) != 3 {
		return errors.New("invalid request line")
	}

	// Check if the URL is valid
	if requestLine[1] != "/" {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return nil
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	return nil
}
