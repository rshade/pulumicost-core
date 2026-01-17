# Implementation Plan - CodeRabbit Issue Resolution

**Feature**: CodeRabbit Issue Resolution (`114-fix-coderabbit-issues`)
**Status**: Draft

## Technical Context

This feature addresses technical debt and code quality issues identified by CodeRabbit. The changes span multiple modules (`cli`, `engine`, `pluginhost`, `proto`) but are individually scoped and well-defined.

**Key Changes:**
1.  **Refactoring**: Consolidate `EnvAnalyzerMode` constant to prevent circular dependencies and duplication.
2.  **Error Handling**: Stop swallowing errors in `plugin_inspect.go` (close, write) and `plugin_list.go` (launch).
3.  **Standards**: Use centralized configuration for paths, enforce linting rules in `engine.go`.
4.  **Documentation**: Update `GEMINI.md`, `README.md`, and add GoDocs/Stringers for types.

**Component Interaction:**
- `cli` -> `internal/config`: For path resolution.
- `cli` -> `internal/constants`: New package for shared constants.
- `pluginhost` -> `internal/constants`: New package for shared constants.

## Constitution Check

### I. Plugin-First Architecture
- **Status**: Compliant. Changes improve plugin hosting reliability (error handling).

### II. Test-Driven Development
- **Status**: Compliant.
- **Plan**:
    - Add test for `CompatibilityResult.String()` output.
    - Update `Adapter` tests to use `require.Error`.
    - `cli` changes are hard to unit test without refactoring `main` but manual verification and existing integration tests cover them.

### III. Cross-Platform Compatibility
- **Status**: Compliant.
- **Plan**: Switch to `config.New().PluginDir` ensures platform-correct path handling (handling Windows vs Linux separators correctly).

### IV. Documentation Synchronization
- **Status**: Compliant.
- **Plan**: `README.md` and `GEMINI.md` are explicitly targeted for updates in this plan.

### VI. Implementation Completeness
- **Status**: Compliant. No stubs will be added. All error paths will be fully handled.

## Phase 0: Research & Validation

**Goal**: Verify current codebase state against issue reports. (Confirmed via grep).

**Research Tasks**:
- [x] Confirm duplicate `EnvAnalyzerMode` locations.
- [x] Verify `renderTable` signature.
- [x] Check `FieldMapping` struct tags.

## Phase 1: Design & Implementation

**Goal**: Apply fixes and refactors.

### 1. Shared Constants
- Create `internal/constants/env.go`.
- Move `EnvAnalyzerMode` there (use existing value `"FINFOCUS_ANALYZER_MODE"`).
- Update `cli` and `pluginhost` to import from `internal/constants`.

### 2. CLI Reliability
- **`plugin_inspect.go`**:
    - Refactor `defer client.Close()` to log errors at **Debug Level**.
    - Update `renderTable` to return `error`.
    - Check all writes (`fmt.Fprintf`, `w.Write`).
    - Update `findPluginPath` to use `config.PluginDir`.
- **`plugin_list.go`**:
    - Log launch errors at **Debug Level** before cancel.

### 3. Engine & Proto Quality
- **`engine.go`**: Add `//nolint:funlen,gocognit` directive.
- **`types.go`**: Add `omitempty` tags and GoDocs.
- **`version.go`**: Implement `String()` for `CompatibilityResult`.
- **`adapter_test.go`**: Use `require.Error`.

### 4. Documentation
- **`GEMINI.md`**: Deduplicate architecture info.
- **`README.md`**: Add `plugin inspect` output examples.

## Phase 2: Verification

**Goal**: Ensure no regressions and linting passes.

**Tasks**:
- Run `make lint` (must pass).
- Run `make test` (must pass).
- Manual verification of `plugin inspect` output.

## Gates & Reviews

- [ ] **Lint Check**: `golangci-lint` must be clean.
- [ ] **Test Check**: All tests pass.
- [ ] **Doc Check**: README contains new examples.