package main

import (
    "log"
    "net"
    "strings"
)

func main() {
    l, err := net.Listen("tcp", "0.0.0.0:4221")
    if err != nil {
        log.Fatalln("Failed to bind to port 4221")
    }

    conn, err := l.Accept()
    if err != nil {
        log.Fatalln("Error accepting connection: ", err.Error())
    }

    defer conn.Close()

    buffer := make([]byte, 1024)
    conn.Read(buffer)

    request := strings.Split(string(buffer), "\r\n")
    line := strings.Split(request[0], " ")
    //method := line[0]
    requestTarget := line[1]
    //protocol := line[2]

    //fmt.Printf("Method: %v\nTarget: %v\nProtocol: %v\n", method, requestTarget, protocol)
    if requestTarget == "/" || requestTarget == "/index.html" {
        conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
    } else {
        conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
    }
}
