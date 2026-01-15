# Implementation Plan - Pulumi Tool Plugin Integration

**Status**: In Progress
**Branch**: `110-pulumi-tool-plugin`
**Spec**: [spec.md](./spec.md)

## Technical Context

<!--
  ACTION REQUIRED: Identify all technical components, libraries, and unknowns.
  Mark unknowns as [NEEDS CLARIFICATION] to trigger Phase 0 research.
-->

### Architecture Components

- **CLI Entrypoint (`internal/cli/root.go`)**: Needs modification to detect invocation name (`os.Args[0]`) and environment variables (`FINFOCUS_PLUGIN_MODE`).
- **Configuration Loader (`internal/config/loader.go`)**: Needs update to respect `PULUMI_HOME` when in plugin mode.
- **Help/Usage Generator (`cobra` integration)**: Needs to dynamically adjust `Use` string and examples based on mode.
- **Build System (`Makefile` / `.goreleaser.yaml`)**: Needs to support building the binary as `pulumi-tool-cost`.

### Libraries & Dependencies

- **`github.com/spf13/cobra`**: Existing dependency, used for CLI structure.
- **`github.com/spf13/pflag`**: Existing dependency, used for flags.
- **`os` package**: Standard library for arg and env var detection.
- **`path/filepath`**: Standard library for path manipulation.

### Unknowns & Risks

- **[NEEDS CLARIFICATION]**: Does Pulumi *require* specific exit codes for tool plugins beyond standard 0/1? The spec mentions "Pulumi-First" but are there specific codes documented for tools specifically?
- **[NEEDS CLARIFICATION]**: How does `PULUMI_HOME` interact with XDG standards on Linux? Does Pulumi enforce one over the other in its own logic that we should mimic?

## Constitution Check

<!--
  ACTION REQUIRED: Review the Constitution (.specify/memory/constitution.md).
  Confirm compliance or flag violations.
-->

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Plugin-First** | ✅ Compliant | This feature enables the core to *be* a plugin itself, doesn't violate plugin architecture. |
| **II. TDD** | ✅ Compliant | Plan includes tests for CLI mode detection and config loading. |
| **III. Cross-Platform** | ✅ Compliant | `path/filepath` and OS-agnostic checks will be used. |
| **IV. Docs as Code** | ✅ Compliant | Plan includes updating usage docs. |
| **V. Protocol Stability** | N/A | No changes to gRPC protocol. |
| **VI. Completeness** | ✅ Compliant | Implementation will be full, no stubs. |

## Gates & Checks

<!--
  ACTION REQUIRED: Evaluate project-specific gates.
  Stop if any fail.
-->

- [x] **Spec Completeness**: Spec has all sections and no [NEEDS CLARIFICATION].
- [x] **Impact Analysis**: Low risk, primarily CLI wrapper logic changes.
- [x] **Research Complete**: Phase 0 complete. See [research.md](./research.md).

## Phase 0: Research & Decisions

<!--
  ACTION REQUIRED:
  1. Research all [NEEDS CLARIFICATION] items from Technical Context.
  2. Document decisions in research.md.
  3. Update this section with summary of findings.
-->

**Summary of Decisions**:
1.  **Exit Codes**: Adopt standard POSIX (0/1). No special Pulumi codes required for tools.
2.  **Config Path**: `PULUMI_HOME/finfocus/` takes precedence if `PULUMI_HOME` is set. Fallback to XDG or Home directory otherwise.

## Phase 1: Design Artifacts

- [x] **Data Model**: [data-model.md](./data-model.md) (Config & Plugin Context)
- [x] **Contracts**: [contracts/cli.md](./contracts/cli.md) (CLI Signals & Env Vars)
- [x] **Quickstart**: [quickstart.md](./quickstart.md) (Local Dev Guide)

## Phase 2: Task Decomposition

<!--
  ACTION REQUIRED:
  Break down the work into TDD-compatible tasks.
  Use the checklist agent to generate the final list.
-->

### 1. Build & Infrastructure
- [ ] **chore**: Add `make build-plugin` target to Makefile to produce `pulumi-tool-cost` binary.
- [ ] **chore**: Update `.gitignore` to ignore `pulumi-tool-cost`.

### 2. Core Logic (TDD)
- [ ] **feat**: Implement `DetectPluginMode()` in `internal/cli/util.go` (or similar) with unit tests for:
    - Binary name matching (`pulumi-tool-cost`, case-insensitive).
    - Env var `FINFOCUS_PLUGIN_MODE`.
- [ ] **feat**: Update `internal/config` loader to check `PULUMI_HOME` env var.
    - Test: If `PULUMI_HOME` set, config path is `$PULUMI_HOME/finfocus/`.
    - Test: If unset, fallback to default.

### 3. CLI Integration
- [ ] **feat**: Update `root.go` to use `DetectPluginMode()`.
    - Dynamically set `Use` string.
    - Dynamically update `Example` strings in help text.

### 4. Documentation
- [ ] **docs**: Add "Running as Pulumi Plugin" section to `docs/user-guide.md`.