# Feature Specification: CodeRabbit Issue Resolution

**Feature Branch**: `114-fix-coderabbit-issues`  
**Created**: 2026-01-16  
**Status**: Draft  
**Input**: User description: "fix coderabbit issues from 398..." (Detailed list of refactoring, error handling, and documentation tasks)

## Clarifications

### Session 2026-01-16

- Q: What log level should be used for plugin launch failures in plugin_list? → A: Debug Level
- Q: What should be the name of the shared package for cross-cutting constants? → A: internal/constants

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reliable Error Handling & Logging (Priority: P1)

As a System Operator, I need the CLI and Plugin system to reliably report internal errors (like connection closures or write failures) so that I can troubleshoot issues without silent failures.

**Why this priority**: Silent failures in plugin management and output rendering mask critical issues, making debugging impossible and data potentially unreliable.

**Independent Test**: Can be tested by simulating errors (e.g., closed connections, full disks) and verifying they are logged/returned.

**Acceptance Scenarios**:

1. **Given** a plugin inspection session, **When** the client connection closes, **Then** any errors during closure are logged (debug level) rather than ignored.
2. **Given** a plugin inspection output, **When** writing to the output stream fails (e.g., pipe closed), **Then** the application returns an error immediately.
3. **Given** a plugin listing command, **When** a plugin fails to launch, **Then** the specific launch error is logged before the operation is cancelled.
4. **Given** a test server shutdown, **When** `Serve()` returns, **Then** the code explicitly handles or documents the expected `ErrServerStopped` behavior.

---

### User Story 2 - Code Consistency & Maintainability (Priority: P2)

As a Developer, I want the codebase to use centralized constants and standard directory paths so that the system behavior is consistent across modules and easier to maintain.

**Why this priority**: Duplicate constants and hardcoded paths lead to drift where one part of the system uses different configuration than another.

**Independent Test**: Can be verified by code inspection and compilation (ensuring no duplicate definitions exist) and runtime checks of path resolution.

**Acceptance Scenarios**:

1. **Given** the application runs in Analyzer Mode, **When** the mode is checked, **Then** it uses the single canonical constant from the shared package.
2. **Given** the CLI resolves a plugin path, **When** looking for the binary, **Then** it uses the centralized configuration directory, not a hardcoded user home path.
3. **Given** cost calculation workers are compiled, **When** linting is run, **Then** complex worker functions have appropriate lint suppression comments matching the project standard.

---

### User Story 3 - Enhanced Documentation & Usability (Priority: P3)

As a New User, I want clear documentation and human-readable log outputs so that I can quickly understand how to use the CLI and interpret its internal state.

**Why this priority**: Improving the README and log formats reduces friction for adoption and debugging.

**Independent Test**: Can be tested by viewing the README and running commands with verbose logging.

**Acceptance Scenarios**:

1. **Given** the README file, **When** a user reads the "plugin inspect" section, **Then** they see a concrete example of the command output (Table and JSON).
2. **Given** verbose logs, **When** a compatibility check result is logged, **Then** it appears as a human-readable string (e.g., "Compatible") instead of a raw integer.
3. **Given** the project documentation (GEMINI.md), **When** checking architecture, **Then** the "Stateless" nature is described exactly once to avoid redundancy.

### Edge Cases

- **Write Failures**: partial writes to stdout/stderr during table rendering.
- **Path Resolution**: Windows vs Linux path separators when joining centralized config paths.
- **Serialization**: Empty fields in `FieldMapping` structure should not appear in JSON/YAML output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST log errors encountered when closing plugin client connections (Debug level).
- **FR-002**: System MUST propagate write errors from Table rendering functions to the caller.
- **FR-003**: System MUST use the centralized configuration directory for resolving plugin paths, avoiding hardcoded user home paths.
- **FR-004**: System MUST log plugin launch failures (Path + Error) at Debug level before cancelling the operation context.
- **FR-005**: System MUST use a single canonical constant for Analyzer Mode configuration defined in `internal/constants` to prevent definition duplication.
- **FR-006**: System MUST represent Compatibility Results as human-readable strings ("Compatible", "MajorMismatch", "Invalid") in logs and debug output.
- **FR-007**: System MUST omit empty conditional and type fields from Field Mapping serialization (JSON/YAML).
- **FR-008**: System documentation (GEMINI.md) MUST have a single, consolidated reference to the application's stateless nature.
- **FR-009**: System README MUST include examples for the plugin inspection command showing both Table and JSON output formats.
- **FR-010**: System tests for Adapter MUST use fatal assertions for error validation cases.
- **FR-011**: System code MUST include documentation comments for Field Mapping Status types and constants.
- **FR-012**: System test code MUST explicitly document or handle expected server shutdown errors to avoid ambiguous empty blocks.
- **FR-013**: Cost calculation worker functions MUST comply with linter complexity rules via appropriate documented suppressions.

### Key Entities

- **CompatibilityResult**: Enum indicating plugin version compatibility.
- **FieldMapping**: Configuration structure for mapping resource fields.
- **AnalyzerMode**: Global configuration setting.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 13 specific CodeRabbit issues listed in the input are resolved (verified by code review).
- **SC-002**: Build process passes with zero linting errors (after adding suppressions).
- **SC-003**: `plugin inspect` command correctly reports errors when piped to a closed stream (exit code non-zero).
- **SC-004**: No duplicate `EnvAnalyzerMode` constants exist in the codebase.