# Fakeway

[![CI](https://github.com/Kishan-Thanki/fakeway/actions/workflows/ci.yml/badge.svg)](https://github.com/Kishan-Thanki/fakeway/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/Kishan-Thanki/fakeway?color=blue)](https://github.com/Kishan-Thanki/fakeway/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Kishan-Thanki/fakeway)](https://github.com/Kishan-Thanki/fakeway)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Fakeway** is a lightweight HTTP mock server written in Go for testing API clients, retries, timeouts, and backend workflows without depending on a real backend.

## Documentation

For detailed endpoint documentation, HTTP status behavior, request examples, and configuration details, see the [Fakeway Guide](docs/GUIDE.md).

## Features

- **HTTP Status Simulation:** Test HTTP responses, including success, redirect, client-error, and server-error scenarios, with status-specific behavior where appropriate.
- **Request Echo:** Inspect HTTP methods, URLs, query parameters, headers, and request bodies through JSON responses.
- **Latency Simulation:** Introduce request delays to test client timeouts and request cancellation.
- **Health Checks:** `GET /health` and `GET /healthz` for service health checks and probes.
- **Version Endpoint:** `GET /version` returns the running build version.
- **CORS Support:** Handle cross-origin requests from browser-based clients.
- **Request Logging:** Log HTTP methods, paths, status codes, and request durations.
- **Graceful Shutdown:** Handle termination signals and allow in-flight requests time to finish.
- **Configurable Server:** Configure the listening address and HTTP read, write, and idle timeouts.
- **No External Dependencies:** Built using the Go standard library.

## Quick Start

### Prerequisites

- Go 1.22 or later
- Docker (optional)

### Install with Go

Install the latest version of Fakeway:

```bash
go install github.com/kishan-thanki/fakeway@latest
```

Start the server:

```bash
fakeway
```

Fakeway listens on `http://localhost:8080` by default.

Ensure your Go binary directory is in your `PATH` so the `fakeway` command is available.

### Run with Docker

Pull and run the published image:

```bash
docker run --rm -p 8080:8080 ghcr.io/kishan-thanki/fakeway:latest
```

Alternatively, build and run the image locally:

```bash
git clone https://github.com/Kishan-Thanki/fakeway.git
cd fakeway

docker build -t fakeway .
docker run --rm -p 8080:8080 fakeway
```

### Try It

Check the server's health:

```bash
curl http://localhost:8080/health
```

Inspect a request:

```bash
curl "http://localhost:8080/echo?name=fakeway"
```

Simulate an error response:

```bash
curl -i http://localhost:8080/status/500
```

Simulate a delayed response:

```bash
curl http://localhost:8080/delay/2s
```

### Configuration

Fakeway accepts the following command-line flags:

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:8080` | HTTP server address |
| `-read-timeout` | `10s` | Maximum request read duration |
| `-write-timeout` | `35s` | Maximum response write duration |
| `-idle-timeout` | `120s` | Maximum idle connection duration |

Example:

```bash
fakeway -addr :9090 -read-timeout 5s
```

When running from the source repository, you can use the same flags with `go run .`:

```bash
go run . -addr :9090 -read-timeout 5s
```

## Run from Source

To contribute to Fakeway or run the latest source code:

```bash
git clone https://github.com/Kishan-Thanki/fakeway.git
cd fakeway

go test ./...
go run .
```

## License

Fakeway is licensed under the [MIT License](LICENSE).
