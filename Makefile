BINARY ?= fakeway
VERSION ?= v1.2.0

.PHONY: build run test docker-build docker-run clean

build: ## Build the Fakeway binary locally
	CGO_ENABLED=0 go build -trimpath -buildvcs=false \
		-ldflags="-s -w -X github.com/kishan-thanki/fakeway/internal/httpapi.Version=$(VERSION)" \
		-o $(BINARY) .

run: ## Run Fakeway locally
	go run -ldflags="-s -w -X github.com/kishan-thanki/fakeway/internal/httpapi.Version=$(VERSION)" .

test: ## Run unit tests with race detection
	go test -v -race ./...

docker-build: ## Build Docker image locally
	docker build --build-arg VERSION=$(VERSION) -t fakeway:$(VERSION) .

docker-run: ## Run Docker container locally on port 8080
	docker run --rm -p 8080:8080 fakeway:$(VERSION)

clean: ## Remove built binaries and artifacts
	rm -f $(BINARY)
	rm -rf dist/
