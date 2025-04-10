package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	var directory string
	flag.StringVar(&directory, "directory", "/tmp/", "The directory to check for files")
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
			app := application{
				filesDir: directory,
			}

			if err := app.handleConn(conn); err != nil {
				log.Println("error handling connection:", err)
			}
		}()
	}
}

type application struct {
	filesDir      string
	requestURL    string
	requestMethod string
}

func (app *application) handleConn(conn net.Conn) error {
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
	app.requestMethod = requestLine[0]
	app.requestURL = requestLine[1]

	// Handle connection based on url
	switch {
	case path == "/":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n\r\n"))
	case strings.HasPrefix(data, "GET /files/"):
		if err := app.fileHandler(conn); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	case strings.HasPrefix(data, "POST /files/"):
		if err := app.postFileHandler(conn, tp); err != nil {
			return fmt.Errorf("failed to handle request to %s: %w", path, err)
		}
	case strings.HasPrefix(path, "/echo/"):
		if err := app.echoHandler(conn); err != nil {
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

func (app application) postFileHandler(conn net.Conn, tp *textproto.Reader) error {
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return err
	}

	contentLengthStr := headers.Get("Content-Length")
	contentLength := 0
	if contentLengthStr != "" {
		contentLength, err = strconv.Atoi(contentLengthStr)
		if err != nil {
			return err
		}
	}

	body := make([]byte, contentLength)
	_, err = io.ReadFull(tp.R, body)
	if err != nil {
		return err
	}

	filename := strings.TrimPrefix(app.requestURL, "/files/")
	path := filepath.Join(app.filesDir, filename)

	if err = os.WriteFile(path, body, 0644); err != nil {
		return err
	}
	
	conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
	return nil
}

func (app *application) fileHandler(conn net.Conn) error {
	filename := strings.TrimPrefix(app.requestURL, "/files/")
	path := filepath.Join(app.filesDir, filename)
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

func (app application) echoHandler(conn net.Conn) error {
	word := strings.TrimPrefix(app.requestURL, "/echo/")
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
