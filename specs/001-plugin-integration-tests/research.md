# Research: Plugin Integration Tests

**Feature**: Integration Tests for Plugin Management (Init, Install, Update, Remove)
**Status**: Research Complete

## 1. Testing Framework & Pattern

**Decision**: Use Go's standard `testing` package with `github.com/stretchr/testify` for assertions, matching existing integration tests in `test/integration/plugin/`.

**Rationale**:
- `testify/assert` and `testify/require` are already used extensively in the codebase.
- Existing tests (e.g., `plugin_communication_test.go`) follow this pattern.
- Ensures consistency and readability.

## 2. Mocking Strategy

**Decision**: Use `httptest.NewServer` to mock the GitHub Registry and inject it via `registry.NewInstallerWithClient`.

**Rationale**:
- **Requirement**: Tests must be sandboxed and offline.
- **Mechanism**: The `internal/registry` package exposes `NewInstallerWithClient`, which accepts a `*registry.GitHubClient`.
- **Implementation**:
  - `GitHubClient` has a `BaseURL` field that can be pointed to the `httptest` server URL.
  - The mock server will serve JSON metadata (simulating GitHub Releases API) and binary artifacts (tar.gz/zip) from a local fixture directory.
  - This allows testing `install`, `update`, and failure scenarios (404, 500) deterministically.

## 3. Concurrency Handling

**Observation**: Current implementation in `internal/registry/installer.go` does *not* appear to implement explicit file locking (mutex/flock).

**Decision**: The test suite will include a concurrency test case (`TestPluginInstall_Concurrent`).

**Implication**:
- If the current implementation fails this test (likely), we will need to implement a file-based locking mechanism (e.g., using a `.lock` file or `syscall.Flock` if cross-platform support allows) in `internal/registry` as part of this feature to meet the spec requirements.
- *Note*: Given the "fully implemented" claim in the prompt, there's a chance I missed something, but `grep` showed no locking. I will proceed assuming implementation might be needed.

## 4. Test Isolation

**Decision**: Use `t.TempDir()` for all file system operations.

**Rationale**:
- Prevents tests from touching the user's actual `~/.pulumicost` directory.
- `registry.NewInstaller` accepts a `pluginDir` argument, which we will populate with `t.TempDir()`.
- Environment variables (e.g., `PULUMICOST_PLUGIN_DIR`) will also be set to this temp dir for commands that rely on env vars.

## 5. Mock Registry Protocol

**Decision**: The mock server will serve minimal JSON mimicking GitHub's API:
- `GET /repos/{owner}/{repo}/releases/latest`
- `GET /repos/{owner}/{repo}/releases/tags/{tag}`
- Binary downloads at arbitrary paths defined in the JSON.

**Rationale**: The `registry` package relies on GitHub's API structure. Emulating this is necessary for the `GitHubClient` to function correctly without modification.
