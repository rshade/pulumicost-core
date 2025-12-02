# Repository Guidelines

## Project Structure & Module Organization

- `cmd/pulumicost`: CLI entrypoint and flag wiring.
- `internal/` packages: core logic (engine, ingest, registry, pluginhost, config,
  specvalidate, logging) kept unexported to guard APIs.
- `pkg/version`: shared version/build metadata used by the CLI.
- `examples/` and `testdata/`: Pulumi plan fixtures and sample specs; prefer
  extending these for reproducible tests.
- `docs/`: Jekyll site and contributor docs; `scripts/` contains helper tooling;
  `bin/` is populated by builds.

## Build, Test, and Development Commands

- `make build`: Compile the `pulumicost` binary to `bin/` with version metadata.
- `make test` | `make test-race`: Run Go tests (optionally with the race
  detector) across all packages.
- `make lint`: Run `golangci-lint` v2.6.2 plus `markdownlint` (expects
  `AGENTS.md` to pass).
- `make validate`: Run `go mod tidy -diff` and `go vet` to verify module state
  and static checks.
- `make run` / `make dev`: Build then run the CLI; use `make inspect` to launch
  the MCP inspector after a build.
- Docs: `make docs-lint`, `make docs-build`, `make docs-serve` for the Jekyll
  site.

## Coding Style & Naming Conventions

- Go 1.25+, tabs for indentation. Format with `gofmt`; imports and line length
  enforced by `goimports`/`golines` via `golangci-lint` (`.golangci.yml`).
- Package names are lowercase and succinct; exported identifiers need Go doc
  comments when part of the CLI surface.
- CLI flags use kebab-case (`--pulumi-json`, `--group-by`); config/env keys use
  uppercase snake (`PULUMICOST_PLUGIN_*`).
- Prefer structured logging through `internal/logging`; keep public structs
  JSON-tagged for CLI/JSON outputs.

## Testing Guidelines

- Author `_test.go` files with `TestXxx`/`BenchmarkXxx`; keep table-driven tests
  near the code they cover.
- Use `go test ./...` (optionally `-cover` or `-race`) before submitting; add
  fixtures to `examples/` or `testdata/` instead of embedding large literals.
- When adding plugins or adapters, include conformance coverage in
  `internal/conformance` and targeted cases in `internal/engine` or
  `internal/registry`.

## Commit & Pull Request Guidelines

- Commit messages follow Conventional Commits (`feat:`, `fix:`, `chore:`…) as
  enforced by `commitlint.config.js`; keep them scoped and imperative.
- PRs should include: a concise summary, linked issues, a short test plan (e.g.,
  `make test`, `make lint`, `make validate`), and CLI output or screenshots when
  user-facing behavior changes.
- Avoid bundling unrelated changes; keep docs and code changes cohesive. Flag
  breaking changes explicitly in the PR description.

## Security & Configuration Tips

- Do not commit secrets; prefer env vars (`PULUMICOST_PLUGIN_AWS_*`,
  `PULUMICOST_PLUGIN_VANTAGE_*`, etc.) and `~/.pulumicost/config.yaml` for local
  config.
- Plugins live under `~/.pulumicost/plugins/`; validate with
  `pulumicost plugin validate` before shipping.
- Treat Pulumi plan JSON files as sensitive if they contain identifiers; scrub
  or use redacted fixtures in examples.
