package pluginhost

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// CompatibilityResult represents the outcome of a version compatibility check.
type CompatibilityResult int

const (
	// Compatible indicates the versions are compatible.
	Compatible CompatibilityResult = iota
	// MinorMismatch indicates a minor version difference (usually compatible).
	MinorMismatch
	// MajorMismatch indicates a major version difference (likely incompatible).
	MajorMismatch
	// Invalid indicates one or both version strings are invalid.
	Invalid
)

// CompareSpecVersions compares the core spec version with the plugin's spec version.
// It returns a compatibility result and an error if parsing fails.
func CompareSpecVersions(coreVersion, pluginVersion string) (CompatibilityResult, error) {
	cVer, err := semver.NewVersion(coreVersion)
	if err != nil {
		return Invalid, fmt.Errorf("invalid core version: %w", err)
	}

	pVer, err := semver.NewVersion(pluginVersion)
	if err != nil {
		return Invalid, fmt.Errorf("invalid plugin version: %w", err)
	}

	if cVer.Major() != pVer.Major() {
		return MajorMismatch, nil
	}

	// Minor version differences are generally compatible in semver,
	// but we track them just in case specific logic is needed later.
	// For now, if major matches, we consider it compatible.
	if cVer.Minor() != pVer.Minor() {
		return Compatible, nil
	}

	return Compatible, nil
}
