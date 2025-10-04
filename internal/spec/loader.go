package spec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrSpecNotFound = errors.New("spec file not found")
)

type Loader struct {
	specDir string
}

func NewLoader(specDir string) *Loader {
	return &Loader{specDir: specDir}
}

type PricingSpec struct {
	Provider string                 `yaml:"provider"`
	Service  string                 `yaml:"service"`
	SKU      string                 `yaml:"sku"`
	Currency string                 `yaml:"currency"`
	Pricing  map[string]interface{} `yaml:"pricing"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

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
func ParseSpecFilename(filename string) (provider, service, sku string, valid bool) {
	// Remove extension
	ext := filepath.Ext(filename)
	if ext != ".yaml" && ext != ".yml" {
		return "", "", "", false
	}
	
	name := filename[:len(filename)-len(ext)]
	
	// Split by dash - should be provider-service-sku format
	parts := strings.Split(name, "-")
	if len(parts) < 3 {
		return "", "", "", false
	}
	
	provider = parts[0]
	service = parts[1]
	// Join remaining parts as SKU (in case SKU contains dashes)
	sku = strings.Join(parts[2:], "-")
	
	return provider, service, sku, true
}
