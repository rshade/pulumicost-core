package registry

import (
	"testing"
)

func TestValidateRegistryEntry(t *testing.T) {
	tests := []struct {
		name    string
		entry   RegistryEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: RegistryEntry{
				Name:          "test-plugin",
				Repository:    "owner/repo",
				SecurityLevel: "community",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			entry: RegistryEntry{
				Repository: "owner/repo",
			},
			wantErr: true,
		},
		{
			name: "missing repository",
			entry: RegistryEntry{
				Name: "test-plugin",
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			entry: RegistryEntry{
				Name:       "test-plugin",
				Repository: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid security level",
			entry: RegistryEntry{
				Name:          "test-plugin",
				Repository:    "owner/repo",
				SecurityLevel: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid official security level",
			entry: RegistryEntry{
				Name:          "test-plugin",
				Repository:    "owner/repo",
				SecurityLevel: "official",
			},
			wantErr: false,
		},
		{
			name: "valid experimental security level",
			entry: RegistryEntry{
				Name:          "test-plugin",
				Repository:    "owner/repo",
				SecurityLevel: "experimental",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryEntry(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegistryEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsePluginSpecifier(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		wantName    string
		wantVersion string
		wantIsURL   bool
		wantErr     bool
	}{
		{
			name:        "simple name",
			spec:        "kubecost",
			wantName:    "kubecost",
			wantVersion: "",
			wantIsURL:   false,
		},
		{
			name:        "name with version",
			spec:        "kubecost@v1.0.0",
			wantName:    "kubecost",
			wantVersion: "v1.0.0",
			wantIsURL:   false,
		},
		{
			name:        "github url",
			spec:        "github.com/owner/pulumicost-plugin-test",
			wantName:    "test",
			wantVersion: "",
			wantIsURL:   true,
		},
		{
			name:        "github url with version",
			spec:        "github.com/owner/repo@v2.0.0",
			wantName:    "repo",
			wantVersion: "v2.0.0",
			wantIsURL:   true,
		},
		{
			name:    "empty spec",
			spec:    "",
			wantErr: true,
		},
		{
			name:    "invalid github url",
			spec:    "github.com/invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePluginSpecifier(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePluginSpecifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if result.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantName)
			}
			if result.Version != tt.wantVersion {
				t.Errorf("Version = %v, want %v", result.Version, tt.wantVersion)
			}
			if result.IsURL != tt.wantIsURL {
				t.Errorf("IsURL = %v, want %v", result.IsURL, tt.wantIsURL)
			}
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid url",
			url:       "github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "invalid format",
			url:     "github.com/invalid",
			wantErr: true,
		},
		{
			name:    "wrong domain",
			url:     "gitlab.com/owner/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseGitHubURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("owner = %v, want %v", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}
