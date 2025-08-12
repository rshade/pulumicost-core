# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PulumiCost Core is a CLI tool and plugin host system for calculating cloud infrastructure costs from Pulumi infrastructure definitions. It provides both projected cost estimates and actual historical cost analysis through a plugin-based architecture.

## Build Commands

- `make build` - Build the pulumicost binary to bin/pulumicost
- `make test` - Run all tests
- `make lint` - Run golangci-lint (requires installation)
- `make run` - Build and run with --help
- `make dev` - Build and run without arguments
- `make clean` - Remove build artifacts

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`) - Cobra-based command interface with subcommands:
   - `cost projected` - Calculate projected costs from Pulumi preview JSON
   - `cost actual` - Fetch actual historical costs with time ranges
   - `plugin list` - List installed plugins
   - `plugin validate` - Validate plugin installations

2. **Engine** (`internal/engine/`) - Core cost calculation logic:
   - Orchestrates between plugins and local pricing specs
   - Handles resource mapping and cost aggregation  
   - Supports multiple output formats (table, JSON, NDJSON)

3. **Plugin Host System** (`internal/pluginhost/`) - gRPC plugin management:
   - `Client` - Wraps plugin gRPC connections
   - `ProcessLauncher` - Launches plugins as TCP processes
   - `StdioLauncher` - Alternative stdio-based plugin communication

4. **Registry** (`internal/registry/`) - Plugin discovery and lifecycle:
   - Scans `~/.pulumicost/plugins/<name>/<version>/` for binaries
   - Manages plugin manifests and metadata

5. **Ingestion** (`internal/ingest/`) - Pulumi plan parsing:
   - Converts `pulumi preview --json` output to resource descriptors
   - Extracts provider and resource type information

6. **Spec System** (`internal/spec/`) - Local pricing specification:
   - YAML-based pricing specs in `~/.pulumicost/specs/`
   - Fallback when plugins don't provide pricing

### Plugin Protocol

Plugins communicate via gRPC using protocol buffers defined in the `pulumicost-spec` repository. Current implementation uses mock protobuf definitions (`internal/proto/mock.go`) until the spec repository is fully implemented.

Key plugin methods:
- `Name()` - Plugin identification
- `GetProjectedCost()` - Calculate estimated costs for resources
- `GetActualCost()` - Retrieve historical costs from cloud APIs

## Development Workflow

1. **Resource Flow**: Pulumi JSON → Resource Descriptors → Plugin Queries → Cost Results → Output Rendering

2. **Plugin Discovery**: Registry scans plugin directories → Launches processes → Establishes gRPC connections → Makes API calls

3. **Cost Calculation**: Try plugins first → Fallback to local specs → Aggregate results → Render output

## Key Files

- `cmd/pulumicost/main.go` - CLI entry point
- `internal/engine/engine.go` - Core orchestration logic
- `internal/pluginhost/host.go` - Plugin client management  
- `internal/ingest/pulumi_plan.go` - Pulumi plan parsing
- `examples/plans/aws-simple-plan.json` - Sample Pulumi plan for testing
- `examples/specs/aws-ec2-t3-micro.yaml` - Sample pricing specification

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/grpc` - Plugin communication
- `gopkg.in/yaml.v3` - YAML spec parsing
- `github.com/rshade/pulumicost-spec` - Protocol definitions (via replace directive to ../pulumicost-spec)

## Testing

Use the provided example files:
```bash
./bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json
./bin/pulumicost plugin list
./bin/pulumicost plugin validate
```