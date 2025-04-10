package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/textproto"
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

func handleConn(conn net.Conn) error {
	tp := textproto.NewReader(bufio.NewReader(conn))

	// Read the request line
	data, err := tp.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read request line: %w", err)
	}

	// Check if the request line is empty
	if data == "" {
		return errors.New("empty request header")
	}
	requestLine := strings.Split(data, " ")

	// Check if the request line is valid
	if len(requestLine) != 3 {
		return errors.New("invalid request line")
	}

	path := requestLine[1]

	// Handle connection based on url
	switch {
	case path == "/":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n\r\n"))
	case strings.HasPrefix(path, "/echo/"):
		if err := echoHandler(conn, path); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	case path == "/user-agent":
		if err := userAgentHandler(conn); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

	return nil
}

func userAgentHandler(conn net.Conn) error {
	tp := textproto.NewReader(bufio.NewReader(conn))

	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("failed to read MIME header: %w", err)
	}

	return nil
}

func echoHandler(conn net.Conn, path string) error {
	word := strings.TrimPrefix(path, "/echo/")
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(word), word)
	conn.Write([]byte(response))
	return nil
}
