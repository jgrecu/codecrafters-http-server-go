package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestHttpStatusString(t *testing.T) {
	tests := []struct {
		status   HttpStatus
		expected string
	}{
		{StatusOK, "OK"},
		{StatusCreated, "Created"},
		{StatusNotFound, "Not Found"},
		{StatusNotAllowed, "Not Allowed"},
		{HttpStatus(999), ""},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("HttpStatus(%d).String() = %q, want %q", tt.status, got, tt.expected)
		}
	}
}

func TestHTTPResponseWrite(t *testing.T) {
	resp := HTTPResponse{
		Headers:     map[string]string{"X-Test": "true"},
		HttpVersion: "HTTP/1.1",
		Code:        StatusOK,
		Body:        []byte("hello"),
	}
	out := resp.Write()
	s := string(out)
	if !bytes.Contains(out, []byte("HTTP/1.1 200 OK")) {
		t.Errorf("Response missing status line: %q", s)
	}
	if !bytes.Contains(out, []byte("X-Test: true")) {
		t.Errorf("Response missing header: %q", s)
	}
	if !bytes.Contains(out, []byte("hello")) {
		t.Errorf("Response missing body: %q", s)
	}
}

func TestHTTPResponseWriteNoBody(t *testing.T) {
	resp := HTTPResponse{
		Headers:     map[string]string{"X-Test": "true"},
		HttpVersion: "HTTP/1.1",
		Code:        StatusCreated,
		Body:        nil,
	}
	out := resp.Write()
	s := string(out)
	if !strings.Contains(s, "201 Created") {
		t.Errorf("Expected status line for 201 Created, got %q", s)
	}
	if !strings.Contains(s, "X-Test: true") {
		t.Errorf("Expected header, got %q", s)
	}
	if strings.Contains(s, "Content-Length:") {
		t.Errorf("Should not have Content-Length for nil body, got %q", s)
	}
}

func TestHandleRequestEcho(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{},
		Path:        "/echo/foobar",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusOK {
		t.Errorf("Expected StatusOK, got %v", resp.Code)
	}
	if string(resp.Body) != "foobar" {
		t.Errorf("Expected body 'foobar', got %q", string(resp.Body))
	}
	if resp.Headers["Content-Length"] != strconv.Itoa(len("foobar")) {
		t.Errorf("Expected Content-Length %d, got %s", len("foobar"), resp.Headers["Content-Length"])
	}
}

func TestHandleRequestUserAgent(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{"User-Agent": "GoTest"},
		Path:        "/user-agent",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusOK {
		t.Errorf("Expected StatusOK, got %v", resp.Code)
	}
	if string(resp.Body) != "GoTest" {
		t.Errorf("Expected body 'GoTest', got %q", string(resp.Body))
	}
}

func TestHandleRequestNotFound(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{},
		Path:        "/notfound",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusNotFound {
		t.Errorf("Expected StatusNotFound, got %v", resp.Code)
	}
}

func TestHandleRequestRoot(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{},
		Path:        "/",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusOK {
		t.Errorf("Expected StatusOK for root, got %v", resp.Code)
	}
}

func TestHandleRequestEchoEmpty(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{},
		Path:        "/echo/",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusOK {
		t.Errorf("Expected StatusOK for empty echo, got %v", resp.Code)
	}
	if string(resp.Body) != "" {
		t.Errorf("Expected empty body for /echo/, got %q", string(resp.Body))
	}
}

func TestHandleRequestMethodNotAllowed(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{},
		Path:        "/echo/foobar",
		Method:      "POST",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusNotAllowed {
		t.Errorf("Expected StatusNotAllowed for POST, got %v", resp.Code)
	}
}

func TestHTTPResponseWriteNoHeaders(t *testing.T) {
	resp := HTTPResponse{
		Headers:     map[string]string{},
		HttpVersion: "HTTP/1.1",
		Code:        StatusOK,
		Body:        []byte("test"),
	}
	out := resp.Write()
	s := string(out)
	if !bytes.Contains(out, []byte("HTTP/1.1 200 OK")) {
		t.Errorf("Response missing status line: %q", s)
	}
	if !bytes.Contains(out, []byte("test")) {
		t.Errorf("Response missing body: %q", s)
	}
}

func TestHttpStatusStringUnknown(t *testing.T) {
	status := HttpStatus(12345)
	if status.String() != "" {
		t.Errorf("Expected empty string for unknown status, got %q", status.String())
	}
}

// Test gzip encoding for /echo endpoint
func TestHandleRequestEchoGzip(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{"Accept-Encoding": "gzip"},
		Path:        "/echo/foobar",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Headers["Content-Encoding"] != "gzip" {
		t.Errorf("Expected gzip encoding")
	}
	gr, err := gzip.NewReader(bytes.NewReader(resp.Body))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	unzipped, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}
	if string(unzipped) != "foobar" {
		t.Errorf("Expected gzip body 'foobar', got %q", string(unzipped))
	}
}

// Test handleFiles GET for non-existent file
func TestHandleFilesGetNotFound(t *testing.T) {
	resp := &HTTPResponse{Headers: map[string]string{}}
	req := &HTTPRequest{
		Method:      "GET",
		Path:        "/files/doesnotexist.txt",
		HttpVersion: "HTTP/1.1",
	}
	tempDirectory = os.TempDir() + "/"
	handleFiles(req, resp)
	if resp.Code != StatusNotFound {
		t.Errorf("Expected StatusNotFound for missing file, got %v", resp.Code)
	}
}

// Test handleFiles POST and GET for file
func TestHandleFilesPostAndGet(t *testing.T) {
	filename := "testfile.txt"
	filepath := os.TempDir() + "/" + filename
	defer os.Remove(filepath)

	// POST (write file)
	resp := &HTTPResponse{Headers: map[string]string{}}
	req := &HTTPRequest{
		Method:      "POST",
		Path:        "/files/" + filename,
		HttpVersion: "HTTP/1.1",
		Body:        []byte("hello world"),
	}
	tempDirectory = os.TempDir() + "/"
	handleFiles(req, resp)
	if resp.Code != StatusCreated {
		t.Errorf("Expected StatusCreated for POST, got %v", resp.Code)
	}

	// GET (read file)
	resp2 := &HTTPResponse{Headers: map[string]string{}}
	req2 := &HTTPRequest{
		Method:      "GET",
		Path:        "/files/" + filename,
		HttpVersion: "HTTP/1.1",
	}
	handleFiles(req2, resp2)
	if resp2.Code != StatusOK {
		t.Errorf("Expected StatusOK for GET, got %v", resp2.Code)
	}
	if string(resp2.Body) != "hello world" {
		t.Errorf("Expected file content 'hello world', got %q", string(resp2.Body))
	}
}

// Test handleFiles with unsupported method
func TestHandleFilesNotAllowed(t *testing.T) {
	resp := &HTTPResponse{Headers: map[string]string{}}
	req := &HTTPRequest{
		Method:      "PUT",
		Path:        "/files/any.txt",
		HttpVersion: "HTTP/1.1",
	}
	handleFiles(req, resp)
	if resp.Code != StatusNotAllowed {
		t.Errorf("Expected StatusNotAllowed for PUT, got %v", resp.Code)
	}
}

func TestHandleFilesPostWriteError(t *testing.T) {
	resp := &HTTPResponse{Headers: map[string]string{}}
	req := &HTTPRequest{
		Method:      "POST",
		Path:        "/files/shouldfail.txt",
		HttpVersion: "HTTP/1.1",
		Body:        []byte("fail"),
	}
	tempDirectory = "/root/shouldnotexist/"
	handleFiles(req, resp)
	if resp.Code != StatusOK {
		t.Errorf("Expected StatusOK for POST error, got %v", resp.Code)
	}
}

func TestParseRequestSimpleGet(t *testing.T) {
	raw := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	req, closeConn, err := parseRequest(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if req.Method != "GET" || req.Path != "/" || req.HttpVersion != "HTTP/1.1" {
		t.Errorf("Parsed request fields incorrect: %+v", req)
	}
	if closeConn {
		t.Errorf("Expected closeConn to be false")
	}
}

func TestParseRequestWithBody(t *testing.T) {
	raw := "POST /echo/abc HTTP/1.1\r\nHost: localhost\r\nContent-Length: 3\r\n\r\nxyz"
	reader := bufio.NewReader(strings.NewReader(raw))
	req, _, err := parseRequest(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if req.Method != "POST" || req.Path != "/echo/abc" {
		t.Errorf("Parsed request fields incorrect: %+v", req)
	}
	if string(req.Body) != "xyz" {
		t.Errorf("Expected body 'xyz', got %q", string(req.Body))
	}
}

func TestParseRequestConnectionClose(t *testing.T) {
	raw := "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	_, closeConn, err := parseRequest(reader)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !closeConn {
		t.Errorf("Expected closeConn to be true")
	}
}

func TestParseRequestMalformedHeader(t *testing.T) {
	raw := "GET / HTTP/1.1\r\nBadHeader\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	_, _, err := parseRequest(reader)
	if err == nil {
		t.Errorf("Expected error for malformed header")
	}
}

func TestHandleRequestUnknownPath(t *testing.T) {
	req := &HTTPRequest{
		Headers:     map[string]string{},
		Path:        "/unknown/path",
		Method:      "GET",
		HttpVersion: "HTTP/1.1",
	}
	resp := handleRequest(req)
	if resp.Code != StatusNotFound {
		t.Errorf("Expected StatusNotFound for unknown path, got %v", resp.Code)
	}
}
