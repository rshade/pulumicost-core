package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rshade/pulumicost-core/internal/config"
)

const (
	executablePermissionMask = 0111 // File permission mask for executable files
)

// InstallOptions configures plugin installation behavior.
type InstallOptions struct {
	Force     bool   // Reinstall even if version exists
	NoSave    bool   // Don't add to config file
	PluginDir string // Custom plugin directory (default: ~/.pulumicost/plugins)
}

// InstallResult contains the result of a plugin installation.
type InstallResult struct {
	Name       string
	Version    string
	Path       string
	FromURL    bool
	Repository string
}

// Installer handles plugin installation from registry or URLs.
type Installer struct {
	client    *GitHubClient
	pluginDir string
}

// convertToAssetNamingHints converts registry asset hints to the GitHub client format.
func convertToAssetNamingHints(hints *RegistryAssetHints) *AssetNamingHints {
	if hints == nil {
		return nil
	}
	return &AssetNamingHints{
		AssetPrefix:   hints.AssetPrefix,
		Region:        hints.DefaultRegion,
		VersionPrefix: hints.VersionPrefix,
	}
}

// NewInstaller creates a new Installer configured to install plugins into pluginDir.
// If pluginDir is empty, it defaults to "$HOME/.pulumicost/plugins"; if the home
// directory cannot be determined, the default is "./.pulumicost/plugins" relative to
// NewInstaller creates an Installer configured to install plugins into pluginDir.
// If pluginDir is empty, it defaults to $HOME/.pulumicost/plugins; when the user
// home directory cannot be determined it falls back to the current directory.
// The returned Installer contains an initialized GitHub client.
func NewInstaller(pluginDir string) *Installer {
	if pluginDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home cannot be determined
			homeDir = "."
		}
		pluginDir = filepath.Join(homeDir, ".pulumicost", "plugins")
	}
	return &Installer{
		client:    NewGitHubClient(),
		pluginDir: pluginDir,
	}
}

// NewInstallerWithClient creates a new Installer using the provided GitHub client.
// If pluginDir is empty, it defaults to $HOME/.pulumicost/plugins; if the home
// directory cannot be determined it falls back to the current directory.
// NewInstallerWithClient creates an Installer that uses the provided GitHub client and a resolved plugin directory.
// If pluginDir is empty, it defaults to "$HOME/.pulumicost/plugins"; if the user home directory cannot be determined
// it falls back to the current directory ("./") and uses "./.pulumicost/plugins".
// NewInstallerWithClient creates an Installer that uses the provided GitHub client and plugin directory.
// If pluginDir is empty, it is resolved to $HOME/.pulumicost/plugins; if the user's home directory
// cannot be determined, it falls back to the current working directory.
// The returned Installer's client field is set to the provided client and its pluginDir field to the
// resolved path.
func NewInstallerWithClient(client *GitHubClient, pluginDir string) *Installer {
	if pluginDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home cannot be determined
			homeDir = "."
		}
		pluginDir = filepath.Join(homeDir, ".pulumicost", "plugins")
	}
	return &Installer{
		client:    client,
		pluginDir: pluginDir,
	}
}

// acquireLock attempts to acquire a file-based lock for the specified plugin name.
// It returns a function to release the lock, or an error if the lock cannot be acquired.
func (i *Installer) acquireLock(name string) (func(), error) {
	// Ensure plugin base directory exists
	if err := os.MkdirAll(i.pluginDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	lockPath := filepath.Join(i.pluginDir, name+".lock")

	// Try to create the lock file exclusively
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf(
				"plugin %q is currently being modified by another process (lock file %s exists)",
				name, lockPath)
		}
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}
	_ = file.Close()

	return func() {
		_ = os.Remove(lockPath)
	}, nil
}

// Install installs a plugin from a specifier (name or URL with optional version).
func (i *Installer) Install(
	specifier string,
	opts InstallOptions,
	progress func(msg string),
) (*InstallResult, error) {
	spec, err := ParsePluginSpecifier(specifier)
	if err != nil {
		return nil, err
	}

	// Acquire lock for this plugin
	unlock, err := i.acquireLock(spec.Name)
	if err != nil {
		return nil, err
	}
	defer unlock()

	if spec.IsURL {
		return i.installFromURL(spec, opts, progress)
	}
	return i.installFromRegistry(spec, opts, progress)
}

// installFromRegistry installs a plugin from the embedded registry.
func (i *Installer) installFromRegistry(
	spec *PluginSpecifier,
	opts InstallOptions,
	progress func(msg string),
) (*InstallResult, error) {
	// Look up plugin in registry
	entry, err := GetPlugin(spec.Name)
	if err != nil {
		// Check if this is a "not found" error vs other registry errors
		if strings.Contains(err.Error(), "not found in registry") {
			return nil, fmt.Errorf("plugin %q not found in registry", spec.Name)
		}
		return nil, fmt.Errorf("failed to access registry: %w", err)
	}

	// Validate entry
	if validateErr := ValidateRegistryEntry(*entry); validateErr != nil {
		return nil, validateErr
	}

	// Parse repository
	owner, repo, err := parseOwnerRepo(entry.Repository)
	if err != nil {
		return nil, err
	}

	// Get release
	var release *GitHubRelease
	if spec.Version != "" {
		if progress != nil {
			progress(fmt.Sprintf("Fetching release %s for %s...", spec.Version, spec.Name))
		}
		release, err = i.client.GetReleaseByTag(owner, repo, spec.Version)
	} else {
		if progress != nil {
			progress(fmt.Sprintf("Fetching latest release for %s...", spec.Name))
		}
		release, err = i.client.GetLatestRelease(owner, repo)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	// Convert registry hints to asset hints
	assetHints := convertToAssetNamingHints(entry.AssetHints)

	// Install the release
	result, err := i.installRelease(
		spec.Name,
		release,
		entry.Repository,
		opts,
		progress,
		assetHints,
	)
	if err != nil {
		return nil, err
	}

	result.Repository = entry.Repository
	return result, nil
}

// installFromURL installs a plugin directly from a GitHub URL.
func (i *Installer) installFromURL(
	spec *PluginSpecifier,
	opts InstallOptions,
	progress func(msg string),
) (*InstallResult, error) {
	// Get release
	var release *GitHubRelease
	var err error
	if spec.Version != "" {
		if progress != nil {
			progress(
				fmt.Sprintf(
					"Fetching release %s from %s/%s...",
					spec.Version,
					spec.Owner,
					spec.Repo,
				),
			)
		}
		release, err = i.client.GetReleaseByTag(spec.Owner, spec.Repo, spec.Version)
	} else {
		if progress != nil {
			progress(fmt.Sprintf("Fetching latest release from %s/%s...", spec.Owner, spec.Repo))
		}
		release, err = i.client.GetLatestRelease(spec.Owner, spec.Repo)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	// Install the release (no hints for URL-based installs)
	repository := fmt.Sprintf("%s/%s", spec.Owner, spec.Repo)
	result, err := i.installRelease(spec.Name, release, repository, opts, progress, nil)
	if err != nil {
		return nil, err
	}

	result.FromURL = true
	result.Repository = fmt.Sprintf("%s/%s", spec.Owner, spec.Repo)
	return result, nil
}

// installRelease downloads and installs a specific release.
//
//nolint:gocognit,funlen // Complex but necessary installation logic with many steps
func (i *Installer) installRelease(
	name string,
	release *GitHubRelease,
	repository string,
	opts InstallOptions,
	progress func(msg string),
	hints *AssetNamingHints,
) (*InstallResult, error) {
	version := release.TagName

	// Determine plugin directory
	pluginDir := i.pluginDir
	if opts.PluginDir != "" {
		pluginDir = opts.PluginDir
	}

	// Check if already installed
	installDir := filepath.Join(pluginDir, name, version)
	if _, err := os.Stat(installDir); err == nil && !opts.Force {
		return nil, fmt.Errorf(
			"plugin %s@%s already installed. Use --force to reinstall",
			name,
			version,
		)
	}

	// Find platform-specific asset (use hints if provided)
	asset, err := FindPlatformAssetWithHints(release, name, hints)
	if err != nil {
		return nil, err
	}

	if progress != nil {
		progress(fmt.Sprintf("Downloading %s (%d bytes)...", asset.Name, asset.Size))
	}

	// Determine extension for temp file
	pattern := "pulumicost-plugin-*"
	switch {
	case strings.HasSuffix(asset.Name, extZip):
		pattern += extZip
	case strings.HasSuffix(asset.Name, extTarGz):
		pattern += extTarGz
	case strings.HasSuffix(asset.Name, ".tgz"):
		pattern += ".tgz"
	default:
		pattern += filepath.Ext(asset.Name)
	}

	// Create temp file for download
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	// Download asset
	downloadProgress := func(downloaded, total int64) {
		if progress != nil && total > 0 {
			pct := float64(downloaded) / float64(total) * 100 //nolint:mnd // percentage calculation
			progress(fmt.Sprintf("Downloading... %.0f%%", pct))
		}
	}
	if downloadErr := i.client.DownloadAsset(asset.BrowserDownloadURL, tmpPath, downloadProgress); downloadErr != nil {
		return nil, fmt.Errorf("failed to download: %w", downloadErr)
	}

	// Create install directory
	if mkdirErr := os.MkdirAll(installDir, 0750); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create directory: %w", mkdirErr)
	}

	if progress != nil {
		progress("Extracting archive...")
	}

	// Extract archive
	if extractErr := ExtractArchive(tmpPath, installDir); extractErr != nil {
		_ = os.RemoveAll(installDir)
		return nil, fmt.Errorf("failed to extract: %w", extractErr)
	}

	// Find and validate binary
	binaryPath := findPluginBinary(installDir, name)
	if binaryPath == "" {
		_ = os.RemoveAll(installDir)
		return nil, errors.New("no executable found in archive")
	}

	if validateErr := ValidateBinary(binaryPath); validateErr != nil {
		_ = os.RemoveAll(installDir)
		return nil, validateErr
	}

	// Save to config unless --no-save
	if !opts.NoSave {
		plugin := config.InstalledPlugin{
			Name:    name,
			URL:     fmt.Sprintf("github.com/%s", repository),
			Version: version,
		}
		if addErr := config.AddInstalledPlugin(plugin); addErr != nil {
			// Non-fatal, just warn
			if progress != nil {
				progress(fmt.Sprintf("Warning: failed to save to config: %v", addErr))
			}
		}
	}

	if progress != nil {
		progress(fmt.Sprintf("Successfully installed %s@%s", name, version))
	}

	return &InstallResult{
		Name:    name,
		Version: version,
		Path:    installDir,
	}, nil
}

// parseOwnerRepo parses a repository string in the "owner/repo" format and returns
// the owner and repository name. It returns an error if the input does not contain
// exactly one '/' separator or if either the owner or repo component is empty.
func parseOwnerRepo(repo string) (string, string, error) {
	parts := strings.SplitN(repo, "/", 2) //nolint:mnd // split into 2 parts
	if len(parts) != 2 {                  //nolint:mnd // expect 2 parts
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]
	if owner == "" || repoName == "" {
		return "", "", errors.New("invalid repository format: owner or repo empty")
	}
	return owner, repoName, nil
}

// findPluginBinary searches dir for an executable plugin binary matching name.
// It first checks common filename patterns (e.g. name, name.exe,
// pulumicost-plugin-name, pulumicost-plugin-name.exe). If none match, it
// scans the directory for any executable file (on Windows files must end with
// .exe; on other systems the file must have an executable bit). It returns the
// full path to the first matching file, or an empty string if no binary is
// found.
func findPluginBinary(dir, name string) string {
	// Check common patterns
	patterns := []string{
		name,
		name + ".exe",
		"pulumicost-plugin-" + name,
		"pulumicost-plugin-" + name + ".exe",
	}

	for _, pattern := range patterns {
		path := filepath.Join(dir, pattern)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	// Fall back to scanning directory
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
			// Check executability based on OS
			var isExecutable bool
			if runtime.GOOS == osWindows {
				isExecutable = strings.HasSuffix(path, ".exe")
			} else {
				isExecutable = info.Mode()&executablePermissionMask != 0
			}
			if isExecutable {
				return path
			}
		}
	}

	return ""
}

// UpdateOptions configures plugin update behavior.
type UpdateOptions struct {
	DryRun    bool   // Show what would be updated without changes
	Version   string // Specific version to update to (empty = latest)
	PluginDir string // Custom plugin directory
}

// UpdateResult contains the result of a plugin update.
type UpdateResult struct {
	Name        string
	OldVersion  string
	NewVersion  string
	Path        string
	WasUpToDate bool
}

// Update updates an installed plugin to the latest or specified version.
//
//nolint:gocognit,funlen // Complex but necessary update logic with version comparison
func (i *Installer) Update(
	name string,
	opts UpdateOptions,
	progress func(msg string),
) (*UpdateResult, error) {
	// Acquire lock for this plugin
	unlock, err := i.acquireLock(name)
	if err != nil {
		return nil, err
	}
	defer unlock()

	// Get installed plugin info
	installed, err := config.GetInstalledPlugin(name)
	if err != nil {
		return nil, fmt.Errorf("plugin %q is not installed", name)
	}

	// Look up in registry first, then try as URL
	owner, repo, assetHints, err := i.resolvePluginSource(name, installed.URL)
	if err != nil {
		return nil, err
	}

	// Get target version
	var release *GitHubRelease
	if opts.Version != "" {
		if progress != nil {
			progress(fmt.Sprintf("Fetching release %s for %s...", opts.Version, name))
		}
		release, err = i.client.GetReleaseByTag(owner, repo, opts.Version)
	} else {
		if progress != nil {
			progress(fmt.Sprintf("Checking for updates to %s...", name))
		}
		release, err = i.client.GetLatestRelease(owner, repo)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	newVersion := release.TagName
	oldVersion := installed.Version

	// Compare versions
	cmp, err := CompareVersions(newVersion, oldVersion)
	if err != nil {
		// If version comparison fails, proceed anyway
		cmp = 1
	}

	if cmp == 0 {
		return &UpdateResult{
			Name:        name,
			OldVersion:  oldVersion,
			NewVersion:  newVersion,
			WasUpToDate: true,
		}, nil
	}

	if cmp < 0 && opts.Version == "" {
		// Latest is older than installed (shouldn't happen normally)
		return &UpdateResult{
			Name:        name,
			OldVersion:  oldVersion,
			NewVersion:  newVersion,
			WasUpToDate: true,
		}, nil
	}

	if opts.DryRun {
		if progress != nil {
			progress(fmt.Sprintf("Would update %s from %s to %s", name, oldVersion, newVersion))
		}
		return &UpdateResult{
			Name:       name,
			OldVersion: oldVersion,
			NewVersion: newVersion,
		}, nil
	}

	// Install new version
	pluginDir := i.pluginDir
	if opts.PluginDir != "" {
		pluginDir = opts.PluginDir
	}

	installOpts := InstallOptions{
		Force:     true, // Allow overwriting
		NoSave:    true, // We'll update config ourselves
		PluginDir: pluginDir,
	}

	repository := fmt.Sprintf("%s/%s", owner, repo)
	result, err := i.installRelease(name, release, repository, installOpts, progress, assetHints)
	if err != nil {
		return nil, err
	}

	// Remove old version directory
	oldDir := filepath.Join(pluginDir, name, oldVersion)
	if oldVersion != newVersion {
		_ = os.RemoveAll(oldDir)
	}

	// Update config
	if updateErr := config.UpdateInstalledPluginVersion(name, newVersion); updateErr != nil {
		if progress != nil {
			progress(fmt.Sprintf("Warning: failed to update config: %v", updateErr))
		}
	}

	return &UpdateResult{
		Name:       name,
		OldVersion: oldVersion,
		NewVersion: newVersion,
		Path:       result.Path,
	}, nil
}

// resolvePluginSource resolves the GitHub owner, repo, and asset hints for a plugin.
// It first looks up the plugin in the embedded registry. If not found, it parses
// the installed URL to extract owner and repo. Returns an error if neither lookup
// succeeds.
func (i *Installer) resolvePluginSource(
	name, installedURL string,
) (string, string, *AssetNamingHints, error) {
	// Try registry first
	entry, err := GetPlugin(name)
	if err == nil {
		// Found in registry - parse owner/repo and get hints
		owner, repo, parseErr := parseOwnerRepo(entry.Repository)
		if parseErr != nil {
			return "", "", nil, parseErr
		}
		hints := convertToAssetNamingHints(entry.AssetHints)
		return owner, repo, hints, nil
	}

	// Not in registry - try parsing installed URL
	if installedURL == "" {
		return "", "", nil, fmt.Errorf("plugin %q not found in registry and no URL available", name)
	}

	// Remove github.com/ prefix if present
	urlPath := strings.TrimPrefix(installedURL, "github.com/")
	owner, repo, parseErr := parseOwnerRepo(urlPath)
	if parseErr != nil {
		return "", "", nil, fmt.Errorf("failed to parse plugin URL %q: %w", installedURL, parseErr)
	}

	// URL-based plugins have no hints
	return owner, repo, nil, nil
}

// RemoveOptions configures plugin removal behavior.
type RemoveOptions struct {
	KeepConfig bool   // Don't remove from config file
	PluginDir  string // Custom plugin directory
}

// Remove removes an installed plugin.
func (i *Installer) Remove(name string, opts RemoveOptions, progress func(msg string)) error {
	// Acquire lock for this plugin
	unlock, err := i.acquireLock(name)
	if err != nil {
		return err
	}
	defer unlock()

	// Get installed plugin info
	installed, err := config.GetInstalledPlugin(name)
	if err != nil {
		return fmt.Errorf("plugin %q is not installed", name)
	}

	pluginDir := i.pluginDir
	if opts.PluginDir != "" {
		pluginDir = opts.PluginDir
	}

	// Remove plugin directory
	pluginPath := filepath.Join(pluginDir, name, installed.Version)
	if progress != nil {
		progress(fmt.Sprintf("Removing %s@%s...", name, installed.Version))
	}

	if removeErr := os.RemoveAll(pluginPath); removeErr != nil {
		return fmt.Errorf("failed to remove plugin directory: %w", removeErr)
	}

	// Remove parent directory if empty
	parentDir := filepath.Join(pluginDir, name)
	entries, err := os.ReadDir(parentDir)
	if err == nil && len(entries) == 0 {
		if rmErr := os.Remove(parentDir); rmErr != nil && progress != nil {
			progress(
				fmt.Sprintf("Warning: failed to remove parent directory %s: %v", parentDir, rmErr),
			)
		}
	}

	// Remove from config unless --keep-config
	if !opts.KeepConfig {
		if removeErr := config.RemoveInstalledPlugin(name); removeErr != nil {
			if progress != nil {
				progress(fmt.Sprintf("Warning: failed to remove from config: %v", removeErr))
			}
		}
	}

	if progress != nil {
		progress(fmt.Sprintf("Successfully removed %s", name))
	}

	return nil
}