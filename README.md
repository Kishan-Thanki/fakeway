# fakeway

[![Go Reference](https://pkg.go.dev/badge/github.com/kishan-thanki/fakeway)](https://pkg.go.dev/github.com/kishan-thanki/fakeway)
[![CI](https://github.com/Kishan-Thanki/fakeway/actions/workflows/ci.yml/badge.svg)](https://github.com/Kishan-Thanki/fakeway/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/Kishan-Thanki/fakeway?color=blue)](https://github.com/Kishan-Thanki/fakeway/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Kishan-Thanki/fakeway)](https://github.com/Kishan-Thanki/fakeway)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`fakeway` HTTP mock server built with Go 1.22+. Test your server workflows, API clients, retries, and backend workflows.

## Features

- **Zero External Dependencies**: Built entirely using the Go standard library (`net/http`, `http.ServeMux`).
- **Release Verification & Health Checks**:
  - `GET /version`: Returns live build version (injected via `ldflags`) to verify release cutovers (`{"version":"vX.Y.Z"}`).
  - `GET /health` & `GET /healthz`: Health endpoints for load balancer and Kubernetes liveness/readiness probes (`OK`).
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

### Run with Docker / GHCR
 
```bash
# Pull and run the latest image from GHCR
docker run -d -p 8080:8080 ghcr.io/kishan-thanki/fakeway:latest

# Or build locally
docker build -t fakeway .
docker run -d -p 8080:8080 fakeway
```

### GHCR Package Access & Authentication

If the package is set to **Public**, any server can pull without credentials:
```bash
docker pull ghcr.io/kishan-thanki/fakeway:latest
```

If the package is **Private**, authenticate on each server with a GitHub Personal Access Token (PAT) with `read:packages` scope:
```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u <GITHUB_USERNAME> --password-stdin
```

To make the package public:
1. Go to your GitHub profile or repository -> **Packages**.
2. Select `fakeway`.
3. In **Package settings**, scroll down to **Danger Zone** -> **Change package visibility** and choose **Public**.

