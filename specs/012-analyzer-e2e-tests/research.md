# Research: Analyzer E2E Tests

**Feature**: Analyzer E2E Tests for Pulumi Analyzer Plugin Integration
**Date**: 2025-12-08
**Status**: Completed

## Decision Log

### 1. Pulumi Analyzer Plugin Configuration Strategy

**Context**: We need to configure Pulumi to use the locally-built `finfocus` binary as an analyzer plugin during `pulumi preview`.

**Decision**: Use the `plugins.analyzers` section in `Pulumi.yaml` to reference the local binary path.

**Rationale**:

- The `plugins` configuration is specifically "intended for use in developing pulumi plugins" per Pulumi documentation
- No need for `--policy-pack` flag - analyzers configured in `Pulumi.yaml` are automatically invoked
- Path can be absolute or relative to the project directory
- Version field is optional and "will match any version the engine requests" if omitted

**Configuration Format**:

```yaml
name: finfocus-analyzer-e2e
runtime: yaml
description: E2E test project for analyzer plugin validation

plugins:
  analyzers:
    - name: finfocus
      path: /absolute/path/to/bin/finfocus
      version: 0.0.0-dev
```

**Alternatives Considered**:

- `--policy-pack` flag: Not applicable - that's for policy packs, not analyzer plugins
- Installing via `pulumi plugin install`: Would require published plugin, not local development
- Environment variables: `PULUMI_DEV=true` enables dev features but doesn't configure analyzers

### 2. Test Fixture Project Design

**Context**: We need to validate that the analyzer produces accurate cost diagnostics with real pricing data.

**Decision**: Use real AWS resources with the `aws-public` plugin, consistent with existing E2E test patterns.

**Rationale**:

- Only way to guarantee analyzer output matches real pricing data
- Validates the complete chain: Pulumi → Analyzer → Engine → Plugin → AWS Pricing
- Consistent with existing E2E infrastructure (`test/e2e/projects/ec2/`)
- AWS credentials already configured in `nightly.yml` for existing E2E tests

**Test Project Structure**:

```yaml
name: finfocus-analyzer-e2e
runtime: yaml
description: E2E test for analyzer plugin with real AWS resources

plugins:
  analyzers:
    - name: finfocus
      path: ${FINFOCUS_BINARY}

resources:
  # Real EC2 instance for accurate cost validation
  test-instance:
    type: aws:ec2:Instance
    properties:
      instanceType: t3.micro
      ami: ${ami.id}
      tags:
        Name: analyzer-e2e-test

variables:
  ami:
    fn::invoke:
      function: aws:ec2/getAmi:getAmi
      arguments:
        mostRecent: true
        owners:
          - amazon
        filters:
          - name: name
            values:
              - amzn2-ami-hvm-*-x86_64-gp2
```

**Why Real AWS Resources**:

- Validates actual cost calculations (~$7.59/month for t3.micro)
- Ensures plugin integration works end-to-end
- Matches existing E2E test patterns for consistency
- AWS credentials already available in CI (nightly.yml)

**Alternatives Rejected**:

- Mock/random resources: Cannot validate real pricing accuracy
- Simulated costs: Defeats purpose of E2E testing

### 3. Analyzer Output Validation Strategy

**Context**: We need to verify that analyzer diagnostics appear correctly in `pulumi preview` output.

**Decision**: Parse both stdout and stderr from `pulumi preview`, looking for specific diagnostic message patterns.

**Rationale**:

- Diagnostics from analyzers are written to stdout as part of preview output
- The analyzer emits messages like "Estimated Monthly Cost: $X.XX USD"
- Stack summary includes "Total Estimated Monthly Cost: $X.XX USD"
- Can use regex or string matching to verify presence of expected patterns

**Expected Patterns to Match**:

```text
# Per-resource diagnostic
"Estimated Monthly Cost: $"

# Stack summary diagnostic
"Total Estimated Monthly Cost: $"

# Policy pack attribution
"finfocus"

# Graceful degradation message
"Unable to estimate cost" or "No pricing information available"
```

**Alternatives Considered**:

- `--json` output: More structured but may not include analyzer diagnostics in same format
- gRPC interception: Too complex for E2E testing
- Log file parsing: Requires debug logging enabled, less reliable

### 4. Local Backend Strategy

**Context**: E2E tests should run in CI without requiring Pulumi Cloud credentials.

**Decision**: Continue using `pulumi login --local` with file-based backend, consistent with existing E2E tests.

**Rationale**:

- Already established pattern in `test/e2e/main_test.go`
- No external dependencies or credentials required
- State stored in temp directory, cleaned up after test
- Works on all platforms (Linux, macOS, Windows)

**Implementation**:

```go
// Use same pattern as existing E2E tests
stateDir := filepath.Join(workDir, ".pulumi-state")
env := []string{"PULUMI_BACKEND_URL=file://" + stateDir}
runCmdWithEnv(ctx, workDir, env, "pulumi", "login", "--local")
```

### 5. Graceful Skip Strategy

**Context**: Tests should skip gracefully when Pulumi CLI is not installed.

**Decision**: Check for `pulumi` binary in PATH at test start, skip with informative message if not found.

**Rationale**:

- Prevents cryptic failures in environments without Pulumi
- Consistent with existing E2E test skip patterns
- Allows running unit tests without Pulumi installed

**Implementation**:

```go
func skipIfPulumiNotInstalled(t *testing.T) {
    if _, err := exec.LookPath("pulumi"); err != nil {
        t.Skip("Pulumi CLI not installed, skipping analyzer E2E test")
    }
}
```

### 5. Plugin Installation Strategy

**Context**: Need to ensure `aws-public` plugin is installed before running analyzer tests.

**Decision**: Use existing `PluginManager` from `test/e2e/plugin_helpers.go` to install plugin.

**Rationale**:

- Consistent with existing E2E test patterns (see `TestProjectedCost_EC2_WithPlugin`)
- Handles installation, verification, and optional cleanup
- Already tested and working in CI environment

**Implementation**:

```go
pm := NewPluginManager(t)
err := pm.EnsurePluginInstalled(ctx, "aws-public")
require.NoError(t, err, "Failed to install aws-public plugin")
defer pm.DeferPluginCleanup(ctx, "aws-public")()
```

## Open Questions Resolved

- **Q**: How do we configure a local analyzer plugin for development?
  - **A**: Use `plugins.analyzers` section in `Pulumi.yaml` with `name`, `path`, and optional `version` fields.

- **Q**: Do we need `--policy-pack` flag?
  - **A**: No. Analyzers configured in `Pulumi.yaml` are automatically invoked during preview.

- **Q**: Do we need AWS credentials for analyzer E2E tests?
  - **A**: Yes. We use real AWS resources to validate actual pricing accuracy. AWS credentials are already configured in `nightly.yml`.

- **Q**: Where is analyzer output written?
  - **A**: Diagnostics appear in stdout as part of `pulumi preview` output. The server writes logs to stderr only.

- **Q**: How do we validate cost accuracy?
  - **A**: Parse diagnostic output for cost values and compare against expected pricing (~$7.59/month for t3.micro with ±5% tolerance).

## References

- [Pulumi Project File Reference - plugins.analyzers](https://www.pulumi.com/docs/iac/concepts/projects/project-file/)
- [Pulumi Plugin Install CLI](https://www.pulumi.com/docs/reference/cli/pulumi_plugin_install/)
- [Pulumi Environment Variables](https://www.pulumi.com/docs/iac/cli/environment-variables/)
- Existing E2E tests: `test/e2e/main_test.go`, `test/e2e/projects/ec2/Pulumi.yaml`
