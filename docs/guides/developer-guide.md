---
layout: default
title: Developer Guide
description: Complete guide for engineers - extend PulumiCost and build plugins
---

# PulumiCost Developer Guide

This guide is for **engineers and developers** who want to extend PulumiCost by building plugins or contributing to the core project.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Architecture Overview](#architecture-overview)
3. [Development Setup](#development-setup)
4. [Building Plugins](#building-plugins)
5. [Contributing to Core](#contributing-to-core)
6. [Testing](#testing)
7. [Deployment](#deployment)

---

## Getting Started

### Prerequisites

- Go 1.24+ (for core development)
- Git
- Make
- Node.js 18+ (for documentation tools)
- Docker (optional, for containerized testing)

### Quick Setup

```bash
# Clone repository
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core

# Build
make build

# Test
make test

# Run
./bin/pulumicost --help
```

---

## Architecture Overview

### Core Components

```
┌─────────────────┐
│   Pulumi JSON   │  User's infrastructure definition
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Ingestion     │  Parse resources from Pulumi output
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Engine       │  Orchestrate cost calculation
├─────────────────┤
│  • Resource     │
│    Mapping      │
│  • Cost         │
│    Calculation  │
│  • Aggregation  │
└────────┬────────┘
         │
    ┌────┴─────┐
    ▼          ▼
┌────────┐  ┌──────┐
│Plugins │  │Specs │  Cost sources
└────────┘  └──────┘
    │          │
    └────┬─────┘
         │
         ▼
┌─────────────────┐
│     Output      │  Table, JSON, NDJSON
└─────────────────┘
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `internal/cli` | Command-line interface (Cobra) |
| `internal/engine` | Cost calculation logic |
| `internal/ingest` | Pulumi plan parsing |
| `internal/pluginhost` | Plugin gRPC communication |
| `internal/registry` | Plugin discovery |
| `internal/spec` | Local pricing specifications |
| `pkg/pluginsdk` | Plugin SDK for developers |

---

## Development Setup

### Local Development

```bash
# Install dependencies
go mod download

# Build binary
make build

# Run with example plan
./bin/pulumicost cost projected \
  --pulumi-json examples/plans/aws-simple-plan.json

# Run tests
make test

# Run linters
make lint
```

### Documentation Development

```bash
# Install Ruby dependencies
cd docs
bundle install
cd ..

# Serve docs locally
make docs-serve
# Visit http://localhost:4000/pulumicost-core/

# Lint docs
make docs-lint
```

### IDE Setup

**VS Code:**
```json
{
  "go.lintOnSave": "package",
  "go.useLanguageServer": true,
  "[go]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "golang.go"
  }
}
```

---

## Building Plugins

### Quick Plugin Template

Create a new plugin project:

```bash
cd ../
mkdir pulumicost-plugin-myservice
cd pulumicost-plugin-myservice
go mod init github.com/yourname/pulumicost-plugin-myservice
```

### Minimal Plugin

```go
package main

import (
    "context"
    "log"

    pb "github.com/rshade/pulumicost-spec/sdk/go/proto/pulumicost/v1"
)

type MyPlugin struct{}

func (p *MyPlugin) GetProjectedCost(ctx context.Context,
    req *pb.GetProjectedCostRequest) (*pb.GetProjectedCostResponse, error) {

    // Fetch cost from your API
    costs := make([]*pb.Cost, 0)

    for _, resource := range req.Resources {
        cost := &pb.Cost{
            ResourceId: resource.Id,
            TotalCost:  calculateCost(resource),
            Currency:   "USD",
        }
        costs = append(costs, cost)
    }

    return &pb.GetProjectedCostResponse{
        Costs: costs,
    }, nil
}

func calculateCost(resource *pb.Resource) float64 {
    // Your cost calculation logic
    return 0.0
}
```

### Full Plugin Development

See [Plugin Development Guide](../plugins/plugin-development.md) for:
- Complete implementation walkthrough
- gRPC service setup
- Error handling patterns
- Testing strategies
- Deployment instructions

### Example: Vantage Plugin

The Vantage plugin is a complete reference implementation:

```bash
# See implementation at
cat ../pulumicost-plugin-vantage/main.go
```

---

## Contributing to Core

### Setting Up Dev Branch

```bash
# Fetch latest
git fetch upstream

# Create feature branch
git checkout -b feature/my-feature upstream/main

# Make changes
# ... edit files ...

# Test changes
make test
make lint
make docs-validate
```

### Code Style

**Go:**
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Run `gofmt` on your code
- Use `golangci-lint` for linting
- Write clear variable names
- Add godoc comments for exported functions

**Example:**
```go
// GetActualCost retrieves actual historical costs for resources.
// It supports filtering by tags and grouping by dimension.
func (e *Engine) GetActualCost(ctx context.Context,
    req *ActualCostRequest) (*ActualCostResponse, error) {
    // Implementation
}
```

**Markdown:**
- Follow [Google style guide](https://developers.google.com/style)
- Use clear headings
- Provide code examples
- Run `make docs-lint` before committing

### Commit Messages

```
type: Brief description

More detailed explanation of changes.
- What changed
- Why it changed
- Any implementation notes

Closes #123
```

**Types:**
- `feature` - New functionality
- `fix` - Bug fixes
- `docs` - Documentation
- `test` - Tests
- `refactor` - Code restructuring
- `perf` - Performance improvements

### Testing Your Changes

```bash
# Run all tests
make test

# Run specific package
go test -v ./internal/engine/...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestActualCost ./internal/engine/...
```

### Pull Request Process

1. **Update from main:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks:**
   ```bash
   make test
   make lint
   make validate
   make docs-validate
   ```

3. **Push and create PR:**
   ```bash
   git push origin feature/my-feature
   # Create PR on GitHub
   ```

4. **Address feedback:**
   - Respond to comments
   - Push additional commits
   - Rebase if requested

---

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# View coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

```go
func TestGetActualCost(t *testing.T) {
    // Arrange - Set up test data
    request := &ActualCostRequest{
        StartDate: "2024-01-01",
        EndDate:   "2024-01-31",
    }

    // Act - Execute function
    response, err := engine.GetActualCost(context.Background(), request)

    // Assert - Verify results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if response == nil {
        t.Fatal("expected response, got nil")
    }
}
```

### Integration Testing

For testing with real plugins:

```bash
# Ensure plugins are installed
./bin/pulumicost plugin list

# Test with example plan
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json
```

---

## Deployment

### Building Releases

```bash
# Create version tag
git tag v0.1.0
git push origin v0.1.0

# GitHub Actions automatically:
# 1. Builds cross-platform binaries
# 2. Creates release
# 3. Uploads checksums
```

### Plugin Installation

Users install plugins to: `~/.pulumicost/plugins/<name>/<version>/`

**Structure:**
```
~/.pulumicost/plugins/
├── myplugin/
│   └── 0.1.0/
│       ├── pulumicost-myplugin    # Plugin binary
│       └── plugin.manifest.json    # Metadata
```

### Docker Deployment

```dockerfile
FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /app/bin/pulumicost /usr/local/bin/
ENTRYPOINT ["pulumicost"]
```

---

## Useful Commands

```bash
# Development
make build              # Build binary
make test               # Run tests
make lint               # Code linting
make validate           # Validation
make clean              # Clean artifacts

# Documentation
make docs-lint          # Lint docs
make docs-serve         # Serve locally
make docs-build         # Build site
make docs-validate      # Validate structure

# Git
git fetch upstream      # Get latest changes
git rebase upstream/main # Rebase on main
git push origin branch  # Push changes
```

---

## Resources

- **Plugin Development:** [Plugin Development Guide](../plugins/plugin-development.md)
- **Plugin SDK:** [Plugin SDK Reference](../plugins/plugin-sdk.md)
- **Examples:** [Code Examples](../plugins/plugin-examples.md)
- **Architecture:** [System Architecture](../architecture/system-overview.md)
- **Contributing:** [Contributing Guide](../support/contributing.md)
- **Vantage Plugin:** [Vantage Implementation Example](../plugins/vantage/README.md)

---

**Last Updated:** 2025-10-29
