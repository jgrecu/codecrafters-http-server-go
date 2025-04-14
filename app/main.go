package main

import (
    "fmt"
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
    n, err := conn.Read(buffer)
    if err != nil {
        log.Fatalln("Error reading request data: ", err.Error())
    }

    message := string(buffer[:n])
    requestTarget := strings.Split(message, " ")[1]

    response := ""
    if requestTarget == "/" || requestTarget == "/index.html" {
        response = "HTTP/1.1 200 OK\r\n\r\n"
    } else if strings.HasPrefix(requestTarget, "/echo/") {
        text := requestTarget[len("/echo/"):]
        response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(text), text)
    } else {
        response = "HTTP/1.1 404 Not Found\r\n\r\n"
    }

    conn.Write([]byte(response))
}
