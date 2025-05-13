[![progress-banner](https://backend.codecrafters.io/progress/http-server/a55159a9-f1f1-491f-a2a4-38a2d6c14f5a)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a Go implementation to the
["Build Your Own HTTP server" Challenge](https://app.codecrafters.io/courses/http-server/overview).

[HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) is the
protocol that powers the web. In this challenge, you'll build a HTTP/1.1 server
that is capable of serving multiple clients.

Along the way you'll learn about TCP servers,
[HTTP request syntax](https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html),
and more.

**Note**: If you're viewing this repo on GitHub, head over to
[codecrafters.io](https://codecrafters.io) to try the challenge.

## Testing & Coverage

See [TESTING.md](./TESTING.md) for details on running tests, checking coverage, and what is covered by the test suite.

- To run all tests: `go test ./app/...`
- To check coverage: `go test -cover ./app/...`
- To generate a coverage report: `go test -coverprofile=coverage.out ./app/... && go tool cover -html=coverage.out`

> Current coverage: Over 45% and improving!
