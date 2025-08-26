package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rshade/pulumicost-core/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSpec(t *testing.T) {
	tests := []struct {
		name        string
		specContent string
		provider    string
		service     string
		sku         string
		expectError bool
	}{
		{
			name: "valid yaml spec",
			specContent: `
provider: aws
service: ec2
sku: t3.micro
currency: USD
hourly_cost: 0.0104
monthly_cost: 7.592
region: us-east-1
`,
			provider:    "aws",
			service:     "ec2", 
			sku:         "t3.micro",
			expectError: false,
		},
		{
			name: "invalid yaml spec",
			specContent: `
invalid: yaml: structure
  - missing
    - proper
`,
			provider:    "aws",
			service:     "ec2",
			sku:         "invalid",
			expectError: true,
		},
		{
			name:        "non-existent spec",
			specContent: "",
			provider:    "nonexistent",
			service:     "service",
			sku:         "sku",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for specs
			tempDir := t.TempDir()
			specDir := filepath.Join(tempDir, "specs")
			err := os.MkdirAll(specDir, 0755)
			require.NoError(t, err)

			// Create spec file if content provided
			if tt.specContent != "" {
				specFile := filepath.Join(specDir, tt.provider+"-"+tt.service+"-"+tt.sku+".yaml")
				err := os.WriteFile(specFile, []byte(tt.specContent), 0644)
				require.NoError(t, err)
			}

			loader := spec.NewLoader(specDir)
			result, err := loader.LoadSpec(tt.provider, tt.service, tt.sku)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestValidateSpec(t *testing.T) {
	tests := []struct {
		name        string
		specData    map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid spec",
			specData: map[string]interface{}{
				"provider":     "aws",
				"service":      "ec2",
				"sku":          "t3.micro",
				"currency":     "USD",
				"hourly_cost":  0.0104,
				"monthly_cost": 7.592,
				"region":       "us-east-1",
			},
			expectError: false,
		},
		{
			name: "missing required fields",
			specData: map[string]interface{}{
				"provider": "aws",
				// missing service, sku, currency
			},
			expectError: true,
			errorMsg:    "missing required field",
		},
		{
			name: "invalid cost values",
			specData: map[string]interface{}{
				"provider":     "aws",
				"service":      "ec2", 
				"sku":          "t3.micro",
				"currency":     "USD",
				"hourly_cost":  -1.0, // negative cost
				"monthly_cost": 7.592,
			},
			expectError: true,
			errorMsg:    "negative cost",
		},
		{
			name: "invalid currency",
			specData: map[string]interface{}{
				"provider":     "aws",
				"service":      "ec2",
				"sku":          "t3.micro", 
				"currency":     "INVALID",
				"hourly_cost":  0.0104,
				"monthly_cost": 7.592,
			},
			expectError: true,
			errorMsg:    "invalid currency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := spec.ValidateSpec(tt.specData)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestSpecFilePatterns(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		provider string
		service  string
		sku      string
		isValid  bool
	}{
		{
			name:     "standard pattern",
			filename: "aws-ec2-t3.micro.yaml",
			provider: "aws",
			service:  "ec2",
			sku:      "t3.micro",
			isValid:  true,
		},
		{
			name:     "with dashes in sku",
			filename: "azure-compute-standard-d2s-v3.yaml",
			provider: "azure",
			service:  "compute",
			sku:      "standard-d2s-v3",
			isValid:  true,
		},
		{
			name:     "yml extension",
			filename: "gcp-compute-n1-standard-1.yml",
			provider: "gcp",
			service:  "compute", 
			sku:      "n1-standard-1",
			isValid:  true,
		},
		{
			name:     "invalid extension",
			filename: "aws-ec2-t3.micro.json",
			isValid:  false,
		},
		{
			name:     "invalid pattern",
			filename: "aws-ec2.yaml",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, service, sku, valid := spec.ParseSpecFilename(tt.filename)

			assert.Equal(t, tt.isValid, valid)
			
			if tt.isValid {
				assert.Equal(t, tt.provider, provider)
				assert.Equal(t, tt.service, service)
				assert.Equal(t, tt.sku, sku)
			}
		})
	}
}

func TestSpecLoader(t *testing.T) {
	t.Run("creates new loader with directory", func(t *testing.T) {
		tempDir := t.TempDir()
		loader := spec.NewLoader(tempDir)
		assert.NotNil(t, loader)
	})

	t.Run("handles empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		loader := spec.NewLoader(tempDir)
		
		result, err := loader.LoadSpec("aws", "ec2", "t3.micro")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles non-existent directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
		loader := spec.NewLoader(nonExistentDir)
		
		result, err := loader.LoadSpec("aws", "ec2", "t3.micro")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSpecCache(t *testing.T) {
	t.Run("caches loaded specs", func(t *testing.T) {
		// Create a temporary directory for specs
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, "specs")
		err := os.MkdirAll(specDir, 0755)
		require.NoError(t, err)

		// Create a spec file
		specContent := `
provider: aws
service: ec2
sku: t3.micro
currency: USD
hourly_cost: 0.0104
monthly_cost: 7.592
region: us-east-1
`
		specFile := filepath.Join(specDir, "aws-ec2-t3.micro.yaml")
		err = os.WriteFile(specFile, []byte(specContent), 0644)
		require.NoError(t, err)

		loader := spec.NewLoader(specDir)
		
		// Load spec twice
		result1, err1 := loader.LoadSpec("aws", "ec2", "t3.micro")
		result2, err2 := loader.LoadSpec("aws", "ec2", "t3.micro")

		// Both should succeed
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotNil(t, result1)
		assert.NotNil(t, result2)

		// Results should be equivalent (though may not be identical due to caching)
		assert.Equal(t, result1, result2)
	})
}

func TestSpecErrorHandling(t *testing.T) {
	t.Run("handles file permission errors", func(t *testing.T) {
		// Create a temporary directory for specs
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, "specs")
		err := os.MkdirAll(specDir, 0755)
		require.NoError(t, err)

		// Create a spec file with restricted permissions
		specFile := filepath.Join(specDir, "aws-ec2-t3.micro.yaml")
		err = os.WriteFile(specFile, []byte("test: data"), 0000) // No permissions
		require.NoError(t, err)

		loader := spec.NewLoader(specDir)
		result, err := loader.LoadSpec("aws", "ec2", "t3.micro")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles malformed yaml", func(t *testing.T) {
		// Create a temporary directory for specs
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, "specs")
		err := os.MkdirAll(specDir, 0755)
		require.NoError(t, err)

		// Create a malformed YAML file
		malformedYAML := `
invalid: yaml
  - missing
    colon here
    - another: item
`
		specFile := filepath.Join(specDir, "aws-ec2-invalid.yaml")
		err = os.WriteFile(specFile, []byte(malformedYAML), 0644)
		require.NoError(t, err)

		loader := spec.NewLoader(specDir)
		result, err := loader.LoadSpec("aws", "ec2", "invalid")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "yaml") // Should mention YAML parsing issue
	})
}