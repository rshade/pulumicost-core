package pluginsdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDefaultManifest(t *testing.T) {
	manifest := CreateDefaultManifest("test-plugin", "Test Author", []string{"aws", "azure"})

	if manifest.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got %s", manifest.Name)
	}

	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", manifest.Version)
	}

	if manifest.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got %s", manifest.Author)
	}

	if len(manifest.SupportedProviders) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(manifest.SupportedProviders))
	}

	if manifest.SupportedProviders[0] != "aws" || manifest.SupportedProviders[1] != "azure" {
		t.Errorf("Expected providers [aws, azure], got %v", manifest.SupportedProviders)
	}
}

func TestManifestValidation(t *testing.T) {
	testCases := []struct {
		name        string
		manifest    *Manifest
		expectError bool
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				Name:               "test-plugin",
				Version:            "1.0.0",
				Description:        "Test plugin",
				Author:             "Test Author",
				SupportedProviders: []string{"aws"},
				Protocols:          []string{"grpc"},
				Binary:             "./bin/test",
			},
			expectError: false,
		},
		{
			name: "missing name",
			manifest: &Manifest{
				Version:            "1.0.0",
				Description:        "Test plugin",
				Author:             "Test Author",
				SupportedProviders: []string{"aws"},
				Protocols:          []string{"grpc"},
				Binary:             "./bin/test",
			},
			expectError: true,
		},
		{
			name: "invalid name",
			manifest: &Manifest{
				Name:               "Test_Plugin",
				Version:            "1.0.0",
				Description:        "Test plugin",
				Author:             "Test Author",
				SupportedProviders: []string{"aws"},
				Protocols:          []string{"grpc"},
				Binary:             "./bin/test",
			},
			expectError: true,
		},
		{
			name: "invalid version",
			manifest: &Manifest{
				Name:               "test-plugin",
				Version:            "1.0",
				Description:        "Test plugin",
				Author:             "Test Author",
				SupportedProviders: []string{"aws"},
				Protocols:          []string{"grpc"},
				Binary:             "./bin/test",
			},
			expectError: true,
		},
		{
			name: "missing providers",
			manifest: &Manifest{
				Name:               "test-plugin",
				Version:            "1.0.0",
				Description:        "Test plugin",
				Author:             "Test Author",
				SupportedProviders: []string{},
				Protocols:          []string{"grpc"},
				Binary:             "./bin/test",
			},
			expectError: true,
		},
		{
			name: "invalid protocol",
			manifest: &Manifest{
				Name:               "test-plugin",
				Version:            "1.0.0",
				Description:        "Test plugin",
				Author:             "Test Author",
				SupportedProviders: []string{"aws"},
				Protocols:          []string{"http"},
				Binary:             "./bin/test",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.manifest.Validate()
			if tc.expectError && err == nil {
				t.Errorf("Expected validation error, got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestManifestSaveLoad(t *testing.T) {
	manifest := CreateDefaultManifest("test-plugin", "Test Author", []string{"aws"})

	// Test YAML format
	t.Run("YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		yamlPath := filepath.Join(tmpDir, "test.yaml")

		// Save manifest
		err := manifest.SaveManifest(yamlPath)
		if err != nil {
			t.Fatalf("Failed to save YAML manifest: %v", err)
		}

		// Load manifest
		loaded, err := LoadManifest(yamlPath)
		if err != nil {
			t.Fatalf("Failed to load YAML manifest: %v", err)
		}

		// Compare
		if loaded.Name != manifest.Name {
			t.Errorf("Names don't match: expected %s, got %s", manifest.Name, loaded.Name)
		}
	})

	// Test JSON format
	t.Run("JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonPath := filepath.Join(tmpDir, "test.json")

		// Save manifest
		err := manifest.SaveManifest(jsonPath)
		if err != nil {
			t.Fatalf("Failed to save JSON manifest: %v", err)
		}

		// Load manifest
		loaded, err := LoadManifest(jsonPath)
		if err != nil {
			t.Fatalf("Failed to load JSON manifest: %v", err)
		}

		// Compare
		if loaded.Name != manifest.Name {
			t.Errorf("Names don't match: expected %s, got %s", manifest.Name, loaded.Name)
		}
	})

	// Test unsupported format
	t.Run("Unsupported", func(t *testing.T) {
		tmpDir := t.TempDir()
		txtPath := filepath.Join(tmpDir, "test.txt")

		err := manifest.SaveManifest(txtPath)
		if err == nil {
			t.Errorf("Expected error for unsupported format, got none")
		}
	})
}

func TestLoadManifestErrors(t *testing.T) {
	// Test non-existent file
	_, err := LoadManifest("non-existent.yaml")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got none")
	}

	// Test invalid YAML
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	err = os.WriteFile(invalidPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	_, err = LoadManifest(invalidPath)
	if err == nil {
		t.Errorf("Expected error for invalid YAML, got none")
	}
}