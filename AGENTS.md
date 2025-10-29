# Agent Guidelines for pulumicost-core

## Build/Test Commands

- `make build` - Compile binary with version metadata
- `make test` - Run all tests (`go test -v ./...`)
- `make lint` - Run golangci-lint + markdownlint (120+ linters enabled)
- `make validate` - Run `go mod tidy` and `go vet`
- Single test: `go test -v ./path/to/package -run TestName`

## Code Style (Go 1.24+)

- Line length: 120 chars (golines), imports: goimports
- Error handling: `fmt.Errorf("context: %w", err)` for wrapping
- No globals/init functions, document exported identifiers
- Tests: testify/assert+require, table-driven, separate _test packages
- Naming: descriptive packages (e.g., `internal/costcalc`, `pkg/version`)
- Functions: concise and scoped, clear comments

## Project Structure

- CLI: `cmd/pulumicost`, shared libs: `pkg/`, services: `internal/`
- Tests: `*_test.go` files, fixtures in `testdata/`, binaries in `bin/`
- Examples: `examples/`, docs: `docs/`

## Development Workflow

- Always run `make lint` and `make test` before commits
- Follow Conventional Commits with scopes
- PRs: explain changes, link issues, include test results
- No cloud credentials/plugin binaries in repo
