BINARY=pulumicost
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GOLANGCI_LINT?=$(HOME)/go/bin/golangci-lint
GOLANGCI_LINT_VERSION?=2.5.0
MARKDOWNLINT?=markdownlint
MARKDOWNLINT_FILES?=AGENTS.md

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
	@echo "Running golangci-lint (expected version $(GOLANGCI_LINT_VERSION))..."
	@$(GOLANGCI_LINT) --version | grep -q "$(GOLANGCI_LINT_VERSION)" || \
		(echo "golangci-lint $(GOLANGCI_LINT_VERSION) required. Install with"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$HOME/go/bin v$(GOLANGCI_LINT_VERSION)"; exit 1)
	$(GOLANGCI_LINT) run --allow-parallel-runners
	@echo "Running markdownlint..."
	@command -v $(MARKDOWNLINT) >/dev/null 2>&1 || \
		(echo "markdownlint CLI not found. Install with"; \
		echo "  npm install -g markdownlint-cli@0.45.0"; exit 1)
	$(MARKDOWNLINT) $(MARKDOWNLINT_FILES)

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
	@echo "  lint     - Run Go + Markdown linters"
	@echo "  validate - Run validation (go mod, vet, format)"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Build and run with --help"
	@echo "  dev      - Build and run"
	@echo "  help     - Show this help message"
