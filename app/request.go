package main

import "strings"

type Request struct {
    Method   string
    Route    string
    Params   []string
    Headers  map[string]string
    Protocol string
    Body     any
}

func newRequest(b []byte) *Request {
    request := &Request{
        Headers: make(map[string]string),
    }

    lines := strings.Split(string(b), "\r\n\r\n")
    if len(lines) > 0 {
        headers := strings.Split(lines[0], "\r\n")
        parts := strings.Fields(headers[0])
        request.Method = parts[0]
        params := strings.Split(strings.TrimPrefix(strings.TrimSpace(parts[1]), "/"), "/")
        request.Route = params[0]

        if len(params) > 1 {
            request.Params = params[1:]
        }

        request.Protocol = parts[2]

        for _, item := range headers[1:] {
            headerParts := strings.Split(item, ": ")
            if len(headerParts) == 2 {
                request.Headers[headerParts[0]] = headerParts[1]
            }
        }
    }
    return request
}
