# Research: Plugin Info and DryRun Discovery

## Decision: gRPC "Unimplemented" Handling
**Decision**: Use `google.golang.org/grpc/status` to check for `codes.Unimplemented`.
**Rationale**: FR-004 requires graceful handling of legacy plugins that don't implement the new RPCs. Checking the gRPC status code is the standard way to detect missing methods.
**Alternatives considered**: Catching all errors, but that might hide real network/timeout issues.

## Decision: Spec Version Validation
**Decision**: Use `github.com/Masterminds/semver/v3` for compatibility checks.
**Rationale**: FR-005 requires "best-effort" compatibility. SemVer allows comparing major/minor versions to decide if a warning or failure is appropriate.
**Alternatives considered**: Manual string parsing, but SemVer is already a project dependency and more robust.

## Decision: CLI Command Structure
**Decision**: Implement `pulumicost plugin inspect <plugin> <resource-type>`.
**Rationale**: Aligns with `pulumicost plugin install/remove` style. It provides a clear way to trigger the `DryRun` RPC for a specific resource type.
**Alternatives considered**: Adding it to `plugin list`, but that would be too verbose and slow.

## Decision: Plugin Initialization Timeout
**Decision**: 5 seconds for `GetPluginInfo`, 10 seconds for `DryRun`.
**Rationale**: `GetPluginInfo` is expected to be fast (static metadata). `DryRun` might involve more logic (mapping introspection), so a longer timeout is safer.
**Alternatives considered**: Using a single global timeout, but specific timeouts per RPC type are more precise.

## Decision: Core Supported Spec Version
**Decision**: Core will use `pluginsdk.SpecVersion` as its baseline for compatibility checks.
**Rationale**: Since core imports the `pluginsdk`, it is inherently tied to that version of the spec.
**Alternatives considered**: Manually defining a version string in `pkg/version`, but this risks drift.
