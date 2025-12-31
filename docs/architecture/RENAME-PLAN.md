# Engineering Plan: Project Rename to FinFocus

**Target Identity:** `FinFocus`
**CLI Command:** `finfocus` (User Alias: `fin`)
**Version Target:** v0.2.0 (Breaking Change)

## Executive Summary

This document outlines the engineering steps required to rename the `pulumicost` ecosystem to `finfocus`. This is a **breaking change** across all repositories. The goal is to establish a professional, enterprise-grade brand identity aligned with the FinOps FOCUS specification.

## impacted Repositories

| Repository | Current Role | New Name | Impact Level |
| :--- | :--- | :--- | :--- |
| `pulumicost-spec` | Protocol Definitions | `finfocus-spec` | **CRITICAL** (Root dependency) |
| `pulumicost-core` | CLI & Orchestrator | `finfocus-core` | **HIGH** (Main logic) |
| `pulumicost-plugin-*` | Cloud Providers | `finfocus-plugin-*` | **HIGH** (Re-compile required) |
| `pulumicost-mcp` | AI Integration | `finfocus-mcp` | **MEDIUM** |
| `homebrew-pulumicost` | Distribution | `homebrew-finfocus` | **LOW** |

---

## Phase 1: The Specification (`pulumicost-spec`)

*The foundation. Must be completed first.*

### 1.1 Protobuf Definition
*   **Package Rename:** Change `package pulumicost.v1;` to `package finfocus.v1;` in all `.proto` files.
*   **Service Rename:** Change `service PulumicostPlugin` to `service FinFocusPlugin`.
*   **Go Option:** Update `option go_package = "github.com/rshade/finfocus-spec/proto;finfocusv1";`.

### 1.2 Module Identity
*   **Go Module:** `go mod edit -module github.com/rshade/finfocus-spec`
*   **Release:** Tag `v0.1.0` of the new module.

---

## Phase 2: The Core (`pulumicost-core`)

*The heavy lifting.*

### 2.1 Dependency Update
*   Update `go.mod` to require `github.com/rshade/finfocus-spec`.
*   Run global find/replace for imports.
*   Regenerate any local mocks.

### 2.2 CLI & Branding
*   **Binary Name:** Update `Makefile`, `.goreleaser.yaml`, and `package.json` to output `finfocus`.
*   **Cobra Command:** Update `root.go` to use `finfocus`.
*   **TUI:** Update headers in `internal/tui/model.go` to display "FinFocus".
*   **User Alias:** Add post-install message:
    > "Recommended: Add `alias fin=finfocus` to your shell profile."

### 2.3 Configuration & State Paths
*   **Directory:** Switch default from `~/.pulumicost` to `~/.finfocus`.
*   **Env Vars:**
    *   `PULUMICOST_HOME` → `FINFOCUS_HOME`
    *   `PULUMICOST_LOG_LEVEL` → `FINFOCUS_LOG_LEVEL`
*   **Migration (Optional but nice):** On startup, if `~/.finfocus` is missing but `~/.pulumicost` exists, log a warning or offer to copy config.

### 2.4 Plugin Discovery
*   **Path:** Look for plugins in `~/.finfocus/plugins`.
*   **Naming Convention:** Expect executables named `finfocus-plugin-<name>`.
*   **Registry:** Update the default registry URL to `raw.githubusercontent.com/rshade/finfocus-registry/...`.

### 2.5 Pulumi Analyzer Integration
*   **Magic Name:** The binary must be capable of being invoked as `pulumi-analyzer-finfocus`.
*   **Logic:** Update `cmd/finfocus/main.go` to detect this executable name for gRPC handshake mode.

---

## Phase 3: The Plugins (aws-public, etc.)

*The ecosystem update.*

### 3.1 Re-branding
*   **Module:** `github.com/rshade/finfocus-plugin-aws-public`.
*   **Imports:** Update to `finfocus-spec`.
*   **Code Generation:** Re-run `protoc` against new `finfocus.v1` definitions.

### 3.2 Executable
*   Output binary as `finfocus-plugin-aws-public`.

---

## Phase 4: Downstream & Docs

### 4.1 MCP Server
*   Update to use `finfocus` CLI commands.
*   Update tool descriptions to refer to "FinFocus".

### 4.2 Documentation
*   Global string replacement `Pulumicost` → `FinFocus`.
*   Update installation code blocks.
*   Update `GEMINI.md` and `CLAUDE.md` context files.

---

## Migration Checklist

- [ ] **Step 1:** Rename & Release `finfocus-spec`.
- [ ] **Step 2:** Refactor `finfocus-core` to use new spec.
- [ ] **Step 3:** Implement config directory migration logic.
- [ ] **Step 4:** Verify `finfocus cost` works with new paths.
- [ ] **Step 5:** Rename & Release `finfocus-plugin-aws-public`.
- [ ] **Step 6:** End-to-End test: `finfocus plugin install aws-public`.
- [ ] **Step 7:** Update Docs & Readme.
