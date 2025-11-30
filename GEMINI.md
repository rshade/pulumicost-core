# pulumicost-core Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-22

## Active Technologies

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
