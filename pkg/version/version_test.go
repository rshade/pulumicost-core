package version

import "testing"

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion() returned empty string")
	}
	// Default version is "0.1.0"
	expected := "0.1.0"
	if v != expected {
		t.Errorf("GetVersion() = %q, want %q", v, expected)
	}
}

func TestGetGitCommit(t *testing.T) {
	commit := GetGitCommit()
	if commit == "" {
		t.Error("GetGitCommit() returned empty string")
	}
	// Default is "unknown" when not set via ldflags
	expected := "unknown"
	if commit != expected {
		t.Errorf("GetGitCommit() = %q, want %q", commit, expected)
	}
}

func TestGetBuildDate(t *testing.T) {
	date := GetBuildDate()
	if date == "" {
		t.Error("GetBuildDate() returned empty string")
	}
	// Default is "unknown" when not set via ldflags
	expected := "unknown"
	if date != expected {
		t.Errorf("GetBuildDate() = %q, want %q", date, expected)
	}
}
