package registry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// VersionConstraint represents a semantic version constraint for dependencies.
type VersionConstraint struct {
	Raw        string
	Constraint *semver.Constraints
}

// ParseVersionConstraint parses a version constraint string.
// Supported formats:
//   - ">=1.0.0" - Greater than or equal
//   - "<2.0.0" - Less than
//   - ">=1.0.0,<2.0.0" - Range (AND)
//   - "~1.2.3" - Patch-level changes (>=1.2.3,<1.3.0)
//   - "^1.2.3" - Minor-level changes (>=1.2.3,<2.0.0)
func ParseVersionConstraint(s string) (*VersionConstraint, error) {
	if s == "" {
		return nil, errors.New("empty version constraint")
	}

	constraint, err := semver.NewConstraint(s)
	if err != nil {
		return nil, fmt.Errorf("invalid version constraint %q: %w", s, err)
	}

	return &VersionConstraint{
		Raw:        s,
		Constraint: constraint,
	}, nil
}

// SatisfiesConstraint checks if a version satisfies the constraint.
func SatisfiesConstraint(version string, constraint *VersionConstraint) (bool, error) {
	if constraint == nil || constraint.Constraint == nil {
		return false, fmt.Errorf("nil version constraint")
	}

	// Strip 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	v, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("invalid version %q: %w", version, err)
	}

	return constraint.Constraint.Check(v), nil
}

// CompareVersions compares two versions.
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	ver1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", v1, err)
	}

	ver2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", v2, err)
	}

	return ver1.Compare(ver2), nil
}

// IsValidVersion checks if a string is a valid semantic version.
func IsValidVersion(version string) bool {
	version = strings.TrimPrefix(version, "v")
	_, err := semver.NewVersion(version)
	return err == nil
}
