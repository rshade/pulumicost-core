package spec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// ExpectedPartsCount is the number of parts expected in a spec filename (provider-service-sku).
	ExpectedPartsCount = 3
)

var (
	// ErrSpecNotFound is returned when a requested spec file does not exist.
	ErrSpecNotFound = errors.New("spec file not found")
)

// Loader loads pricing specifications from a directory.
type Loader struct {
	specDir string
}

// NewLoader creates a new spec loader for the given directory.
func NewLoader(specDir string) *Loader {
	return &Loader{specDir: specDir}
}

// PricingSpec represents a pricing specification for a cloud service SKU.
type PricingSpec struct {
	Provider string                 `yaml:"provider"`
	Service  string                 `yaml:"service"`
	SKU      string                 `yaml:"sku"`
	Currency string                 `yaml:"currency"`
	Pricing  map[string]interface{} `yaml:"pricing"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// LoadSpec loads a pricing specification by provider, service, and SKU.
func (l *Loader) LoadSpec(provider, service, sku string) (interface{}, error) {
	filename := fmt.Sprintf("%s-%s-%s.yaml", provider, service, sku)
	path := filepath.Join(l.specDir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSpecNotFound
		}
		return nil, fmt.Errorf("reading spec file: %w", err)
	}

	var spec PricingSpec
	if unmarshalErr := yaml.Unmarshal(data, &spec); unmarshalErr != nil {
		return nil, fmt.Errorf("parsing spec YAML: %w", unmarshalErr)
	}

	return &spec, nil
}

// ListSpecs returns a list of all available spec filenames in the spec directory.
func (l *Loader) ListSpecs() ([]string, error) {
	entries, err := os.ReadDir(l.specDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading spec directory: %w", err)
	}

	var specs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			specs = append(specs, entry.Name())
		}
	}

	return specs, nil
}

// ParseSpecFilename parses a spec filename to extract provider, service, and SKU.
func ParseSpecFilename(filename string) (string, string, string, bool) {
	// Remove extension
	ext := filepath.Ext(filename)
	if ext != ".yaml" && ext != ".yml" {
		return "", "", "", false
	}

	name := filename[:len(filename)-len(ext)]

	// Split by dash - should be provider-service-sku format
	parts := strings.Split(name, "-")
	if len(parts) < ExpectedPartsCount {
		return "", "", "", false
	}

	provider := parts[0]
	service := parts[1]
	// Join remaining parts as SKU (in case SKU contains dashes)
	sku := strings.Join(parts[2:], "-")

	// Validate that all parts are non-empty
	if provider == "" || service == "" || sku == "" {
		return "", "", "", false
	}

	return provider, service, sku, true
}
