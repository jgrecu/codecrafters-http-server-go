package main

import (
    "log"
    "net"
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

    conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
