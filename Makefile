BINARY=pulumicost
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GOLANGCI_LINT?=$(HOME)/go/bin/golangci-lint
GOLANGCI_LINT_VERSION?=2.6.2
MARKDOWNLINT?=markdownlint
MARKDOWNLINT_FILES?=AGENTS.md

LDFLAGS=-ldflags "-X 'github.com/rshade/pulumicost-core/pkg/version.version=$(VERSION)' \
                  -X 'github.com/rshade/pulumicost-core/pkg/version.gitCommit=$(COMMIT)' \
                  -X 'github.com/rshade/pulumicost-core/pkg/version.buildDate=$(BUILD_DATE)'"

.PHONY: all build test lint validate clean run dev inspect help docs-lint docs-serve docs-build docs-validate

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

inspect: build ## Launch the MCP Inspector for interactive testing
	@echo "Starting MCP Inspector for $(BINARY)..."
	@echo "Open the URL shown below in your browser to interact with the MCP server"
	npx @modelcontextprotocol/inspector $$(realpath bin/$(BINARY))

docs-lint:
	@echo "Linting documentation..."
	@command -v markdownlint-cli2 >/dev/null 2>&1 || \
		(echo "markdownlint-cli2 not found. Install with:"; \
		echo "  npm install -g markdownlint-cli2"; exit 1)
	markdownlint-cli2 --config docs/.markdownlint-cli2.jsonc 'docs/**/*.md' --ignore 'docs/_site/**' || true
	@echo "Documentation linting complete."

docs-serve:
	@echo "Serving documentation locally at http://localhost:4000/pulumicost-core"
	@cd docs && bundle install > /dev/null 2>&1 || true
	@cd docs && bundle exec jekyll serve --host 0.0.0.0

docs-build:
	@echo "Building documentation site..."
	@cd docs && bundle install > /dev/null 2>&1 || true
	@cd docs && bundle exec jekyll build
	@echo "Documentation built to docs/_site/"

docs-validate: docs-lint
	@echo "Validating documentation structure..."
	@test -f docs/README.md || (echo "Missing: docs/README.md"; exit 1)
	@test -f docs/plan.md || (echo "Missing: docs/plan.md"; exit 1)
	@test -f docs/llms.txt || (echo "Missing: docs/llms.txt"; exit 1)
	@test -f docs/_config.yml || (echo "Missing: docs/_config.yml"; exit 1)
	@test -f docs/.markdownlint-cli2.jsonc || (echo "Missing: docs/.markdownlint-cli2.jsonc"; exit 1)
	@echo "✓ All required documentation files present"
	@echo "✓ Documentation validation passed"

help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  lint         - Run Go + Markdown linters"
	@echo "  validate     - Run validation (go mod, vet, format)"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run with --help"
	@echo "  dev          - Build and run"
	@echo "  inspect      - Launch MCP Inspector for interactive testing"
	@echo ""
	@echo "Documentation targets:"
	@echo "  docs-lint    - Lint documentation markdown"
	@echo "  docs-build   - Build documentation site"
	@echo "  docs-serve   - Serve documentation locally (http://localhost:4000)"
	@echo "  docs-validate- Validate documentation structure"
	@echo ""
	@echo "  help         - Show this help message"
