# pulumicost-core Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-22

## Active Technologies
- Local Pulumi state (ephemeral), no persistent DB. (010-e2e-cost-testing)
- N/A (Stateless operation) (008-analyzer-plugin)

- Go 1.25.5
- `github.com/rshade/pulumicost-spec`
- `google.golang.org/grpc` (002-implement-supports-handler)

## Project Structure

```text
cmd/
internal/
pkg/
test/
testdata/
```

## Commands

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Run
make run
```

## Code Style

Go 1.24.10: Follow standard conventions

<!-- MANUAL ADDITIONS START -->
## Workflow Restrictions

- **NEVER COMMIT**: Do not execute `git commit`. Always stop after `git add` and ask the user to review/commit.
<!-- MANUAL ADDITIONS END -->

## Integration & Testing

- **E2E Testing**: Uses `pulumi preview --json` against real AWS infrastructure (ephemeral stacks).
- **Pulumi Plan JSON**: The structure of `pulumi preview --json` output nests resource state under `newState` (for creates/updates). Ingest logic MUST check `newState` to correctly extract `Inputs` and `Type`.
- **Resource Type Compatibility**: Plugins must handle Pulumi-style resource types (e.g., `aws:ec2/instance:Instance`) or Core must normalize them. Currently, plugins are expected to handle the mapping.
- **Property Extraction**: Core (`adapter.go`) relies on populated `Inputs` to extract SKU and Region. If `Inputs` are empty (due to ingest issues), pricing lookup fails.

## Recent Changes
- 010-e2e-cost-testing: Fixed E2E test failure by parsing `newState` in Pulumi plan JSON.
- 010-e2e-cost-testing: Patched `aws-public` plugin to support `aws:ec2/instance:Instance` resource type.
- 010-e2e-cost-testing: Verified cost calculation accuracy ($7.59/month for t3.micro).
- 010-e2e-cost-testing: Added Go 1.25
- 010-e2e-cost-testing: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
- 010-e2e-cost-testing: Added Go 1.25
- 008-analyzer-plugin: Added Go 1.25 + `github.com/pulumi/pulumi/sdk/v3` (for `pulumirpc`), `google.golang.org/grpc`, `github.com/spf13/cobra`
- 008-analyzer-plugin: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
