package registry

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// RegistryEntry represents a plugin in the embedded registry.
//
//nolint:revive // type stuttering is acceptable for clarity
type RegistryEntry struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Repository         string   `json:"repository"`
	Author             string   `json:"author"`
	License            string   `json:"license"`
	Homepage           string   `json:"homepage"`
	SupportedProviders []string `json:"supported_providers"`
	Capabilities       []string `json:"capabilities"`
	SecurityLevel      string   `json:"security_level"`
	MinSpecVersion     string   `json:"min_spec_version"`
}

// PluginSpecifier represents a parsed plugin specifier (name or URL with optional version).
type PluginSpecifier struct {
	Name    string
	Version string
	IsURL   bool
	Owner   string
	Repo    string
}

var (
	repoPattern   = regexp.MustCompile(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`)
	githubPattern = regexp.MustCompile(`^github\.com/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)$`)
)

// ValidateRegistryEntry validates all required fields of a registry entry.
func ValidateRegistryEntry(entry RegistryEntry) error {
	if entry.Name == "" {
		return errors.New("registry entry missing required field: name")
	}
	if entry.Repository == "" {
		return fmt.Errorf("registry entry %q missing required field: repository", entry.Name)
	}
	if !repoPattern.MatchString(entry.Repository) {
		return fmt.Errorf(
			"registry entry %q has invalid repository format: %s (expected owner/repo)",
			entry.Name,
			entry.Repository,
		)
	}
	validLevels := map[string]bool{"official": true, "community": true, "experimental": true}
	if entry.SecurityLevel != "" && !validLevels[entry.SecurityLevel] {
		return fmt.Errorf("registry entry %q has invalid security_level: %s", entry.Name, entry.SecurityLevel)
	}
	return nil
}

// ParsePluginSpecifier parses a plugin specifier string.
// Formats:
//   - "kubecost" - registry plugin, latest version
//   - "kubecost@v1.0.0" - registry plugin, specific version
//   - "github.com/owner/repo" - GitHub URL, latest version
//   - "github.com/owner/repo@v1.0.0" - GitHub URL, specific version
func ParsePluginSpecifier(spec string) (*PluginSpecifier, error) {
	if spec == "" {
		return nil, errors.New("empty plugin specifier")
	}

	// Split by @ to extract version
	parts := strings.SplitN(spec, "@", 2) //nolint:mnd // split into 2 parts
	nameOrURL := parts[0]
	version := ""
	if len(parts) == 2 { //nolint:mnd // check for version part
		version = parts[1]
	}

	// Check if it's a GitHub URL
	if strings.HasPrefix(nameOrURL, "github.com/") {
		matches := githubPattern.FindStringSubmatch(nameOrURL)
		if matches == nil {
			return nil, fmt.Errorf("invalid GitHub URL format: %s", nameOrURL)
		}
		// Derive plugin name from repo name
		repoName := matches[2]
		pluginName := strings.TrimPrefix(repoName, "pulumicost-plugin-")

		return &PluginSpecifier{
			Name:    pluginName,
			Version: version,
			IsURL:   true,
			Owner:   matches[1],
			Repo:    matches[2],
		}, nil
	}

	// It's a registry name
	return &PluginSpecifier{
		Name:    nameOrURL,
		Version: version,
		IsURL:   false,
	}, nil
}

// ParseGitHubURL extracts owner and repo from a GitHub URL.
//
//nolint:nonamedreturns // clear return value names for documentation
func ParseGitHubURL(url string) (owner, repo string, err error) {
	matches := githubPattern.FindStringSubmatch(url)
	if matches == nil {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
	}
	return matches[1], matches[2], nil
}
