package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type HTTPRequest struct {
	Headers map[string]string
	Url     string
	Method  string
	Body    []byte
}

type HTTPResponse struct {
	Headers map[string]string
	Code    int
	Body    []byte
}

const (
	StatusOK       = 200
	StatusCreated  = 201
	StatusNotFound = 404
)

func StatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusNotFound:
		return "Not Found"
	}
	return ""
}

func (resp HTTPResponse) Write() []byte {
	str := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.Code, StatusText(resp.Code))

	for header, value := range resp.Headers {
		str += fmt.Sprintf("%s: %s\r\n", header, value)
	}

	if len(resp.Body) > 0 {
		str += fmt.Sprintf("Content-Length: %d\r\n", len(resp.Body))
	}

	str += "\r\n"

	if len(resp.Body) > 0 {
		str += string(resp.Body)
	}

	return []byte(str)
}

var tempDirectory string

func main() {
	log.Println("Logs from your program will appear here!")

	if len(os.Args) > 2 {
		tempDirectory = os.Args[2]
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatalln("Failed to bind to port 4221")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("Error accepting connection: ", err.Error())
		}

		go listenRequest(conn)
	}
}

func listenRequest(conn net.Conn) {
	defer conn.Close()

	read := bufio.NewReader(conn)
	buffer := make([]byte, 4096)

	n, err := read.Read(buffer)
	if err != nil {
		log.Fatalln("Error reading request data: ", err.Error())
	}

	rawRequest := buffer[:n]
	parts := strings.Split(string(rawRequest), "\r\n\r\n")
	metaParts := strings.Split(parts[0], "\r\n")
	requestLineParts := strings.Split(metaParts[0], " ")
	headers := make(map[string]string)

	for i := 1; i < len(metaParts); i++ {
		headerParts := strings.Split(metaParts[i], ": ")
		if len(headerParts) >= 2 {
			headers[headerParts[0]] = strings.Join(headerParts[1:], "")
		}
	}

	contentLength, _ := strconv.Atoi(headers["Content-Length"])

	request := HTTPRequest{
		Url:     requestLineParts[1],
		Headers: headers,
		Method:  requestLineParts[0],
		Body:    []byte(parts[1][:contentLength]),
	}

	response := HTTPResponse{
		Code: StatusNotFound,
	}

	if request.Url == "/" {
		response.Code = StatusOK
	} else if strings.HasPrefix(request.Url, "/echo") {
		uriParts := strings.Split(request.Url, "/")
		if len(uriParts) <= 3 {
			content := uriParts[2]

			foundEncoding := false
			if encodingStr, ok := request.Headers["Accept-Encoding"]; ok {
				encodings := strings.SplitSeq(encodingStr, ", ")
				for encoding := range encodings {
					if encoding == "gzip" {
						var encodedContent bytes.Buffer
						gz := gzip.NewWriter(&encodedContent)
						if _, err := gz.Write([]byte(content)); err != nil {
							log.Fatal(err)
						}
						gz.Close()

						response.Code = StatusOK
						response.Headers = map[string]string{"Content-Type": "text/plain", "Content-Encoding": encoding}
						response.Body = encodedContent.Bytes()
						foundEncoding = true
						break
					}
				}
			}

			if !foundEncoding {
				response.Code = StatusOK
				response.Headers = map[string]string{"Content-Type": "text/plain"}
				response.Body = []byte(content)
			}
		}
	} else if strings.HasPrefix(request.Url, "/user-agent") {
		content := request.Headers["User-Agent"]
		response.Code = StatusOK
		response.Headers = map[string]string{"Content-Type": "text/plain"}
		response.Body = []byte(content)
	} else if strings.HasPrefix(request.Url, "/files") {
		uriParts := strings.Split(request.Url, "/")
		if len(uriParts) <= 3 {
			path := uriParts[2]

			if request.Method == "GET" {
				if _, err := os.Stat(fmt.Sprintf("/%s/%s", tempDirectory, path)); errors.Is(err, os.ErrNotExist) {
					response.Code = StatusNotFound
				} else {
					content, _ := os.ReadFile(fmt.Sprintf("/%s/%s", tempDirectory, path))
					response.Code = StatusOK
					response.Headers = map[string]string{"Content-Type": "application/octet-stream"}
					response.Body = content
				}
			} else if request.Method == "POST" {
				os.WriteFile(fmt.Sprintf("/%s/%s", tempDirectory, path), request.Body, 0666)
				response.Code = StatusCreated
			}
		}
	}
	conn.Write(response.Write())
}
