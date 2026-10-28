.DEFAULT_GOAL := help

GO ?= go
DOCKER ?= docker
CURL ?= curl
BINARY ?= fakeway
VERSION ?= v1.2.0
IMAGE ?= fakeway
PORT ?= 8080
DIST_DIR ?= dist
COVERAGE_FILE ?= coverage.out
GOVULNCHECK_VERSION ?= v1.9.0
DOCKER_SMOKE_PORT ?= 18081

GO_FILES := $(wildcard *.go)
BUILD_FLAGS := -trimpath -buildvcs=false
LDFLAGS := -s -w -X github.com/kishan-thanki/fakeway/internal/httpapi.Version=$(VERSION)

.PHONY: help fmt fmt-check mod-check vet test vulncheck check full build run install \
        release-binaries verify-checksums docker-build docker-run docker-smoke clean

help: 
	@grep -E '^[a-zA-Z0-9_-]+:.*##' "$(firstword $(MAKEFILE_LIST))" | awk 'BEGIN {FS = ":.*## "} {printf "%-20s %s\n", $$1, $$2}'

fmt: 
	@test -n "$(GO_FILES)" || { echo "No Go source files found."; exit 1; }
	gofmt -w $(GO_FILES)

fmt-check: 
	@test -n "$(GO_FILES)" || { echo "No Go source files found."; exit 1; }
	@unformatted="$$(gofmt -l $(GO_FILES))"; \
	if [ -n "$$unformatted" ]; then \
		echo "Run 'make fmt' to format these files:"; \
		printf '%s\n' "$$unformatted"; \
		exit 1; \
	fi

mod-check:
	$(GO) list -mod=readonly -deps -test ./... >/dev/null
	$(GO) mod verify

vet: 
	$(GO) vet ./...

test: 
	$(GO) test -race -coverprofile="$(COVERAGE_FILE)" -covermode=atomic ./...
	@$(GO) tool cover -func="$(COVERAGE_FILE)" | tail -n 1

vulncheck: 
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

check: fmt-check mod-check vet test vulncheck 

full: clean 
	$(MAKE) check
	$(MAKE) build

build: 
	@mkdir -p "$(dir $(BINARY))"
	CGO_ENABLED=0 $(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o "$(BINARY)" .

run: 
	$(GO) run -ldflags="$(LDFLAGS)" . $(ARGS)

install: 
	CGO_ENABLED=0 $(GO) install $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" .

release-binaries: 
	@test -n "$(VERSION)" && [ "$(VERSION)" != "dev" ] || { echo "Set VERSION, e.g. make release-binaries VERSION=v1.2.3"; exit 1; }
	@mkdir -p "$(DIST_DIR)"
	@rm -f "$(DIST_DIR)"/fakeway-* "$(DIST_DIR)"/SHA256SUMS
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/fakeway-linux-amd64" .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/fakeway-linux-arm64" .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/fakeway-darwin-amd64" .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/fakeway-darwin-arm64" .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o "$(DIST_DIR)/fakeway-windows-amd64.exe" .
	@set -eu; cd "$(DIST_DIR)"; \
	if command -v sha256sum >/dev/null 2>&1; then sha256sum fakeway-* > SHA256SUMS; \
	elif command -v shasum >/dev/null 2>&1; then shasum -a 256 fakeway-* > SHA256SUMS; \
	else echo "Install sha256sum or shasum to create release checksums."; exit 1; fi
	@$(MAKE) --no-print-directory verify-checksums DIST_DIR="$(DIST_DIR)"

verify-checksums: 
	@test -f "$(DIST_DIR)/SHA256SUMS" || { echo "$(DIST_DIR)/SHA256SUMS not found."; exit 1; }
	@set -eu; cd "$(DIST_DIR)"; \
	if command -v sha256sum >/dev/null 2>&1; then sha256sum --check SHA256SUMS; \
	elif command -v shasum >/dev/null 2>&1; then shasum -a 256 --check SHA256SUMS; \
	else echo "Install sha256sum or shasum to verify release checksums."; exit 1; fi

docker-build: 
	$(DOCKER) build --build-arg VERSION="$(VERSION)" -t "$(IMAGE):$(VERSION)" .

docker-run: 
	$(DOCKER) run --rm -p "$(PORT):8080" "$(IMAGE):$(VERSION)"

docker-smoke: docker-build 
	@set -eu; \
	command -v "$(CURL)" >/dev/null 2>&1 || { echo "curl is required for Docker smoke tests."; exit 1; }; \
	container_name="fakeway-smoke-$$$$"; \
	"$(DOCKER)" run --detach --name "$$container_name" -p "127.0.0.1:$(DOCKER_SMOKE_PORT):8080" "$(IMAGE):$(VERSION)" >/dev/null; \
	trap '"$(DOCKER)" rm --force "$$container_name" >/dev/null 2>&1 || true' EXIT; \
	attempt=0; \
	while [ $$attempt -lt 50 ]; do \
		if "$(CURL)" --fail --silent "http://127.0.0.1:$(DOCKER_SMOKE_PORT)/health" >/dev/null 2>&1; then echo "Docker smoke test passed."; exit 0; fi; \
		attempt=$$((attempt + 1)); sleep 0.1; \
	done; \
	"$(DOCKER)" logs "$$container_name" || true; echo "Docker smoke test timed out."; exit 1

clean: 
	rm -f "$(BINARY)" "$(COVERAGE_FILE)"
	rm -rf "$(DIST_DIR)"
