package main

import (
    "fmt"
    "log"
    "net"
)

func main() {
    l, err := net.Listen("tcp", "0.0.0.0:4221")
    if err != nil {
        log.Fatalln("Failed to bind to port 4221")
    }

    for {
        conn, err := l.Accept()
        if err != nil {
            log.Fatalln("Error accepting connection: ", err.Error())
        }

        go func(conn net.Conn) {
            defer conn.Close()

            buffer := make([]byte, 1024)
            n, err := conn.Read(buffer)
            if err != nil {
                log.Fatalln("Error reading request data: ", err.Error())
            }

            message := buffer[:n]
            request := newRequest(message)

            response := ""
            if request.Route == "" {
                response = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 0\r\n\r\n"
            } else if request.Route == "echo" && len(request.Params) > 0 {
                response = fmt.Sprintf(
                    "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
                    len(request.Params[0]),
                    request.Params[0],
                )
            } else if request.Route == "user-agent" {
                response = fmt.Sprintf(
                    "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
                    len(request.Headers["User-Agent"]),
                    request.Headers["User-Agent"],
                )
            } else {
                response = "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\nContent-Length: 0\r\n\r\n"
            }

            conn.Write([]byte(response))
        }(conn)
    }
}
