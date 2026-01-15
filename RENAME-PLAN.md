# Engineering Plan: Project Rename to FinFocus

**Target Identity:** `FinFocus`
**CLI Command:** `finfocus` (User Alias: `fin`)
**Version Target:** v0.2.0 (Breaking Change)
**Last Updated:** 2026-01-13

## Executive Summary

This document outlines the engineering steps required to rename the `pulumicost` ecosystem to `finfocus`.
This is a **breaking change** across all repositories. The goal is to establish a professional,
enterprise-grade brand identity aligned with the FinOps FOCUS specification.

## Current Status

| Phase | Status | Notes |
| :--- | :--- | :--- |
| Phase 1: Specification | ✅ **COMPLETE** | `finfocus-spec` v0.5.1 released |
| Phase 2: Core | ✅ **COMPLETE** | Implemented in 113-rebrand-to-finfocus |
| Phase 3: Plugins | ⏳ Pending | Blocked on Phase 2 |
| Phase 4: Downstream | ⏳ Pending | Blocked on Phase 2 |

## Impacted Repositories

| Repository | Current Name | New Name | Impact Level | Status |
| :--- | :--- | :--- | :--- | :--- |
| Protocol Definitions | `pulumicost-spec` | `finfocus-spec` | **CRITICAL** | ✅ Complete (v0.5.1) |
| CLI & Orchestrator | `pulumicost-core` | `finfocus` | **HIGH** | 🔄 In Progress |
| Cloud Providers | `pulumicost-plugin-*` | `finfocus-plugin-*` | **HIGH** | ⏳ Pending |
| AI Integration | `pulumicost-mcp` | `finfocus-mcp` | **MEDIUM** | ⏳ Pending |
| Distribution | `homebrew-pulumicost` | `homebrew-finfocus` | **LOW** | ⏳ Pending |

---

## Phase 1: The Specification (`pulumicost-spec`) ✅ COMPLETE

*The foundation. Completed and released as `finfocus-spec` v0.5.1.*

### 1.1 Protobuf Definition ✅

- ✅ **Package Rename:** Changed `package pulumicost.v1;` to `package finfocus.v1;` in all `.proto` files.
- ✅ **Service Rename:** Changed `service PulumicostPlugin` to `service CostSourceService`.
- ✅ **Go Option:** Updated `option go_package = "github.com/rshade/finfocus-spec/sdk/go/proto/finfocus/v1;pbc";`.

### 1.2 Module Identity ✅

- ✅ **Go Module:** `github.com/rshade/finfocus-spec`
- ✅ **Release:** Tagged `v0.5.1` with full proto and SDK support.

---

## Phase 2: The Core (`pulumicost-core` → `finfocus`)

*The heavy lifting. Repository will be renamed from `pulumicost-core` to `finfocus`.*

**New Module Path:** `github.com/rshade/finfocus`

### 2.1 Dependency Update

- [ ] Update `go.mod` module path from `github.com/rshade/pulumicost-core` to `github.com/rshade/finfocus`.
- [ ] Update `go.mod` to require `github.com/rshade/finfocus-spec` v0.5.1+.
- [ ] Run global find/replace for imports:
  - `github.com/rshade/pulumicost-core` → `github.com/rshade/finfocus`
  - `github.com/rshade/pulumicost-spec` → `github.com/rshade/finfocus-spec`
- [ ] Regenerate any local mocks.

### 2.2 CLI & Branding

- [ ] **Binary Name:** Update `Makefile`, `.goreleaser.yaml`, and `package.json` to output `finfocus`.
- [ ] **Cobra Command:** Update `root.go` to use `finfocus`.
- [ ] **TUI:** Update headers in `internal/tui/` to display "FinFocus".
- [ ] **User Alias:** Add post-install message:
  > "Recommended: Add `alias fin=finfocus` to your shell profile."

### 2.3 Configuration & State Paths

- [ ] **Directory:** Switch default from `~/.pulumicost` to `~/.finfocus`.
- [ ] **Env Vars:** Rename all environment variables:
  - `PULUMICOST_HOME` → `FINFOCUS_HOME`
  - `PULUMICOST_LOG_LEVEL` → `FINFOCUS_LOG_LEVEL`
  - `PULUMICOST_LOG_FORMAT` → `FINFOCUS_LOG_FORMAT`
  - `PULUMICOST_LOG_FILE` → `FINFOCUS_LOG_FILE`
  - `PULUMICOST_TRACE_ID` → `FINFOCUS_TRACE_ID`
  - `PULUMICOST_OUTPUT_FORMAT` → `FINFOCUS_OUTPUT_FORMAT`
  - `PULUMICOST_CONFIG_STRICT` → `FINFOCUS_CONFIG_STRICT`
  - `PULUMICOST_PLUGIN_PORT` → `FINFOCUS_PLUGIN_PORT`
  - `PULUMICOST_CONCURRENCY_MULTIPLIER` → `FINFOCUS_CONCURRENCY_MULTIPLIER`
- [ ] **Migration Logic:** On startup, if `~/.finfocus` is missing but `~/.pulumicost` exists,
  log a warning and offer to migrate config.

### 2.4 Plugin Discovery

- [ ] **Path:** Look for plugins in `~/.finfocus/plugins`.
- [ ] **Naming Convention:** Expect executables named `finfocus-plugin-<name>`.
- [ ] **Registry:** Update the default registry URL to `raw.githubusercontent.com/rshade/finfocus-registry/...`.

### 2.5 Pulumi Analyzer Integration

- [ ] **Magic Name:** The binary must be capable of being invoked as `pulumi-analyzer-finfocus`.
- [ ] **Logic:** Update `cmd/finfocus/main.go` to detect this executable name for gRPC handshake mode.

### 2.6 GitHub Repository Rename

- [ ] Rename repository from `pulumicost-core` to `finfocus` on GitHub.
- [ ] Update all GitHub Actions workflows.
- [ ] Update issue/PR templates.
- [ ] Update branch protection rules.

---

## Phase 3: The Plugins (aws-public, etc.)

*The ecosystem update. Blocked on Phase 2 completion.*

### 3.1 Re-branding

- [ ] **Module:** `github.com/rshade/finfocus-plugin-aws-public`.
- [ ] **Imports:** Update to `finfocus-spec` v0.5.1+.
- [ ] **Code Generation:** Re-run `protoc` against new `finfocus.v1` definitions.

### 3.2 Executable

- [ ] Output binary as `finfocus-plugin-aws-public`.

---

## Phase 4: Downstream & Docs

### 4.1 MCP Server

- [ ] Update to use `finfocus` CLI commands.
- [ ] Update tool descriptions to refer to "FinFocus".

### 4.2 Documentation

- [ ] Global string replacement `pulumicost` → `finfocus` and `Pulumicost` → `FinFocus`.
- [ ] Update installation code blocks.
- [ ] Update `GEMINI.md` and `CLAUDE.md` context files.
- [ ] Update GitHub Pages URL from `rshade.github.io/pulumicost-core` to `rshade.github.io/finfocus`.

---

## Migration Checklist

- [x] **Step 1:** Rename & Release `finfocus-spec` (v0.5.1).
- [ ] **Step 2:** Update `go.mod` to use new module path `github.com/rshade/finfocus`.
- [ ] **Step 3:** Update all imports and references.
- [ ] **Step 4:** Update CLI branding (binary name, Cobra commands, TUI).
- [ ] **Step 5:** Update configuration paths and environment variables.
- [ ] **Step 6:** Implement config directory migration logic.
- [ ] **Step 7:** Verify `finfocus cost` works with new paths.
- [ ] **Step 8:** Rename GitHub repository.
- [ ] **Step 9:** Rename & Release `finfocus-plugin-aws-public`.
- [ ] **Step 10:** End-to-End test: `finfocus plugin install aws-public`.
- [ ] **Step 11:** Update Docs & Readme.
- [ ] **Step 12:** Update MCP server.

---

## Breaking Changes Summary

Users upgrading from `pulumicost` to `finfocus` will need to:

1. **Reinstall the CLI:** `go install github.com/rshade/finfocus/cmd/finfocus@latest`
2. **Migrate config:** Copy `~/.pulumicost/` to `~/.finfocus/` (or use automatic migration)
3. **Update environment variables:** Replace `PULUMICOST_*` with `FINFOCUS_*`
4. **Reinstall plugins:** Plugins must be reinstalled with new names
5. **Update shell aliases:** Change `pulumicost` to `finfocus` (recommend `alias fin=finfocus`)
6. **Update Pulumi analyzer:** Rename symlink from `pulumi-analyzer-pulumicost` to `pulumi-analyzer-finfocus`

---

## Rollback Plan

If critical issues are discovered post-rename:

1. The old `pulumicost-spec` module remains available (archived).
2. GitHub redirects from `pulumicost-core` to `finfocus` will work automatically.
3. Environment variable fallbacks can be added to support both naming schemes temporarily.
