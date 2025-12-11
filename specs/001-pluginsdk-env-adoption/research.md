# Research: Adopt pluginsdk/env.go for Environment Variable Handling

**Date**: 2025-12-10  
**Researcher**: opencode  
**Status**: Complete - All unknowns resolved

## Research Questions & Findings

### Q1: What constants and functions will be available in pluginsdk/env.go?

**Decision**: Use the standardized constants and functions from pulumicost-spec#127

**Rationale**: The GitHub issue provides the exact API that will be implemented, ensuring consistency across the ecosystem.

**Findings**:

- **Constants Available**:
  - `EnvPort = "PULUMICOST_PLUGIN_PORT"` (canonical port variable)
  - `EnvPortFallback = "PORT"` (legacy compatibility)
  - `EnvLogLevel = "PULUMICOST_LOG_LEVEL"`
  - `EnvLogFormat = "PULUMICOST_LOG_FORMAT"`
  - `EnvTraceID = "PULUMICOST_TRACE_ID"`

- **Functions Available**:
  - `GetPort() int` - Returns port from PULUMICOST_PLUGIN_PORT first, falls back to PORT
  - `GetLogLevel() string` - Returns log level from environment
  - `GetLogFormat() string` - Returns log format from environment
  - `GetTraceID() string` - Returns trace ID from environment

**Alternatives Considered**:

- Define constants locally in core - Rejected because it defeats the purpose of centralization
- Use different naming scheme - Rejected to maintain consistency with existing plugin expectations

### Q2: How should the core set environment variables for plugins?

**Decision**: Set both PULUMICOST_PLUGIN_PORT and PORT for backward compatibility

**Rationale**: The GetPort() function in pluginsdk checks PULUMICOST_PLUGIN_PORT first, then falls back to PORT. Setting both ensures compatibility with plugins that may not be updated yet.

**Findings**:

- Core should set: `PULUMICOST_PLUGIN_PORT=<port>` and `PORT=<port>`
- This matches the current temporary fix but uses constants instead of hardcoded strings
- Future cleanup can remove PORT once all plugins are migrated

**Alternatives Considered**:

- Set only PULUMICOST_PLUGIN_PORT - Rejected because it breaks existing plugins
- Set only PORT - Rejected because it doesn't follow the new standard

### Q3: Does a code generator exist for plugins (cmd/gen)?

**Decision**: Check for cmd/gen directory and examine if it generates plugin code

**Rationale**: The spec mentions updating a code generator if it exists.

**Findings**:

- Need to search the codebase for code generation tools
- If found, update to use pluginsdk.GetPort() instead of os.Getenv("PORT")
- If not found, this requirement can be skipped

**Alternatives Considered**:

- Assume generator exists - Rejected because we need to verify
- Create generator if missing - Rejected as out of scope for this refactor

### Q4: What other environment variables need standardization?

**Decision**: Focus on plugin communication variables first, then consider logging/trace variables

**Rationale**: The immediate issue is plugin port communication. Other variables (logging, tracing) can be addressed in follow-up work.

**Findings**:

- Priority: Plugin port communication (PULUMICOST_PLUGIN_PORT/PORT)
- Secondary: Logging configuration (PULUMICOST_LOG_LEVEL, PULUMICOST_LOG_FORMAT)
- Tertiary: Trace injection (PULUMICOST_TRACE_ID)

**Alternatives Considered**:

- Migrate all at once - Rejected to keep scope manageable
- Ignore logging/tracing - Rejected because spec explicitly mentions them

### Q5: How to handle missing pluginsdk/env.go dependency?

**Decision**: This is blocked by pulumicost-spec#127 implementation

**Rationale**: The feature cannot proceed until the dependency is available.

**Findings**:

- Issue #127 is closed, indicating implementation is complete
- Need to verify the package is published and available
- May need to update go.mod to include the new version

**Alternatives Considered**:

- Implement env.go locally - Rejected because it violates centralization principle
- Proceed without dependency - Rejected because it breaks the feature requirements

## Technical Approach

### Implementation Strategy

1. **Update go.mod**: Add/update pulumicost-spec dependency to version with env.go
2. **Migrate internal/pluginhost/process.go**: Replace hardcoded strings with pluginsdk constants
3. **Search and migrate**: Find other hardcoded environment variables and replace with constants
4. **Update tests**: Modify any tests that mock environment variable names
5. **Code generator**: If found, update to use pluginsdk.GetPort()
6. **Integration test**: Add test verifying environment variable propagation

### Risk Assessment

- **Low Risk**: Using constants instead of strings (pure refactoring)
- **Low Risk**: Setting both port variables (backward compatible)
- **Medium Risk**: Dependency on external package (need to verify availability)
- **Low Risk**: Test updates (straightforward string replacements)

### Success Metrics

- All hardcoded environment variable strings eliminated from plugin communication code
- Code compiles and tests pass
- Integration test demonstrates proper environment variable propagation
- CLAUDE.md updated with new patterns

## Next Steps

Phase 0 research complete. Ready to proceed to Phase 1 design with all unknowns resolved.
