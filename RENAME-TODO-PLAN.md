# FinFocus Rename: Remaining Tasks

This document tracks items missed during the initial rebranding of `pulumicost` to `finfocus`.

## 1. Module Paths & Imports

- [x] Update `test/e2e/go.mod` module path from `github.com/rshade/pulumicost-core/test/e2e` to `github.com/rshade/finfocus/test/e2e`
- [x] Check for any other `go.mod` files (e.g., in `examples/`)
- [x] Global search for `github.com/rshade/pulumicost-core` in all files (especially `.go`, `.md`, `.yml`)
- [x] Global search for `github.com/rshade/pulumicost-spec` (should be `github.com/rshade/finfocus-spec`)

## 2. GitHub Infrastructure

- [x] Update `.github/workflows/` files (replace `pulumicost` with `finfocus` in names, paths, and binary references)
- [x] Update `.github/ISSUE_TEMPLATE/` (if any)
- [x] Update `.release-please-manifest.json`
- [x] Update `release-please-config.json`
- [x] Update `.goreleaser.yaml` (verify all occurrences)

## 3. Project Configuration & Metadata

- [x] Update `package.json` (name, repository, bugs, homepage)
- [x] Update `docker/Dockerfile` (binary name, user name, paths)
- [x] Update `docker/README.md`
- [x] Update `Makefile` (ensure all targets use `finfocus`)

## 4. Documentation & Branding

- [x] `README.md`: Full rebranding, update links
- [x] `CONTRIBUTING.md`: Update repo links and instructions
- [x] `ROADMAP.md`: Update project name and links
- [x] `CONTEXT.md`: Update boundaries and names
- [x] `AGENTS.md`: Update paths and names
- [x] `CHANGELOG.md`: Update repo links
- [x] `docs/`: Global replace `pulumicost` -> `finfocus` and `Pulumicost` -> `FinFocus`
- [x] `docs/_config.yml`: Update `baseurl` and `repository`

## 5. CLI & Internal Strings

- [x] `internal/cli/`: Search for hardcoded "pulumicost" strings in help texts, error messages, and examples
- [x] `internal/tui/`: Verify all headers/views use "FinFocus"
- [x] `internal/analyzer/`: Ensure handshake strings and metadata use "finfocus"

## 6. Tests & Examples

- [x] `test/e2e/`: Update project names in `Pulumi.yaml` files
- [x] `examples/`: Update `README.md` and `Pulumi.yaml` files
- [x] `testdata/`: Check if any test data contains "pulumicost" that should be updated

## 7. Configuration & State Defaults

- [x] Verify `internal/config/paths.go` (or wherever defaults are) uses `.finfocus`
- [x] Verify environment variable prefix in `internal/config/` (Viper setup)
- [x] Ensure migration logic is actually wired up in the root command

## 8. Registry & Plugins

- [x] `internal/registry/`: Update default registry URL if hardcoded
- [x] `plugins/recorder/`: Update manifest and README

---

## Progress Log

- [2026-01-14] Created RENAME-TODO-PLAN.md after finding 1531 missed occurrences of "pulumicost".
- [2026-01-14] Performed global string replacements for imports and branding.
- [2026-01-14] Fixed legacy compatibility logic in `internal/config` and `internal/registry` which was broken by global replace.
- [2026-01-14] Verified root metadata files (`README.md`, `CONTRIBUTING.md`, etc.) are rebranded.
- [2026-01-14] Updated `.claude` and `.gemini` agent configurations.
- [2026-01-14] Final verification check completed. All known occurrences of `pulumicost` (outside of RENAME-PLAN.md) have been addressed.
