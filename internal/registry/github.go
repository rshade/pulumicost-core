package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	osWindows = "windows"
	extExe    = ".exe"
	extZip    = ".zip"
	extTarGz  = ".tar.gz"
)

// GitHubRelease represents release metadata from GitHub API.
type GitHubRelease struct {
	TagName    string         `json:"tag_name"`
	Name       string         `json:"name"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	Assets     []ReleaseAsset `json:"assets"`
}

// ReleaseAsset represents a downloadable asset.
type ReleaseAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// GitHubClient provides GitHub API access for releases.
type GitHubClient struct {
	httpClient *http.Client
	token      string
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient() *GitHubClient {
	token := getGitHubToken()
	return &GitHubClient{
		httpClient: &http.Client{Timeout: 30 * time.Second}, //nolint:mnd // timeout in seconds
		token:      token,
	}
}

// getGitHubToken retrieves GitHub token from environment or gh CLI.
func getGitHubToken() string {
	// Try GITHUB_TOKEN first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	// Try gh auth token
	cmd := exec.Command("gh", "auth", "token") //nolint:noctx // trusted command, no context needed
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

// GetLatestRelease returns the latest release for a repository.
func (c *GitHubClient) GetLatestRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return c.fetchRelease(url)
}

// GetReleaseByTag returns a specific release by tag name.
func (c *GitHubClient) GetReleaseByTag(owner, repo, tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	return c.fetchRelease(url)
}

// fetchRelease fetches release data with retry logic.
//
//nolint:noctx // context not needed for simple HTTP
func (c *GitHubClient) fetchRelease(url string) (*GitHubRelease, error) {
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<attempt) * time.Second / 2) //nolint:mnd // backoff calculation
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if c.token != "" {
			req.Header.Set("Authorization", "token "+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, errors.New("release not found")
		}
		if resp.StatusCode == http.StatusForbidden {
			resp.Body.Close()
			return nil, errors.New("GitHub API rate limit exceeded. Set GITHUB_TOKEN for higher limits")
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}

		var release GitHubRelease
		if decodeErr := json.NewDecoder(resp.Body).Decode(&release); decodeErr != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
		}
		resp.Body.Close()
		return &release, nil
	}
	return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// FindPlatformAsset finds the appropriate asset for the current platform.
func FindPlatformAsset(release *GitHubRelease, projectName string) (*ReleaseAsset, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Determine expected extension
	ext := extTarGz
	if goos == osWindows {
		ext = extZip
	}

	// Build expected asset name pattern
	// Format: {projectname}_{version}_{os}_{arch}.{format}
	version := release.TagName
	expectedName := fmt.Sprintf("%s_%s_%s_%s%s", projectName, version, goos, goarch, ext)

	for i := range release.Assets {
		if release.Assets[i].Name == expectedName {
			return &release.Assets[i], nil
		}
	}

	// List available assets for error message
	var available []string
	for _, asset := range release.Assets {
		available = append(available, asset.Name)
	}

	return nil, fmt.Errorf("no asset found for %s/%s. Available: %v", goos, goarch, available)
}

// DownloadAsset downloads a release asset to a local file.
//
//nolint:mnd,noctx // magic numbers for buffer size, context not needed for downloads
func (c *GitHubClient) DownloadAsset(url, destPath string, progress func(downloaded, total int64)) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	total := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write: %w", writeErr)
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read error: %w", readErr)
		}
	}

	return nil
}
