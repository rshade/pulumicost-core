//go:build ignore

// Package registry provides the internal API contracts for plugin management.
// This file documents the public API surface for the registry package.
// NOTE: This file is not compiled - it serves as design documentation only.
package registry

// Installer provides plugin installation capabilities.
type Installer interface {
	// Install downloads and installs a plugin from registry or URL.
	// Returns the installed plugin path and version.
	Install(specifier string, opts InstallOptions) (*InstalledPlugin, error)

	// Update updates an installed plugin to the latest or specified version.
	Update(name string, opts UpdateOptions) (*InstalledPlugin, error)

	// Remove uninstalls a plugin and optionally removes config entry.
	Remove(name string, opts RemoveOptions) error

	// List returns all installed plugins.
	List() ([]*InstalledPlugin, error)
}

// InstallOptions configures plugin installation behavior.
type InstallOptions struct {
	Force     bool   // Reinstall even if version exists
	NoSave    bool   // Don't add to config file
	PluginDir string // Custom plugin directory (default: ~/.finfocus/plugins)
}

// UpdateOptions configures plugin update behavior.
type UpdateOptions struct {
	AllowDowngrade bool   // Allow installing older version
	DryRun         bool   // Show what would be updated without changes
	Version        string // Specific version to update to (empty = latest)
}

// RemoveOptions configures plugin removal behavior.
type RemoveOptions struct {
	AllVersions bool // Remove all versions of plugin
	KeepConfig  bool // Don't remove from config file
	Force       bool // Remove even if dependencies exist
}

// GitHubClient provides GitHub API access for releases.
type GitHubClient interface {
	// GetLatestRelease returns the latest release for a repository.
	GetLatestRelease(owner, repo string) (*GitHubRelease, error)

	// GetReleaseByTag returns a specific release by tag name.
	GetReleaseByTag(owner, repo, tag string) (*GitHubRelease, error)

	// DownloadAsset downloads a release asset to a local file.
	DownloadAsset(url, destPath string, progress func(downloaded, total int64)) error
}

// RegistryLookup provides registry query capabilities.
type RegistryLookup interface {
	// Get returns a registry entry by plugin name.
	Get(name string) (*RegistryEntry, error)

	// List returns all registered plugins.
	List() []RegistryEntry

	// Validate checks if the registry is valid.
	Validate() error
}

// DependencyResolver resolves plugin dependencies.
type DependencyResolver interface {
	// Resolve returns plugins in installation order, detecting cycles.
	Resolve(manifest *PluginManifest) ([]string, error)

	// CheckConflicts returns any version constraint conflicts.
	CheckConflicts(name string, installedPlugins []*InstalledPlugin) error
}

// ConfigManager manages plugin configuration persistence.
type ConfigManager interface {
	// GetPlugins returns all configured plugins.
	GetPlugins() []PluginConfig

	// AddPlugin adds a plugin to the config.
	AddPlugin(config PluginConfig) error

	// RemovePlugin removes a plugin from the config.
	RemovePlugin(name string) error

	// UpdatePlugin updates a plugin version in the config.
	UpdatePlugin(name, version string) error

	// Save persists the config to disk.
	Save() error
}

// ArchiveExtractor extracts plugin archives.
type ArchiveExtractor interface {
	// Extract extracts an archive to the destination directory.
	// Supports tar.gz and zip formats based on file extension.
	Extract(archivePath, destDir string) error
}
