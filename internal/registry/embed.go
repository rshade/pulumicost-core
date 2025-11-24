package registry

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed registry.json
var registryData []byte

// EmbeddedRegistry represents the embedded plugin registry catalog.
type EmbeddedRegistry struct {
	SchemaVersion string                   `json:"schema_version"`
	Plugins       map[string]RegistryEntry `json:"plugins"`
}

//nolint:gochecknoglobals // sync.Once pattern for lazy loading
var (
	embeddedRegistry     *EmbeddedRegistry
	embeddedRegistryOnce sync.Once
	errEmbeddedRegistry  error
)

// GetEmbeddedRegistry returns the parsed embedded registry catalog.
func GetEmbeddedRegistry() (*EmbeddedRegistry, error) {
	embeddedRegistryOnce.Do(func() {
		embeddedRegistry = &EmbeddedRegistry{}
		errEmbeddedRegistry = json.Unmarshal(registryData, embeddedRegistry)
		if errEmbeddedRegistry != nil {
			errEmbeddedRegistry = fmt.Errorf("failed to parse embedded registry: %w", errEmbeddedRegistry)
		}
	})
	return embeddedRegistry, errEmbeddedRegistry
}

// GetPlugin returns a registry entry by name from the embedded catalog.
func GetPlugin(name string) (*RegistryEntry, error) {
	reg, err := GetEmbeddedRegistry()
	if err != nil {
		return nil, err
	}
	entry, ok := reg.Plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found in registry", name)
	}
	return &entry, nil
}

// ListPluginsFromRegistry returns all plugin names from the embedded registry.
func ListPluginsFromRegistry() ([]string, error) {
	reg, err := GetEmbeddedRegistry()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(reg.Plugins))
	for name := range reg.Plugins {
		names = append(names, name)
	}
	return names, nil
}

// GetAllPluginEntries returns all registry entries from the embedded catalog.
func GetAllPluginEntries() ([]RegistryEntry, error) {
	reg, err := GetEmbeddedRegistry()
	if err != nil {
		return nil, err
	}
	plugins := make([]RegistryEntry, 0, len(reg.Plugins))
	for _, entry := range reg.Plugins {
		plugins = append(plugins, entry)
	}
	return plugins, nil
}
