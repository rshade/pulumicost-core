BINARY=pulumicost
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X 'github.com/rshade/pulumicost-core/pkg/version.Version=$(VERSION)' \
                  -X 'github.com/rshade/pulumicost-core/pkg/version.GitCommit=$(COMMIT)' \
                  -X 'github.com/rshade/pulumicost-core/pkg/version.BuildDate=$(BUILD_DATE)'"

.PHONY: all build test lint clean run dev help

all: build

build:
	@echo "Building $(BINARY)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/pulumicost

test:
	@echo "Running tests..."
	go test -v ./...

lint:
	@echo "Running linter..."
	golangci-lint run

clean:
	@echo "Cleaning..."
	rm -rf bin/

run: build
	@echo "Running $(BINARY)..."
	bin/$(BINARY) --help

dev: build
	@echo "Running development build..."
	bin/$(BINARY)

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  test     - Run tests"
	@echo "  lint     - Run linter"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Build and run with --help"
	@echo "  dev      - Build and run"
	@echo "  help     - Show this help message"