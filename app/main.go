package main

import (
	"fmt"
	"log"
	"net"
	"os"
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
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n\r\n"))
	return err
}
