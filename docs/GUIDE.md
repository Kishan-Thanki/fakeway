# Fakeway Guide

Fakeway is a lightweight HTTP mock server for testing how API clients handle HTTP status codes, headers, redirects, latency, and request data without requiring a real backend.

For installation and a quick start, see the [README](../README.md).

## Contents

- [How Fakeway works](#how-fakeway-works)
- [HTTP endpoints](#http-endpoints)
- [Simulating HTTP status codes](#simulating-http-status-codes)
- [Echoing requests](#echoing-requests)
- [Simulating latency](#simulating-latency)
- [Configuration](#configuration)
- [CORS, logging, and shutdown](#cors-logging-and-shutdown)
- [Error responses](#error-responses)

## How Fakeway works

Fakeway starts an HTTP server, listens for requests, and returns predictable mock responses. It is useful when developing or testing an API client that needs to handle conditions such as:

- successful responses and server errors;
- redirects and resource-creation responses;
- empty responses and partial content;
- authentication challenges and retry-related headers;
- slow responses, timeouts, and cancelled requests;
- request methods, query parameters, headers, and bodies.

Fakeway does not need a real application backend. Start it locally, point your client or test at its endpoints, and stop it when the test is complete.

## HTTP endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/health` | `GET` | Returns `200 OK` with `OK` as plain text. |
| `/healthz` | `GET` | Health-check alias for `/health`. |
| `/version` | `GET` | Returns the running build version as JSON. |
| `/echo` | Any method | Returns the request method, URL, query parameters, headers, and body as JSON. |
| `/status/{code}` | Any method | Simulates a selected final HTTP status code from `200` through `599`, subject to the exceptions described below. |
| `/delay/{duration}` | `GET` | Waits for the requested duration and returns timing information. |

The server uses Go's HTTP router. The `/delay/{duration}` route is registered for `GET`; the other routes above are not restricted to a specific method by their route pattern, although some status simulations require particular methods.

### Health check

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/healthz
```

A successful request returns HTTP `200` and the plain-text body:

```text
OK
```

These endpoints can be used by local scripts, container health checks, or service probes.

### Version

```bash
curl http://localhost:8080/version
```

Example response:

```json
{"version":"development"}
```

The default value of the version is `development`. A release build may set `Version` at build time; the endpoint returns whatever value is present in the running binary.

## Simulating HTTP status codes

Use `/status/{code}` to make Fakeway return a chosen HTTP status:

```bash
curl -i http://localhost:8080/status/200
curl -i http://localhost:8080/status/404
curl -i http://localhost:8080/status/429
curl -i http://localhost:8080/status/500
```

Fakeway accepts final status codes from **200 through 599**, except `305` and `306`. It rejects informational `1xx` codes because they cannot be returned as standalone final responses by this endpoint. Codes `305` (deprecated) and `306` (unused) are also rejected. Invalid codes produce `400 Bad Request` with a problem-details JSON response.

For an ordinary successful status, Fakeway returns JSON similar to:

```json
{
  "code": 200,
  "status": "OK"
}
```

For ordinary error statuses (`400` and above), it returns `application/problem+json`. Known status codes use a status-specific detail message where one is defined; other codes use a generic simulated-error message. Some statuses have additional protocol-specific behavior, described below.

### Statuses with special behavior

| Status | Behavior | Requirements or notes |
|---|---|---|
| `201 Created` | Returns JSON and a `Location` header. | Requires `POST` or `PUT`. Defaults to `/resources/1`. |
| `204 No Content` | Sends an empty response body. | No JSON response body is returned. |
| `205 Reset Content` | Sends an empty response body with `Content-Length: 0`. | No JSON response body is returned. |
| `206 Partial Content` | Serves a byte range from the fixed text `Fakeway mock response!`. | Requires `GET` or `HEAD` and a `Range` header. |
| `300 Multiple Choices` | Returns JSON with a `Location` and two alternatives. | Defaults to `/health` and `/version`. |
| `301`, `302`, `303`, `307`, `308` | Sends a redirect with a `Location` header. | Defaults to `/health`. |
| `304 Not Modified` | Sends an empty not-modified response with an `ETag` and `Cache-Control: no-cache`. | Requires `GET` or `HEAD` and a matching `If-None-Match` header. |
| `401 Unauthorized` | Adds `WWW-Authenticate: Basic realm="Fakeway"`. | Returns problem-details JSON. |
| `405 Method Not Allowed` | Adds an `Allow` header. | The header is generated based on the request method. |
| `407 Proxy Authentication Required` | Adds `Proxy-Authenticate: Basic realm="Fakeway"`. | Returns problem-details JSON. |
| `408 Request Timeout` | Adds `Connection: close` for HTTP/1.x requests. | Returns problem-details JSON. |
| `416 Range Not Satisfiable` | Adds a `Content-Range` header describing the representation size. | Requires `GET` or `HEAD` and a valid, unsatisfiable single byte range. |
| `426 Upgrade Required` | Adds `Upgrade: websocket` and, for HTTP/1.x, `Connection: Upgrade`. | Returns problem-details JSON. |
| `429 Too Many Requests` | Adds `Retry-After: 60`. | Returns problem-details JSON. |
| `503 Service Unavailable` | Adds `Retry-After: 60`. | Returns problem-details JSON. |

Other codes use the general response behavior for their status class.

### Created response (`201`)

A `201 Created` simulation requires `POST` or `PUT`:

```bash
curl -i -X POST \
  'http://localhost:8080/status/201?location=%2Fresources%2F123'
```

The response includes `Location: /resources/123` and a JSON body containing the status and location. If no `location` query parameter is supplied, the location defaults to `/resources/1`.

### Redirects

Redirects use the `location` query parameter. Its value must be a path on the Fakeway server, not an external URL:

```bash
curl -i 'http://localhost:8080/status/302?location=%2Fversion'
curl -i 'http://localhost:8080/status/307?location=%2Fhealth'
```

Without a `location` parameter, redirects point to `/health`. Fakeway rejects external URLs and unsafe location values rather than creating an external redirect.

### Partial content (`206`)

Send a byte-range request to simulate a partial response:

```bash
curl -i -H 'Range: bytes=0-5' \
  http://localhost:8080/status/206
```

The mock representation is the text `Fakeway mock response!`. For the example range, the response body is `Fakewa`; the response includes headers such as `Content-Range` and `Accept-Ranges`. `HEAD` is also accepted, and its response body is omitted.

### Not modified (`304`)

The ETag is generated from the running version. With the default version, it is:

```text
"fakeway-status-development"
```

Send that value in `If-None-Match` to simulate a matching conditional request:

```bash
curl -i -H 'If-None-Match: "fakeway-status-development"' \
  http://localhost:8080/status/304
```

If the running binary uses a different version, use the corresponding ETag value. A missing or non-matching validator produces `400 Bad Request` instead of `304`. `GET` and `HEAD` are accepted.

### Range not satisfiable (`416`)

The mock representation is 22 bytes long. To request a range beyond its end:

```bash
curl -i -H 'Range: bytes=999-' \
  http://localhost:8080/status/416
```

A valid unsatisfiable range produces `416 Range Not Satisfiable` with `Content-Range: bytes */22` and a problem-details response. Malformed ranges, missing ranges, and ranges that are actually satisfiable produce `400 Bad Request`. Only a single byte range is supported for this simulation.

## Echoing requests

Use `/echo` to inspect what Fakeway receives:

Request bodies are limited to 1 MiB. Larger request bodies receive an
HTTP `413 Payload Too Large` response.

```bash
curl 'http://localhost:8080/echo?name=fakeway&mode=test'
```

You can also send a body and custom headers:

```bash
curl -i -X POST http://localhost:8080/echo \
  -H 'Content-Type: application/json' \
  -H 'X-Test-Client: example' \
  -d '{"message":"hello"}'
```

The JSON response contains:

- `method`: the HTTP method used by the client;
- `url`: the request URL, including its query string;
- `query`: query parameters, represented as keys with arrays of values;
- `headers`: the request headers;
- `body`: the request body as a string.

This is useful for checking whether an HTTP client is sending the expected method, query parameters, headers, and payload.

## Simulating latency

Use `/delay/{duration}` to make Fakeway wait before returning a response:

```bash
curl -i http://localhost:8080/delay/500ms
curl -i http://localhost:8080/delay/2s
```

The duration must be a valid Go duration between `0s` and `30s`, inclusive. Examples include `500ms`, `2s`, and `1m` (although a duration greater than 30 seconds is rejected). A successful response looks like:

```json
{
  "duration": "2s",
  "elapsed": "2.00...s"
}
```

The `elapsed` value reports the measured time spent waiting; it may differ slightly from the requested duration due to scheduling and processing overhead. Invalid, negative, or over-30-second durations return `400 Bad Request`.

The handler also watches the request context. If the client disconnects or cancels the request while the delay is in progress, Fakeway stops waiting and returns without attempting a response.

## Configuration

Fakeway accepts these command-line flags:

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:8080` | Address on which the HTTP server listens. |
| `-read-timeout` | `10s` | Maximum duration for reading the request. |
| `-write-timeout` | `35s` | Maximum duration for writing the response. |
| `-idle-timeout` | `120s` | Maximum duration for an idle connection. |

After installing Fakeway, pass flags directly to the executable:

```bash
fakeway -addr :9090
fakeway -addr 127.0.0.1:9090 -read-timeout 5s -write-timeout 20s
```

When running from the cloned source repository, use the same flags with `go run .`:

```bash
go run . -addr :9090 -read-timeout 5s
```

The delay endpoint has its own maximum of `30s`; changing the server's HTTP timeouts does not increase that limit.

## CORS, logging, and shutdown

### CORS

Fakeway adds these CORS headers to requests:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With, If-None-Match, Range`

An `OPTIONS` request is handled as a CORS preflight and receives `204 No Content`.

### Request logging

Fakeway logs each request that reaches the logging middleware. Log entries include the HTTP method, URL path, response status code, and elapsed time.

### Graceful shutdown

Press `Ctrl+C` or send a termination signal such as `SIGTERM` to stop the server. Fakeway asks the HTTP server to finish in-flight requests, allowing up to 30 seconds for graceful shutdown before forcing the server closed.

## Error responses

Fakeway uses the `application/problem+json` media type for simulated HTTP errors and for invalid simulation requests. A typical problem response has this shape:

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "The requested resource was not found.",
  "instance": "/status/404"
}
```

The fields are:

| Field | Meaning |
|---|---|
| `type` | Problem type URI; Fakeway currently uses `about:blank`. |
| `title` | Short status description, or `Unknown Status` for codes Go does not recognize. |
| `status` | HTTP status code returned in the response. |
| `detail` | Explanation of the simulated error or invalid request. |
| `instance` | Request path associated with the response. |

For example, `/status/invalid`, `/status/100`, `/status/600`, `/status/305`, and `/status/306` return `400 Bad Request` because the requested simulation is not supported. A status simulation may also return `400` when its required method or headers are missing or invalid.

## License

Fakeway is distributed under the MIT License. See [`LICENSE`](LICENSE).
