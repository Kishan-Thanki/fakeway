# syntax=docker/dockerfile:1
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies (if any in future)
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build argument for version injection
ARG VERSION=dev

# Build fully static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o fakeway .

# Final minimal stage
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates and curl for health check
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy binary from builder
COPY --from=builder /app/fakeway /app/fakeway

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/fakeway"]
