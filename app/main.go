package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var app application
	flag.StringVar(&app.directory, "directory", "/tmp/", "The directory to check for files")
	flag.Parse()

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
			if err := app.handleConn(conn); err != nil {
				log.Println("error handling connection:", err)
			}
		}()
	}
}

type application struct {
	directory string
}

func (app application) handleConn(conn net.Conn) error {
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
	case strings.HasPrefix(path, "/files/"):
		if err := app.fileHandler(conn, path); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	case strings.HasPrefix(path, "/echo/"):
		if err := app.echoHandler(conn, path); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	case path == "/user-agent":
		if err := app.userAgentHandler(conn, tp); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	default:
		statusNotFoundResponse(conn)
	}

	return nil
}

func (app application) userAgentHandler(conn net.Conn, tp *textproto.Reader) error {
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("failed to read MIME header: %w", err)
	}

	userAgent := headers.Get("User-Agent")
	statusOKResponse(conn, "text/plain", []byte(userAgent))
	return nil
}

func (app application) fileHandler(conn net.Conn, urlPath string) error {
	filename := strings.TrimPrefix(urlPath, "/files/")
	path := filepath.Join(app.directory, filename)
	file, err := os.ReadFile(path)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			statusNotFoundResponse(conn)
			return nil
		default:
			return err
		}
	}

	statusOKResponse(conn, "application/octet-stream", file)
	return nil
}

func (app application) echoHandler(conn net.Conn, path string) error {
	word := strings.TrimPrefix(path, "/echo/")
	statusOKResponse(conn, "text/plain", []byte(word))
	return nil
}

func statusNotFoundResponse(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func statusOKResponse(conn net.Conn, contentType string, body []byte) {
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", contentType, len(body), string(body))
	conn.Write([]byte(response))
}
