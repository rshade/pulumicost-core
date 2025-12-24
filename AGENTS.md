# Repository Guidelines

## Project Structure & Module Organization

- `cmd/pulumicost`: CLI entrypoint and flag wiring.
- `internal/` packages: core logic (engine, ingest, registry, pluginhost, config,
  specvalidate, logging, **analyzer**) kept unexported to guard APIs.
- `pkg/version`: shared version/build metadata used by the CLI.
- `examples/` and `testdata/`: Pulumi plan fixtures and sample specs; prefer
  extending these for reproducible tests.
- `test/e2e/fixtures/`: Real Pulumi project fixtures for E2E tests.
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
- `pulumicost plugin init`: Initialize a new plugin project.
- `pulumicost plugin install`: Install a plugin from a registry or URL.
- `pulumicost plugin update`: Update an installed plugin.
- `pulumicost plugin remove`: Remove an installed plugin.
- `pulumicost plugin certify`: Run certification tests for a plugin.
- `pulumicost analyzer serve`: Start the Pulumi Analyzer gRPC server.
- Docs: `make docs-lint`, `make docs-build`, `make docs-serve` for the Jekyll
  site.

## Coding Style & Naming Conventions

- Go 1.25.5+, tabs for indentation. Format with `gofmt`; imports and line length
  enforced by `goimports`/`golines` via `golangci-lint` (`.golangci.yml`).
- Package names are lowercase and succinct; exported identifiers need Go doc
  comments when part of the CLI surface.
- CLI flags use kebab-case (`--pulumi-json`, `--group-by`); config/env keys use
  uppercase snake (`PULUMICOST_PLUGIN_*`, `PULUMICOST_CONFIG_STRICT`).
- Prefer structured logging through `internal/logging`; keep public structs
  JSON-tagged for CLI/JSON outputs.

## Documentation Standards

- Run `make docs-lint` before committing documentation changes
- Use frontmatter YAML with `title`, `description`, and `layout` fields
- **CRITICAL**: Files with frontmatter must NOT have duplicate H1 - the frontmatter
  `title` serves as the page H1, content should start with H2 or text

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

## Session Analysis - Recommended Updates

Based on recent development sessions, consider adding:

### Go Version Management

- **Version Consistency**: When updating Go versions, update both `go.mod` and ALL
  markdown files simultaneously
- **Search Pattern**: Use `grep "Go.*1\." --include="*.md"` to find all version
  references in documentation
- **Files to Check**: go.mod, all .md files in docs/, specs/, examples/, and
  root-level documentation
- **Docker Images**: Update Docker base images (e.g., `golang:1.24` → `golang:1.25.5`)
  in documentation examples

### Systematic Version Updates

- **Process**: 1) Update go.mod first, 2) Find all references with grep, 3) Update
  each file systematically, 4) Verify with final grep search
- **Common Patterns**: Update both specific versions (1.24.10 → 1.25.5) and minimum
  requirements (Go 1.24+ → Go 1.25.5+)
- **CI Workflows**: Update GitHub Actions go-version parameters in documentation
  examples

This ensures complete version consistency across the entire codebase and documentation.

## AI Agent File Maintenance

This file (AGENTS.md) provides general repository guidelines for AI agents. To
maintain its effectiveness:

### Update Requirements

- **Review regularly** when repository structure or processes change
- **Update Go version references** immediately when versions change
- **Maintain current project structure** documentation
- **Keep build/test/lint commands** accurate and functional
- **Update coding standards** as conventions evolve
- **Refresh testing guidelines** when practices change

### When to Update

- Repository structure changes
- Go versions are updated
- Build or development commands change
- Coding standards evolve
- Testing practices change
- New security or configuration requirements emerge

### Integration with GitHub Copilot

- This file is automatically read by GitHub Copilot via `.github/instructions/ai-agent-files.instructions.md`
- Use it as the source of truth for repository guidelines
- Follow the documented standards and conventions
- Keep guidelines current for consistent AI assistance

### Maintenance Checklist

- [ ] Project structure documentation is accurate
- [ ] Build, test, and development commands work
- [ ] Go version information is current
- [ ] Coding style guidelines reflect current practices
- [ ] Testing guidelines are up to date
- [ ] Security and configuration tips remain relevant
- [ ] Commit and PR guidelines are followed

## Active Technologies

- Go 1.25.5 + existing CLI infrastructure, test helpers
- Go 1.25.5 + pluginsdk from pulumicost-spec
- Environment variable constants via pluginsdk
- Go 1.25.5 + charmbracelet/lipgloss v1.0.0, golang.org/x/term v0.37.0

## Recent Changes

- Added charmbracelet/lipgloss v1.0.0 and golang.org/x/term for TUI components
