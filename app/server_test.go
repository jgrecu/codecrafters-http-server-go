package main

import (
	"bytes"
	"strconv"
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
