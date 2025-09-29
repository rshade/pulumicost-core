# Repository Guidelines

## Project Structure & Module Organization

- Core CLI lives in `cmd/pulumicost`; shared libraries stay under `pkg/`.
- Service modules belong in `internal/`; docs live in `docs/`;
  examples in `examples/`.
- Test fixtures reside in `testdata/`; compiled binaries land in `bin/`.
- Keep scratch output out of source folders; use `bin/` or ignored workspace
  paths.

## Build, Test, and Development Commands

- Use Go 1.24+ for local work.
- `make build` compiles `bin/pulumicost` with version metadata embedded.
- `make test` runs `go test -v ./...`; add `GOFLAGS=-count=1` to avoid cache
  reuse.
- `make lint` wraps `golangci-lint` (goimports + golines) for style enforcement.
- `make validate` runs `go mod tidy -diff` and `go vet` for dependency checks.
- After building, run the sanity check:

  ```bash
  bin/pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json
  ```

## Coding Style & Naming Conventions

- Format Go sources with `gofmt` or `goimports`; golines keeps lines near 120
  chars.
- Prefer descriptive package names like `internal/costcalc` and `pkg/version`.
- Keep functions concise and scoped; document exported identifiers with clear
  comments.
- Wrap errors using `fmt.Errorf("context: %w", err)` to preserve stack context.
- Import new packages under `github.com/rshade/pulumicost-core/...`.

## Testing Guidelines

- Tests rely on the Go standard library and `testify` assertions.
- Use table driven tests and mirror target packages with `*_test.go` files.
- Store golden data and fixtures in `testdata/` directories.
- Run `go test ./internal/...` before pushing. Generate coverage with
  `go test -cover ./... > coverage.out`.
- Capture representative CLI output when modifying analyzers or renderers.

## Commit & Pull Request Guidelines

- Follow Conventional Commits (`feat:`, `fix(deps):`, `📝 docs:`) and add scopes
  when useful.
- Squash WIP commits before sharing branches to keep history clean.
- PRs should explain the change, link issues, and note `make lint` plus
  `make test` results.
- Include CLI screenshots or snippets when altering UX or output formatting.
- Convert drafts only after checks pass and feedback is addressed.

## Security & Configuration Tips

- Never commit cloud credentials or plugin binaries; keep custom adapters under
  `~/.pulumicost/plugins`.
- Validate configuration updates against `docs/plugin-system.md` for sandbox
  safety.
- Store secrets in local tooling (e.g., `.envrc`) rather than tracked
  configuration files.
