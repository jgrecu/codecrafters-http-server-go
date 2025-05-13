package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type HTTPRequest struct {
	Headers     map[string]string
	Path        string
	Method      string
	HttpVersion string
	Body        []byte
}

type HTTPResponse struct {
	Headers     map[string]string
	HttpVersion string
	Code        HttpStatus
	Body        []byte
}

type HttpStatus int16

const (
	StatusOK         HttpStatus = 200
	StatusCreated    HttpStatus = 201
	StatusNotFound   HttpStatus = 404
	StatusNotAllowed HttpStatus = 405
)

func (h HttpStatus) String() string {
	switch h {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusNotFound:
		return "Not Found"
	case StatusNotAllowed:
		return "Not Allowed"
	default:
		return ""
	}
}

func (resp HTTPResponse) Write() []byte {
	str := fmt.Sprintf("%s %d %s\r\n", resp.HttpVersion, resp.Code, resp.Code)

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
			log.Fatalf("Error accepting connection: %v", err)
		}

		go handleCon(conn)
	}
}

func handleCon(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		request, closeConn, err := parseRequest(reader)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading request:", err)
			}
			return
		}
		response := handleRequest(request)

		if closeConn {
			response.Headers["Connection"] = "close"
		}

		_, err = conn.Write(response.Write())
		if err != nil {
			return
		}
		if closeConn {
			return
		}
	}
}

func handleRequest(request *HTTPRequest) HTTPResponse {
	paths := strings.Split(request.Path, "/")

	var response HTTPResponse

	response.HttpVersion = request.HttpVersion
	response.Headers = make(map[string]string)
	response.Headers["Content-Type"] = "text/plain"

	switch {
	case paths[1] == "user-agent":
		userAgent := request.Headers["User-Agent"]
		response.Code = StatusOK
		response.Headers["Content-Length"] = strconv.Itoa(len(userAgent))
		response.Body = []byte(userAgent)
	case paths[1] == "echo":
		if request.Method != "GET" {
			response.Code = StatusNotAllowed
			break
		}
		response.Code = StatusOK
		if strings.Contains(request.Headers["Accept-Encoding"], "gzip") {
			var b bytes.Buffer
			enc := gzip.NewWriter(&b)
			enc.Write([]byte(paths[2]))
			enc.Close()
			response.Headers["Content-Encoding"] = "gzip"
			response.Headers["Content-Length"] = strconv.Itoa(len(b.String()))
			response.Body = b.Bytes()
		} else {
			response.Headers["Content-Length"] = strconv.Itoa(len(paths[2]))
			response.Body = []byte(paths[2])
		}
	case paths[1] == "files":
		handleFiles(request, &response)
	case request.Path == "/":
		response.Code = StatusOK
	default:
		response.Code = StatusNotFound
	}
	return response
}

func parseRequest(reader *bufio.Reader) (*HTTPRequest, bool, error) {
	// Parse request information i.e. request method, url, http version
	requestInfo, err := reader.ReadString('\n')
	if err != nil {
		return nil, false, fmt.Errorf("there was an error requesting info: %v", err)
	}

	urlParts := strings.Split(requestInfo, "\r\n")

	parsedRequest := HTTPRequest{}
	headers := make(map[string]string)

	methodPathVersion := strings.Split(urlParts[0], " ")
	if len(methodPathVersion) < 3 {
		return nil, false, fmt.Errorf("invalid request line: %q", urlParts[0])
	}
	parsedRequest.Method = methodPathVersion[0]
	parsedRequest.Path = methodPathVersion[1]
	parsedRequest.HttpVersion = methodPathVersion[2]

	// Parse headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, false, fmt.Errorf("there was an error requesting info: %v", err)
		}

		line = strings.Trim(line, "\n\r")

		// End of headers
		if line == "" {
			break
		}

		name, value, found := strings.Cut(line, ": ")
		if !found {
			return nil, false, fmt.Errorf("wrong header format: %v", err)
		}
		headers[name] = value
	}

	// Check for Connection: close
	closeConn := false
	if val, ok := headers["Connection"]; ok && strings.ToLower(val) == "close" {
		closeConn = true
	}

	parsedRequest.Headers = headers

	// Read the body if there's a Content-Length header
	if contentLength, ok := parsedRequest.Headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLength)
		if err == nil {
			body := make([]byte, length)
			_, err := io.ReadFull(reader, body)
			if err != nil {
				log.Println("Error reading body:", err.Error())
			}
			parsedRequest.Body = body
		}
	}
	if err != nil {
		return nil, false, fmt.Errorf("there was an error parsing body: %v", err)
	}

	return &parsedRequest, closeConn, nil
}

func handleFiles(request *HTTPRequest, response *HTTPResponse) {
	switch request.Method {
	case "GET":
		// dir := os.Args[2]
		fileName := strings.TrimPrefix(request.Path, "/files/")
		fileString := fmt.Sprintf("%s%s", tempDirectory, fileName)
		file, err := os.ReadFile(fileString)
		if err != nil {
			response.Code = StatusNotFound
		} else {
			response.Code = StatusOK
			response.Headers["Content-Length"] = strconv.Itoa(len(file))
			response.Headers["Content-Type"] = "application/octet-stream"
			response.Body = file
		}
	case "POST":
		// dir := os.Args[2]
		fileName := strings.TrimPrefix(request.Path, "/files/")
		fileString := fmt.Sprintf("%s%s", tempDirectory, fileName)
		content := []byte(request.Body)
		err := os.WriteFile(fileString, content, 0644)
		if err != nil {
			log.Println(err)
			response.Code = StatusOK
			break
		} else {
			response.Code = StatusCreated
		}
	default:
		response.Code = StatusNotAllowed
	}
}
