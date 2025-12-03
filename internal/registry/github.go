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
	HTTPClient *http.Client
	BaseURL    string
	token      string
}

// NewGitHubClient returns a GitHubClient configured with a 30-second HTTP timeout.
// The client will include an authentication token obtained from the GITHUB_TOKEN
// NewGitHubClient creates and returns a GitHubClient configured to access the GitHub API.
// The client has an HTTP client with a 30-second timeout and BaseURL set to https://api.github.com.
// It initializes the authentication token by reading GITHUB_TOKEN or, if unset, attempting to obtain it from the `gh` CLI.
func NewGitHubClient() *GitHubClient {
	token := getGitHubToken()
	return &GitHubClient{
		HTTPClient: &http.Client{Timeout: 30 * time.Second}, //nolint:mnd // timeout in seconds
		BaseURL:    "https://api.github.com",
		token:      token,
	}
}

// getGitHubToken returns a GitHub authentication token or an empty string if none is available.
// It first returns the value of the GITHUB_TOKEN environment variable, and if unset it attempts
// to obtain a token from the `gh auth token` command.
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
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.BaseURL, owner, repo)
	return c.fetchRelease(url)
}

// GetReleaseByTag returns a specific release by tag name.
func (c *GitHubClient) GetReleaseByTag(owner, repo, tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", c.BaseURL, owner, repo, tag)
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

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			return nil, errors.New("release not found")
		}
		if resp.StatusCode == http.StatusForbidden {
			_ = resp.Body.Close()
			return nil, errors.New("GitHub API rate limit exceeded. Set GITHUB_TOKEN for higher limits")
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}

		var release GitHubRelease
		if decodeErr := json.NewDecoder(resp.Body).Decode(&release); decodeErr != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
		}
		_ = resp.Body.Close()
		return &release, nil
	}
	return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// FindPlatformAsset locates the release asset matching the current OS and architecture for the given project.
// It expects assets named "{project}_{version}_{os}_{arch}.{ext}" where ext is ".zip" on Windows and ".tar.gz" otherwise.
// release is the GitHubRelease to search; projectName is the project name used in asset filenames.
// It returns the matching ReleaseAsset or an error that includes the available asset names when no match is found.
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

	resp, err := c.HTTPClient.Do(req)
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