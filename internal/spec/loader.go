package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Loader struct {
	specDir string
}

func NewLoader(specDir string) *Loader {
	return &Loader{specDir: specDir}
}

type PricingSpec struct {
	Provider  string                 `yaml:"provider"`
	Service   string                 `yaml:"service"`
	SKU       string                 `yaml:"sku"`
	Currency  string                 `yaml:"currency"`
	Pricing   map[string]interface{} `yaml:"pricing"`
	Metadata  map[string]interface{} `yaml:"metadata,omitempty"`
}

func (l *Loader) LoadSpec(provider, service, sku string) (interface{}, error) {
	filename := fmt.Sprintf("%s-%s-%s.yaml", provider, service, sku)
	path := filepath.Join(l.specDir, filename)
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading spec file: %w", err)
	}
	
	var spec PricingSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec YAML: %w", err)
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