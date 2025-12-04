# pulumicost-core Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-22

## Active Technologies
- Local Pulumi state (ephemeral), no persistent DB. (010-e2e-cost-testing)

- Go 1.25
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

## Recent Changes
- 010-e2e-cost-testing: Added Go 1.25
- 010-e2e-cost-testing: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
- 010-e2e-cost-testing: Added Go 1.25
