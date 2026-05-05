VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/marstid/nuc/pkg/version.Version=$(VERSION) \
                      -X github.com/marstid/nuc/pkg/version.Commit=$(COMMIT) \
                      -X github.com/marstid/nuc/pkg/version.Date=$(DATE)"
BINARY     := nuc
MCP_BINARY := nuc-mcp

.PHONY: build build-nuc build-mcp test lint install install-nuc install-mcp clean fmt vet run-mcp run-mcp-http help

## build: Build both nuc and nuc-mcp binaries
build: build-nuc build-mcp

## build-nuc: Build the nuc CLI binary
build-nuc:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/nuc

## build-mcp: Build the nuc-mcp server binary
build-mcp:
	go build $(LDFLAGS) -o bin/$(MCP_BINARY) ./cmd/nuc-mcp

## test: Run all tests with race detection
test:
	go test -race -cover ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## install: Install both binaries to $GOPATH/bin
install: install-nuc install-mcp

## install-nuc: Install nuc to $GOPATH/bin
install-nuc:
	go install $(LDFLAGS) ./cmd/nuc

## install-mcp: Install nuc-mcp to $GOPATH/bin
install-mcp:
	go install $(LDFLAGS) ./cmd/nuc-mcp

## clean: Remove build artifacts
clean:
	rm -rf bin/

## fmt: Format code with gofumpt
fmt:
	gofumpt -w .

## vet: Run go vet
vet:
	go vet ./...

## run-mcp: Run the MCP server over stdio
run-mcp: build-mcp
	./bin/$(MCP_BINARY)

## run-mcp-http: Run the MCP server over HTTP on localhost:8080
run-mcp-http: build-mcp
	./bin/$(MCP_BINARY) --transport=http --addr=localhost:8080

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
