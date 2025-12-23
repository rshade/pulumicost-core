package conformance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// CheckProtocolVersion validates that the plugin's protocol version is compatible
// CheckProtocolVersion verifies that pluginVersion is compatible with expectedVersion.
// Versions must be in "major.minor" format. It returns an error if pluginVersion is empty,
// if either version cannot be parsed, or if the major versions do not match or the plugin's
// minor version is less than the expected minor version. On success it returns nil.
func CheckProtocolVersion(pluginVersion, expectedVersion string) error {
	if pluginVersion == "" {
		return errors.New("protocol version is empty")
	}

	pluginMajor, pluginMinor, err := ParseVersion(pluginVersion)
	if err != nil {
		return fmt.Errorf("invalid plugin protocol version %q: %w", pluginVersion, err)
	}

	expectedMajor, expectedMinor, err := ParseVersion(expectedVersion)
	if err != nil {
		return fmt.Errorf("invalid expected protocol version %q: %w", expectedVersion, err)
	}

	if !IsCompatibleVersion(pluginMajor, pluginMinor, expectedMajor, expectedMinor) {
		return fmt.Errorf(
			"protocol version mismatch: got %s, want %s (major versions must match, plugin minor must be >= expected)",
			pluginVersion,
			expectedVersion,
		)
	}

	return nil
}

// versionPartsCount is the expected number of parts in a version string (major.minor).
const versionPartsCount = 2

// ParseVersion parses a version string in "major.minor" format.
// Returns the major and minor version numbers, or an error if the format is invalid.
func ParseVersion(version string) (int, int, error) {
	if version == "" {
		return 0, 0, errors.New("version string is empty")
	}

	parts := strings.Split(version, ".")
	if len(parts) != versionPartsCount {
		return 0, 0, fmt.Errorf("version format must be 'major.minor', got %q", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse minor version: %w", err)
	}

	if major < 0 {
		return 0, 0, fmt.Errorf("major version cannot be negative: %d", major)
	}

	if minor < 0 {
		return 0, 0, fmt.Errorf("minor version cannot be negative: %d", minor)
	}

	return major, minor, nil
}

// IsCompatibleVersion checks if a plugin version is compatible with the expected version.
// Compatibility rules:
//   - Major versions must match exactly.
//   - Plugin minor version must be >= expected minor version (plugins can be ahead).
func IsCompatibleVersion(pluginMajor, pluginMinor, expectedMajor, expectedMinor int) bool {
	// Major versions must match
	if pluginMajor != expectedMajor {
		return false
	}

	// Plugin minor version must be at least the expected minor version
	return pluginMinor >= expectedMinor
}