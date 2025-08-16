BINARY=pulumicost
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X 'github.com/rshade/pulumicost-core/pkg/version.version=$(VERSION)' \
                  -X 'github.com/rshade/pulumicost-core/pkg/version.gitCommit=$(COMMIT)' \
                  -X 'github.com/rshade/pulumicost-core/pkg/version.buildDate=$(BUILD_DATE)'"

.PHONY: all build test lint validate clean run dev help

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
	~/go/bin/golangci-lint run --allow-parallel-runners

validate:
	@echo "Running validation..."
	@echo "Checking go modules..."
	go mod tidy -diff
	@echo "Running go vet..."
	go vet ./...
	@echo "Validation complete."

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
	@echo "  validate - Run validation (go mod, vet, format)"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Build and run with --help"
	@echo "  dev      - Build and run"
	@echo "  help     - Show this help message"