package pluginsdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Manifest represents the plugin manifest structure.
type Manifest struct {
	Name               string            `yaml:"name" json:"name"`
	Version            string            `yaml:"version" json:"version"`
	Description        string            `yaml:"description" json:"description"`
	Author             string            `yaml:"author" json:"author"`
	SupportedProviders []string          `yaml:"supported_providers" json:"supported_providers"`
	Protocols          []string          `yaml:"protocols" json:"protocols"`
	Binary             string            `yaml:"binary" json:"binary"`
	Metadata           map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// ValidationError represents a manifest validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return "no validation errors"
	}
	
	msg := fmt.Sprintf("validation failed with %d error(s):", len(errs))
	for _, err := range errs {
		msg += fmt.Sprintf("\n  - %s", err.Error())
	}
	return msg
}

// LoadManifest loads a plugin manifest from a file path.
// Supports both YAML and JSON formats based on file extension.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}

	var manifest Manifest
	ext := filepath.Ext(path)
	
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("parsing YAML manifest: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("parsing JSON manifest: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported manifest file extension: %s (supported: .yaml, .yml, .json)", ext)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves a plugin manifest to a file path.
// Format is determined by file extension.
func (m *Manifest) SaveManifest(path string) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	ext := filepath.Ext(path)
	var data []byte
	var err error
	
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(m)
		if err != nil {
			return fmt.Errorf("marshaling to YAML: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling to JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported manifest file extension: %s (supported: .yaml, .yml, .json)", ext)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing manifest file: %w", err)
	}

	return nil
}

// Validate validates the manifest fields.
func (m *Manifest) Validate() error {
	var errors ValidationErrors

	// Required fields
	if m.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if !isValidName(m.Name) {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name must contain only lowercase letters, numbers, and hyphens",
		})
	}

	if m.Version == "" {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: "version is required",
		})
	} else if !isValidVersion(m.Version) {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: "version must follow semantic versioning (e.g., 1.0.0)",
		})
	}

	if m.Description == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "description is required",
		})
	}

	if m.Author == "" {
		errors = append(errors, ValidationError{
			Field:   "author",
			Message: "author is required",
		})
	}

	if len(m.SupportedProviders) == 0 {
		errors = append(errors, ValidationError{
			Field:   "supported_providers",
			Message: "at least one supported provider is required",
		})
	}

	if len(m.Protocols) == 0 {
		errors = append(errors, ValidationError{
			Field:   "protocols",
			Message: "at least one protocol is required",
		})
	} else {
		validProtocols := map[string]bool{"grpc": true}
		for _, protocol := range m.Protocols {
			if !validProtocols[protocol] {
				errors = append(errors, ValidationError{
					Field:   "protocols",
					Message: fmt.Sprintf("unsupported protocol: %s (supported: grpc)", protocol),
				})
			}
		}
	}

	if m.Binary == "" {
		errors = append(errors, ValidationError{
			Field:   "binary",
			Message: "binary path is required",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// CreateDefaultManifest creates a default manifest for a plugin.
func CreateDefaultManifest(name, author string, providers []string) *Manifest {
	return &Manifest{
		Name:               name,
		Version:            "1.0.0",
		Description:        fmt.Sprintf("PulumiCost plugin for %s", name),
		Author:             author,
		SupportedProviders: providers,
		Protocols:          []string{"grpc"},
		Binary:             fmt.Sprintf("./bin/pulumicost-plugin-%s", name),
		Metadata:           make(map[string]string),
	}
}

// Regular expressions for validation
var (
	nameRegex    = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)
	versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9-]+)?(\+[a-zA-Z0-9-]+)?$`)
)

func isValidName(name string) bool {
	if len(name) < 2 || len(name) > 50 {
		return false
	}
	return nameRegex.MatchString(name)
}

func isValidVersion(version string) bool {
	return versionRegex.MatchString(version)
}