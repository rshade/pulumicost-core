# Research: Remove PORT Environment Variable

**Feature**: 017-remove-port-env
**Date**: 2025-12-15

## Research Tasks

### 1. Current PORT Usage in Codebase

**Decision**: Remove `PORT` from `startPlugin()` function in `process.go`

**Findings**:
- `envPortFallback = "PORT"` constant defined at line 43 in `process.go`
- Used in `startPlugin()` at line 374: `fmt.Sprintf("%s=%d", envPortFallback, port)`
- Comments at lines 37-43 explicitly reference issue #232 and mark this for removal
- `--port` flag already passed at line 367: `append(args, fmt.Sprintf("--port=%d", port))`

**Rationale**: The `--port` flag is already the primary mechanism. PORT env var is redundant and causes inheritance conflicts.

**Alternatives Considered**:
- Keep both PORT and --port: Rejected - creates maintenance burden and collision risk
- Remove both and use only PULUMICOST_PLUGIN_PORT: Rejected - command-line flag is more explicit and debuggable

### 2. Test Impact Analysis

**Decision**: Update `TestProcessLauncher_StartPluginEnvironment` and `createEnvCheckingScript`

**Findings**:
- `process_test.go` lines 720-882 contain tests that explicitly verify PORT is set
- `createEnvCheckingScript()` creates scripts that fail if PORT is not set
- These tests will fail after PORT removal (intentional - validates the change)

**Required Test Updates**:
1. Modify `createEnvCheckingScript()` to verify PORT is NOT set
2. Update `TestProcessLauncher_StartPluginEnvironment` assertions
3. Add new test for DEBUG logging when PORT detected in environment

### 3. Environment Variable Filtering Strategy

**Decision**: Do not filter inherited PORT from os.Environ()

**Rationale**:
- FR-005 requires PORT not to interfere with plugin communication
- The plugin receives port via `--port` flag (authoritative)
- If user has PORT=3000 in environment, plugin should ignore it (plugin's responsibility via pluginsdk)
- Filtering os.Environ() would be invasive and could break other env var inheritance

**Alternative Considered**:
- Filter PORT from os.Environ(): Rejected - too invasive, breaks other use cases

### 4. Guidance Logging Implementation

**Decision**: Add guidance message in `waitForPluginBind` timeout error handling

**Findings**:
- `waitForPluginBind()` at line 315 handles timeout waiting for plugin bind
- Error path at line 219 logs "plugin failed to bind to port"
- This is the ideal location to add guidance suggesting plugin may need update

**Implementation Pattern**:
```go
log.Warn().
    Ctx(ctx).
    Str("component", "pluginhost").
    Int("port", port).
    Msg("plugin failed to bind - if using an older plugin, ensure it supports --port flag")
```

### 5. DEBUG Logging for PORT Detection

**Decision**: Add DEBUG log in `startPlugin()` when PORT is detected in environment

**Findings**:
- `startPlugin()` at line 362 is called before plugin spawn
- Can check `os.Getenv("PORT")` and log if non-empty
- DEBUG level ensures it's only visible with `--debug` flag

**Implementation Pattern**:
```go
if portEnv := os.Getenv("PORT"); portEnv != "" {
    log.Debug().
        Ctx(ctx).
        Str("component", "pluginhost").
        Str("inherited_port", portEnv).
        Msg("PORT environment variable detected in parent environment (will be ignored, plugin uses --port flag)")
}
```

### 6. External Dependency Verification

**Decision**: Verify pulumicost-spec#129 status before implementation

**Findings**:
- This feature is blocked by pulumicost-spec#129 (Add --port flag to pluginsdk.Serve())
- Plugin SDK must support `--port` flag parsing before core removes PORT
- Status check required before merging this PR

**Verification Steps**:
1. Check pulumicost-spec#129 is merged
2. Update pulumicost-spec dependency if needed
3. Verify pluginsdk.Serve() handles --port flag

## Summary

All research items resolved. No NEEDS CLARIFICATION items remain.

| Topic | Decision | Risk |
|-------|----------|------|
| PORT removal | Remove from startPlugin() | Low - --port flag already in place |
| Test updates | Update env-checking scripts | Low - tests validate behavior change |
| Env filtering | Don't filter inherited PORT | Low - plugin ignores via pluginsdk |
| Guidance logging | Add in waitForPluginBind | Low - non-breaking addition |
| DEBUG logging | Add in startPlugin | Low - non-breaking addition |
| External dependency | Verify spec#129 first | Medium - blocks merge |
