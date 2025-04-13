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

    _, err = l.Accept()
    if err != nil {
        log.Fatalln("Error accepting connection: ", err.Error())
    }
}
