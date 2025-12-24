# Research: Plugin Install/Update/Remove System

**Date**: 2025-11-23
**Feature**: 006-plugin-install

## Research Items

### 1. GitHub Releases API Integration

**Decision**: Use Go's net/http with GitHub REST API v3

**Rationale**:
- Native Go HTTP client is sufficient for simple API calls
- No need for third-party GitHub client library for basic release queries
- Easier to control retry behavior and error handling

**Alternatives Considered**:
- `google/go-github`: Too heavy for just release API; adds unnecessary dependency
- `shurcooL/githubv4`: GraphQL is overkill for simple release lookups

**Implementation Notes**:
- Endpoints: `/repos/{owner}/{repo}/releases/latest`, `/repos/{owner}/{repo}/releases/tags/{tag}`
- Use `GITHUB_TOKEN` env var for authentication
- Fallback to `gh auth token` command output
- Handle 403 rate limit with clear error message suggesting token usage

### 2. Archive Extraction (tar.gz and zip)

**Decision**: Use Go standard library archive/tar, archive/zip, compress/gzip

**Rationale**:
- Standard library handles both formats well
- No external dependencies needed
- Secure against zip-slip attacks with proper path validation

**Alternatives Considered**:
- `mholt/archiver`: More features but adds dependency
- Shell commands (tar, unzip): Not cross-platform

**Implementation Notes**:
- Detect format by file extension (.tar.gz, .zip)
- Validate extracted paths don't escape target directory (zip-slip prevention)
- Set executable bit on binaries after extraction (Unix only)

### 3. Semantic Version Constraint Parsing

**Decision**: Use `Masterminds/semver/v3` library

**Rationale**:
- Well-maintained, widely used in Go ecosystem (Helm, Hugo use it)
- Supports all required operators: >=, <, ~, ^
- Handles version comparison and constraint checking

**Alternatives Considered**:
- Custom parser: More work, error-prone
- `hashicorp/go-version`: Less feature-rich for constraints

**Implementation Notes**:
- Parse constraints from plugin.manifest.json dependencies
- Check if installed version satisfies constraint
- Support comma-separated ranges (>=1.0.0,<2.0.0)

### 4. Configuration File Format

**Decision**: Extend existing config.yaml with plugins section

**Rationale**:
- Consistent with existing config structure
- Uses existing gopkg.in/yaml.v3 dependency
- Users familiar with YAML format

**Format**:
```yaml
plugins:
  - name: kubecost
    url: github.com/rshade/pulumicost-plugin-kubecost
    version: v0.0.1
```

**Implementation Notes**:
- Add `Plugins []PluginConfig` to existing Config struct
- Atomic write with temp file + rename for safety
- Preserve existing config values when updating plugins

### 5. Registry JSON Schema

**Decision**: Embed registry.json using Go's //go:embed directive

**Rationale**:
- Single binary distribution (no external files needed)
- Compile-time validation of JSON structure
- Updates with each release

**Schema Alignment**: Matches registry.proto PluginInfo message fields

**Implementation Notes**:
- `//go:embed registry/registry.json`
- Parse on first access, cache in memory
- Validate required fields on load

### 6. Exponential Backoff for Retries

**Decision**: Custom implementation with configurable parameters

**Rationale**:
- Simple logic: delay = base * 2^attempt
- Clarified: 3 retries, 1 second base, max ~7 seconds total
- No need for external library

**Implementation Notes**:
- First retry: 1s, second: 2s, third: 4s
- Total max wait: 7 seconds
- Retry on network errors, timeout, 5xx responses
- Don't retry on 4xx client errors (except 429 rate limit)

### 7. Plugin Binary Naming Convention

**Decision**: Follow GoReleaser v2 archive naming

**Format**: `{projectname}_{version}_{os}_{arch}.{format}`

**Examples**:
- `pulumicost-plugin-kubecost_v1.0.0_linux_amd64.tar.gz`
- `pulumicost-plugin-kubecost_v1.0.0_windows_amd64.zip`

**OS/Arch Mapping**:
- runtime.GOOS: linux, darwin, windows
- runtime.GOARCH: amd64, arm64

### 8. Dependency Resolution Algorithm

**Decision**: Topological sort with cycle detection

**Rationale**:
- Standard algorithm for dependency graphs
- Detects circular dependencies immediately
- Installs dependencies in correct order

**Implementation Notes**:
- Build dependency graph from manifests
- Kahn's algorithm for topological sort
- Error on cycle with list of involved plugins
- Recursive installation: resolve dependencies before installing dependee

## Dependencies to Add

```go
// go.mod additions
require (
    github.com/Masterminds/semver/v3 v3.2.1
)
```

No other new dependencies needed - use standard library for HTTP, archive, JSON.

## Security Considerations

1. **Zip-slip Prevention**: Validate all extracted paths are within target directory
2. **GitHub Token Handling**: Read from env var, never log or store
3. **Binary Validation**: Check executable bit/extension after extraction
4. **Network Security**: HTTPS only for GitHub API and downloads
5. **File Permissions**: Plugin directory 0755, binaries 0755, config 0644

## Performance Considerations

1. **Parallel Downloads**: Future optimization - download multiple plugins concurrently
2. **Progress Reporting**: Stream download with progress bar for large plugins
3. **Caching**: Registry JSON cached in memory after first parse
4. **Lazy Loading**: Don't load all plugin manifests on startup, only when needed
