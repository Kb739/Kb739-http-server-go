package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

type Req struct {
	path string
}

func parseReq(buffer []byte) Req {
	startLine := strings.Split(string(buffer), "\r\n")[0]
	path := strings.Split(startLine, " ")[1]
	return Req{path}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	if _, err := conn.Read(buffer); err != nil {
		log.Fatal(err.Error())
	}
	req := parseReq(buffer)
	match := regexp.MustCompile(`/.*`).FindString(req.path)
	response := "HTTP/1.1 200 OK\r\n\r\n"
	if match == "" {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	handleConnection(conn)
}
