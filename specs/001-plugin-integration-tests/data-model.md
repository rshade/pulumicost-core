# Data Model: Plugin Integration Tests

**Status**: Phase 1 Design

## Test Entities

These structures are used within the test suite to define scenarios and mock responses.

### 1. Mock Registry Response

Represents the JSON structure served by the mock GitHub API.

```go
type MockRelease struct {
    TagName string      `json:"tag_name"`
    Assets  []MockAsset `json:"assets"`
}

type MockAsset struct {
    Name               string `json:"name"`
    BrowserDownloadURL string `json:"browser_download_url"`
    Size               int64  `json:"size"`
}
```

### 2. Test Case Definition

Standard structure for table-driven tests.

```go
type PluginTestCase struct {
    Name          string
    Args          []string
    Setup         func(t *testing.T, pluginDir string) // Optional setup (e.g., pre-install)
    MockHandler   http.HandlerFunc                     // Custom registry handler
    ExpectError   bool
    ErrorContains string
    Validate      func(t *testing.T, pluginDir string) // Post-execution validation
}
```

## Mock Registry Contract

The mock server mimics a subset of the GitHub REST API.

### `GET /repos/{owner}/{repo}/releases/latest`

**Response:**
```json
{
  "tag_name": "v1.0.0",
  "assets": [
    {
      "name": "plugin-v1.0.0-linux-amd64.tar.gz",
      "browser_download_url": "http://localhost:12345/download/plugin-v1.0.0-linux-amd64.tar.gz",
      "size": 1024
    }
  ]
}
```

### `GET /repos/{owner}/{repo}/releases/tags/{tag}`

**Response:** Same as above, but for a specific tag.

### `GET /download/{filename}`

**Response:** Binary content (application/octet-stream). The test suite will serve a valid dummy zip/tar.gz file here containing a mock executable.
