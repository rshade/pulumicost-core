package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestGetLatestRelease(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			t.Errorf("Expected path /repos/owner/repo/releases/latest, got %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		release := GitHubRelease{
			TagName: "v1.0.0",
			Name:    "Release 1.0.0",
			Assets:  []ReleaseAsset{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Configure client to use mock server
	// Note: Since GetLatestRelease constructs the URL using "https://api.github.com",
	// we can't easily redirect it to our mock server unless we override the base URL.
	// However, fetchRelease takes a full URL.
	// So we should test fetchRelease directly or refactor GetLatestRelease to allow base URL override.

	// For now, let's test fetchRelease directly as it is the core logic.
	client := NewGitHubClient()
	client.HTTPClient = server.Client()

	release, err := client.fetchRelease(server.URL + "/repos/owner/repo/releases/latest")
	if err != nil {
		t.Fatalf("fetchRelease failed: %v", err)
	}

	if release.TagName != "v1.0.0" {
		t.Errorf("Expected tag v1.0.0, got %s", release.TagName)
	}
}

func TestGetReleaseByTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/tags/v1.0.0" {
			http.NotFound(w, r)
			return
		}

		release := GitHubRelease{
			TagName: "v1.0.0",
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.HTTPClient = server.Client()

	// Testing fetchRelease with constructed URL from test
	release, err := client.fetchRelease(server.URL + "/repos/owner/repo/releases/tags/v1.0.0")
	if err != nil {
		t.Fatalf("fetchRelease failed: %v", err)
	}

	if release.TagName != "v1.0.0" {
		t.Errorf("Expected tag v1.0.0, got %s", release.TagName)
	}
}

func TestDownloadAsset(t *testing.T) {
	content := "binary content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.Write([]byte(content))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.HTTPClient = server.Client()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "asset.bin")

	err := client.DownloadAsset(server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadAsset failed: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}
}

func TestFetchRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.HTTPClient = server.Client()

	_, err := client.fetchRelease(server.URL)
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestFetchRelease_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.HTTPClient = server.Client()

	_, err := client.fetchRelease(server.URL)
	if err == nil {
		t.Error("Expected error for 403, got nil")
	}
}
