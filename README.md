# fakeway

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
