## Testing

This project includes a comprehensive suite of unit tests to ensure correctness and robustness.

### Running Tests

To run all tests:

```sh
go test ./app/...
```

### Checking Test Coverage

To check your test coverage (percentage of code exercised by tests):

```sh
go test -cover ./app/...
```

You can also generate a detailed coverage report:

```sh
go test -coverprofile=coverage.out ./app/...
go tool cover -html=coverage.out
```

### What’s Covered

- HTTP status code handling
- HTTP response formatting
- All main endpoints (`/`, `/echo`, `/user-agent`, `/files`)
- Gzip encoding for `/echo`
- File read/write and error handling
- HTTP request parsing, including malformed requests

> **Current coverage:** Over 45% and improving!  
> Contributions to increase coverage are welcome.
