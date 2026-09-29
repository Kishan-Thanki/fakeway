# fakeway

[![Go Reference](https://pkg.go.dev/badge/github.com/kishan-thanki/fakeway)](https://pkg.go.dev/github.com/kishan-thanki/fakeway)
[![CI](https://github.com/Kishan-Thanki/fakeway/actions/workflows/ci.yml/badge.svg)](https://github.com/Kishan-Thanki/fakeway/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/Kishan-Thanki/fakeway?color=blue)](https://github.com/Kishan-Thanki/fakeway/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Kishan-Thanki/fakeway)](https://github.com/Kishan-Thanki/fakeway)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`fakeway` HTTP mock server built with Go 1.22+. Test your server workflows, API clients, retries, and backend workflows.

## Features

- **Zero External Dependencies**: Built entirely using the Go standard library (`net/http`, `http.ServeMux`).
- **Request Reflection (`/echo`)**: Mirrors incoming HTTP methods, query params, headers, and body back in JSON.
- **Status Simulation (`/status/{code}`)**: Returns any valid HTTP status code (100–599) with standard text descriptions.
- **Latency Simulator (`/delay/{duration}`)**: Simulates microservice delays with Go `context` cancellation (aborts timers on early client disconnects).
- **Built-in Middlewares**:
  - **CORS**: Allows seamless browser/frontend cross-origin requests (`*`).
  - **Logging**: Outputs structured request logs with status codes and elapsed latency.
- **Production Server Ready**: Includes configurable timeouts and graceful shutdown via OS signals (`SIGINT`, `SIGTERM`).

## Quick Start

### Prerequisites

- **Go 1.22** or higher installed.

### Run Locally

```bash
# Clone the repository
git clone https://github.com/Kishan-Thanki/fakeway.git
cd fakeway

# Run tests
go test -v ./...

# Start the server (runs on http://localhost:8080)
go run .
```
