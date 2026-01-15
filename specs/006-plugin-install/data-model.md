# Data Model: Plugin Install/Update/Remove System

**Date**: 2025-11-23
**Feature**: 006-plugin-install

## Entities

### RegistryEntry

Represents a plugin in the embedded registry.json.

```go
type RegistryEntry struct {
    Name               string   `json:"name"`
    Description        string   `json:"description"`
    Repository         string   `json:"repository"`          // "owner/repo" format
    Author             string   `json:"author"`
    License            string   `json:"license"`
    Homepage           string   `json:"homepage"`
    SupportedProviders []string `json:"supported_providers"`
    Capabilities       []string `json:"capabilities"`
    SecurityLevel      string   `json:"security_level"`      // "official", "community", "experimental"
    MinSpecVersion     string   `json:"min_spec_version"`
}
```

**Validation Rules**:
- Name: required, alphanumeric with hyphens
- Repository: required, matches pattern `^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`
- SecurityLevel: enum {"official", "community", "experimental"}

### Registry

Container for embedded registry data.

```go
type Registry struct {
    SchemaVersion string                   `json:"schema_version"`
    Plugins       map[string]RegistryEntry `json:"plugins"`
}
```

### PluginConfig

Represents a plugin entry in config.yaml.

```go
type PluginConfig struct {
    Name    string `yaml:"name"`
    URL     string `yaml:"url"`      // "github.com/owner/repo" format
    Version string `yaml:"version"`  // "v1.0.0" format
}
```

**Validation Rules**:
- Name: required, matches registry entry or derived from URL
- URL: required, valid GitHub URL format
- Version: required, valid semver with v prefix

### PluginManifest

Represents plugin metadata from plugin.manifest.json.

```go
type PluginManifest struct {
    Name         string              `json:"name"`
    Version      string              `json:"version"`
    Description  string              `json:"description"`
    Author       string              `json:"author"`
    License      string              `json:"license"`
    Requirements PluginRequirements  `json:"requirements"`
}

type PluginRequirements struct {
    MinSpecVersion string             `json:"min_spec_version"`
    Dependencies   []PluginDependency `json:"dependencies"`
}

type PluginDependency struct {
    Name              string `json:"name"`
    VersionConstraint string `json:"version_constraint"`
    Optional          bool   `json:"optional"`
}
```

### VersionConstraint

Represents a semantic version constraint for dependencies.

```go
type VersionConstraint struct {
    Raw        string              // Original string: ">=1.0.0,<2.0.0"
    Constraint *semver.Constraints // Parsed constraint object
}
```

**Supported Operators**:
- `>=1.0.0` - Greater than or equal
- `<2.0.0` - Less than
- `~1.2.3` - Patch-level changes (>=1.2.3,<1.3.0)
- `^1.2.3` - Minor-level changes (>=1.2.3,<2.0.0)

### GitHubRelease

Represents release metadata from GitHub API.

```go
type GitHubRelease struct {
    TagName    string         `json:"tag_name"`
    Name       string         `json:"name"`
    Draft      bool           `json:"draft"`
    Prerelease bool           `json:"prerelease"`
    Assets     []ReleaseAsset `json:"assets"`
}

type ReleaseAsset struct {
    Name               string `json:"name"`
    Size               int64  `json:"size"`
    BrowserDownloadURL string `json:"browser_download_url"`
    ContentType        string `json:"content_type"`
}
```

### InstalledPlugin

Runtime representation of an installed plugin.

```go
type InstalledPlugin struct {
    Name       string
    Version    string
    Path       string          // Full path to plugin directory
    BinaryPath string          // Full path to executable
    Manifest   *PluginManifest // Loaded manifest, nil if missing
}
```

## Relationships

```text
Registry 1---* RegistryEntry     (registry contains many plugin entries)
Config 1---* PluginConfig        (config contains many plugin configs)
PluginManifest 1---* PluginDependency (manifest has many dependencies)
InstalledPlugin 1---1 PluginManifest  (installed plugin has one manifest)
```

## State Transitions

### Plugin Installation States

```text
[Not Installed] --install--> [Downloading] --extract--> [Installed]
[Installed] --update--> [Downloading] --extract--> [Updated]
[Installed] --remove--> [Not Installed]
```

### Config States

```text
[Plugin Not in Config] --install--> [Plugin in Config]
[Plugin in Config] --remove (no --keep-config)--> [Plugin Not in Config]
[Plugin in Config] --update--> [Plugin in Config (version updated)]
```

## File System Layout

```text
~/.finfocus/
├── config.yaml
│   └── plugins:
│       - {name, url, version}
└── plugins/
    └── {plugin-name}/
        └── {version}/
            ├── finfocus-plugin-{name}    # Binary (or .exe on Windows)
            └── plugin.manifest.json         # Optional manifest
```

## Validation Functions

```go
// ValidateRegistryEntry validates all required fields
func ValidateRegistryEntry(entry RegistryEntry) error

// ValidatePluginConfig validates config entry
func ValidatePluginConfig(config PluginConfig) error

// ParseVersionConstraint parses semver constraint string
func ParseVersionConstraint(s string) (*VersionConstraint, error)

// SatisfiesConstraint checks if version satisfies constraint
func SatisfiesConstraint(version string, constraint *VersionConstraint) bool

// ParseGitHubURL extracts owner/repo from URL
func ParseGitHubURL(url string) (owner, repo string, err error)

// ParsePluginSpecifier parses "name@version" or "github.com/owner/repo@version"
func ParsePluginSpecifier(spec string) (name, version string, isURL bool, err error)
```
