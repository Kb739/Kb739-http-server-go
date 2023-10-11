package main

import (
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

type Req struct {
	path string
}
type HandleFunc func(net.Conn, Req)

var routes map[string]HandleFunc
var absRoutes map[string]HandleFunc

func parseReq(buffer []byte) Req {
	//assume only one slash in between
	startLine := strings.Split(string(buffer), "\r\n")[0]
	path := strings.Split(startLine, " ")[1]
	r := []rune(path)
	if path != "/" && r[len(r)-1] == '/' {
		path = path[:len(path)-1]
	}
	return Req{path}
}

func handleBase(conn net.Conn, req Req) {
	//handle it separately to reduce mess
	fn, _ := routes["*"]
	if fn != nil {
		fn(conn, req)
	} else {
		_, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 2048)
	if _, err := conn.Read(buffer); err != nil {
		log.Fatal(err.Error())
	}
	req := parseReq(buffer)
	if req.path == "/" {
		handleBase(conn, req)
	} else {
		path := "*" + req.path
		arr := strings.Split(path, "/")
		reqRoute := path
		fn, exists := absRoutes[reqRoute]
		if exists == false {
			for i, l := len(arr)-1, 0; i >= 0; i-- {
				fn, exists = routes[reqRoute]
				if exists == true || i == 0 {
					break
				}
				l = l + len(arr[i]) + 1 //extra 1 to emit '/'
				e := len(path) - l
				reqRoute = path[:e]
			}
		}
		if fn != nil {
			fn(conn, req)
		} else {
			_, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}

func handleFunc(route string, fn HandleFunc) {
	runes := []rune(route)
	endChar := runes[len(runes)-1]
	if endChar == '/' {
		path := route[:len(runes)-1]
		routes["*"+path] = fn
	} else {
		absRoutes["*"+route] = fn
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//setting up routes
	routes = make(map[string]HandleFunc)
	absRoutes = make(map[string]HandleFunc)
	handleFunc("/", func(conn net.Conn, req Req) {
		arr := strings.Split(req.path, "/")
		response := "HTTP/1.1 200 OK\r\n\r\n"
		if len(arr) > 2 {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		_, err := conn.Write([]byte(response))
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	handleFunc("/echo/", func(conn net.Conn, req Req) {
		str := ""
		arr := strings.Split(req.path, "/")
		if len(arr) > 2 {
			str = req.path[6:len(req.path)]
		}
		_, err := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)))
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	// l, err := net.Listen("tcp", "127.0.0.1:3000")

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
