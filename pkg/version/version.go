package version

// Build information set during compilation.
var (
	version   = "0.1.0"
	gitCommit = "unknown" //nolint:gochecknoglobals // Build system sets this via ldflags
	buildDate = "unknown" //nolint:gochecknoglobals // Build system sets this via ldflags
)

// GetVersion returns the version string.
func GetVersion() string {
	return version
}

// GetGitCommit returns the git commit hash.
func GetGitCommit() string {
	return gitCommit
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return buildDate
}
