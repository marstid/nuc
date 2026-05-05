VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/marstid/nuc/pkg/version.Version=$(VERSION) \
                      -X github.com/marstid/nuc/pkg/version.Commit=$(COMMIT) \
                      -X github.com/marstid/nuc/pkg/version.Date=$(DATE)"
BINARY  := nuc

.PHONY: build test lint install clean fmt vet

## build: Build the nuc binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/nuc

## test: Run all tests with race detection
test:
	go test -race -cover ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## install: Install nuc to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/nuc

## clean: Remove build artifacts
clean:
	rm -rf bin/

## fmt: Format code with gofumpt
fmt:
	gofumpt -w .

## vet: Run go vet
vet:
	go vet ./...

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
