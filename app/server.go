package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

type Req struct {
	method  string
	url     string
	headers map[string]string
	body    string
}

type HandleFunc func(net.Conn, Req)

var routes map[string]HandleFunc
var absRoutes map[string]HandleFunc

func parseReq(buffer []byte) (req Req) {
	req.headers = make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(buffer)))
	if scanner.Scan() {
		requestLine := scanner.Text()
		sections := strings.Split(requestLine, " ")
		req.method = sections[0]
		req.url = sections[1]
		r := []rune(req.url)
		if req.url != "/" && r[len(r)-1] == '/' {
			req.url = req.url[:len(req.url)-1]
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		pair := strings.SplitN(line, ": ", 2)
		req.headers[pair[0]] = pair[1]

	}
	reqSlice := strings.Split(string(buffer), "\r\n")
	req.body = reqSlice[len(reqSlice)-1]
	return req
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
	defer conn.Close()
	buffer := make([]byte, 4096)
	if _, err := conn.Read(buffer); err != nil {
		log.Fatal(err.Error())
	}
	req := parseReq(buffer)
	if req.url == "/" {
		handleBase(conn, req)
	} else {
		path := "*" + req.url
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

	dir := flag.String("directory", "", "directory to server file")
	flag.Parse()
	// Uncomment this block to pass the first stage
	//setting up routes
	routes = make(map[string]HandleFunc)
	absRoutes = make(map[string]HandleFunc)

	handleFunc("/", func(conn net.Conn, req Req) {
		arr := strings.Split(req.url, "/")
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
		arr := strings.Split(req.url, "/")
		if len(arr) > 2 {
			str = req.url[6:len(req.url)]
		}
		_, err := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)))
		if err != nil {
			log.Fatal(err.Error())
		}
	})
	handleFunc("/user-agent", func(conn net.Conn, req Req) {
		body := req.headers["User-Agent"]
		_, err := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)))
		if err != nil {
			log.Fatal(err.Error())
		}
	})
	handleFunc("/files/", func(conn net.Conn, req Req) {
		filename := strings.Split(req.url, "/")[2]
		path := filepath.Join(*dir, filename)
		res := "HTTP/1.1 405 Method Not Allowed\r\n\r\n"
		if req.method == "GET" {
			content, err := os.ReadFile(path)
			if err == nil {
				res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", "application/octet-stream", len(content), content)
			} else {
				res = "HTTP/1.1 404 Not Found\r\n\r\n"
			}

		} else if req.method == "POST" {
			content := req.body
			err := os.WriteFile(path, []byte(content), fs.FileMode(os.O_TRUNC))
			fmt.Println(len(content))
			if err == nil {
				res = "HTTP/1.1 201 Created\r\n\r\n"
			} else {
				res = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
			}
		}
		_, err := conn.Write([]byte(res))
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

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
