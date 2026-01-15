# Feature Specification: Adopt pluginsdk/env.go for Environment Variable Handling

**Feature Branch**: `014-pluginsdk-env-adoption`  
**Created**: 2025-12-10  
**Status**: Draft  
**Input**: User description: "title: refactor: Adopt pluginsdk/env.go for environment variable handling
state: OPEN
author: rshade
labels:
comments: 0
assignees:
projects:
milestone:
number: 230
--

## Summary

Migrate `finfocus-core` to use the standardized `pluginsdk/env.go` module from `finfocus-spec` for all environment variable handling related to plugin communication. This ensures consistency with plugins and centralizes environment variable definitions.

## Background

During E2E testing, we discovered an environment variable mismatch between core and plugins:

- Core was setting `FINFOCUS_PLUGIN_PORT`
- Plugin SDK was reading `PORT`

A temporary fix was applied to set both variables, but the proper solution is to use shared constants from `pluginsdk/env.go`.

**Current code** (`internal/pluginhost/process.go`):

```go
// TODO: Replace with pluginsdk/env.go constants once available
cmd.Env = append(os.Environ(),
    fmt.Sprintf("PORT=%d", port),
    fmt.Sprintf("FINFOCUS_PLUGIN_PORT=%d", port),
)
```

## Requirements

### 1. Import and Use `pluginsdk/env.go` Constants

Update `internal/pluginhost/process.go`:

```go
import (
    "github.com/rshade/finfocus-spec/sdk/go/pluginsdk"
)

// Replace hardcoded strings with constants
cmd.Env = append(os.Environ(),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPortFallback, port),
    fmt.Sprintf("%s=%d", pluginsdk.EnvPort, port),
)
```

### 2. Update Code Generator

If there's a code generator for plugins (`cmd/gen` or similar), update it to:

- Import `pluginsdk/env.go`
- Generate code that uses `pluginsdk.GetPort()` instead of direct `os.Getenv()`
- Include comments referencing the best practices documentation

### 3. Update Other Environment Variable Usage

Search for and update any other environment variable usage:

- Logging configuration (`FINFOCUS_LOG_LEVEL`, `FINFOCUS_LOG_FORMAT`)
- Trace ID injection (`FINFOCUS_TRACE_ID`)

## Tasks

- [ ] Add `pluginsdk` import to `internal/pluginhost/process.go`
- [ ] Replace hardcoded env var strings with `pluginsdk.Env*` constants
- [ ] Update code generator to use `pluginsdk/env.go` in generated plugin code
- [ ] Search codebase for other env var usage and migrate to constants
- [ ] Update unit tests if any mock env var names
- [ ] Update CLAUDE.md with new patterns
- [ ] Add integration test verifying env var propagation

## Acceptance Criteria

- [ ] No hardcoded environment variable strings for plugin communication
- [ ] Code generator produces code using `pluginsdk.GetPort()`
- [ ] All tests pass
- [ ] Documentation updated

## Dependencies

- **Blocked by**: rshade/finfocus-spec#127 (pluginsdk/env.go implementation)

## Related Issues

- rshade/finfocus-spec#127 - Create pluginsdk/env.go
- rshade/finfocus-plugin-aws-public (issue TBD) - Plugin migration

## Labels

`refactor`, `dependencies`, `pluginsdk`"

## Clarifications

### Session 2025-12-10

- Q: What happens when pluginsdk/env.go is not available (dependency not installed)? → A: Feature blocked until dependency available - Add error handling for missing import

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Plugin Communication Consistency (Priority: P1)

As a plugin developer, I want the core system to use standardized environment variable names so that plugins can reliably read port information without hardcoded fallbacks.

**Why this priority**: This is the core issue that caused E2E testing failures and requires immediate resolution for plugin compatibility.

**Independent Test**: Can be tested by verifying that plugins receive the correct environment variables and can start successfully without manual configuration.

**Acceptance Scenarios**:

1. **Given** a plugin is launched by the core system, **When** the core sets environment variables using pluginsdk constants, **Then** the plugin can read the port using pluginsdk.GetPort() without errors
2. **Given** environment variable names change in pluginsdk, **When** core is updated to use new constants, **Then** plugins continue to work without code changes

---

### User Story 2 - Centralized Environment Variable Management (Priority: P2)

As a maintainer, I want all environment variable handling to use shared constants so that changes to variable names only require updates in one place.

**Why this priority**: Reduces maintenance burden and prevents inconsistencies across the codebase.

**Independent Test**: Can be tested by searching the codebase for hardcoded environment variable strings and verifying they are replaced with constants.

**Acceptance Scenarios**:

1. **Given** environment variable names need to change, **When** constants are updated in pluginsdk/env.go, **Then** all core code automatically uses the new names without individual file changes
2. **Given** new environment variables are added to pluginsdk, **When** core needs to use them, **Then** they are available as constants without manual string definitions

---

### User Story 3 - Code Generator Updates (Priority: P3)

As a plugin author using the code generator, I want generated code to use pluginsdk functions instead of direct os.Getenv() calls for better error handling and consistency.

**Why this priority**: Improves generated code quality and reduces boilerplate in plugins.

**Independent Test**: Can be tested by generating a new plugin and verifying it uses pluginsdk.GetPort() instead of os.Getenv("PORT").

**Acceptance Scenarios**:

1. **Given** the code generator is run, **When** it generates plugin code, **Then** the code uses pluginsdk.GetPort() with proper error handling
2. **Given** pluginsdk adds new environment variable functions, **When** the generator is updated, **Then** it uses the new functions automatically

---

### Edge Cases

- **Missing pluginsdk dependency**: Feature is blocked until finfocus-spec#127 is implemented and dependency is available. Implementation should include error handling for missing imports.
- How does the system handle environment variables that are not defined in pluginsdk?
- What if multiple environment variables serve the same purpose (like PORT and FINFOCUS_PLUGIN_PORT)?

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST import pluginsdk/env.go in internal/pluginhost/process.go
- **FR-002**: System MUST replace hardcoded "PORT" and "FINFOCUS_PLUGIN_PORT" strings with pluginsdk.EnvPort and pluginsdk.EnvPortFallback constants
- **FR-003**: System MUST search codebase for other hardcoded environment variable strings and replace with constants where available
- **FR-004**: System MUST update code generator (if exists) to import pluginsdk/env.go and use pluginsdk.GetPort() instead of os.Getenv()
- **FR-005**: System MUST update unit tests that mock environment variable names to use the new constants
- **FR-006**: System MUST add integration test that verifies environment variables are properly propagated to plugins
- **FR-007**: System MUST update CLAUDE.md with patterns for using pluginsdk environment variable constants

### Key Entities _(include if feature involves data)_

- **Environment Variables**: Standardized names and values for plugin communication, defined in pluginsdk/env.go
- **Plugin Process**: Child process launched by core with environment variables set
- **Code Generator**: Tool that generates plugin boilerplate code

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: All hardcoded environment variable strings for plugin communication are eliminated from the codebase
- **SC-002**: Code generator produces plugin code using pluginsdk.GetPort() instead of direct os.Getenv() calls
- **SC-003**: All existing tests pass after migration to constants
- **SC-004**: Integration test successfully verifies environment variable propagation between core and plugins
- **SC-005**: CLAUDE.md is updated with guidance for using pluginsdk environment variable patterns
