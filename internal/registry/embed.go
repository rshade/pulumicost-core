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
// It initializes and parses the embedded registry data on first call in a thread-safe manner.
// Returns the parsed *EmbeddedRegistry, or nil and a non-nil error if parsing the embedded registry failed.
func GetEmbeddedRegistry() (*EmbeddedRegistry, error) {
	embeddedRegistryOnce.Do(func() {
		embeddedRegistry = &EmbeddedRegistry{}
		errEmbeddedRegistry = json.Unmarshal(registryData, embeddedRegistry)
		if errEmbeddedRegistry != nil {
			errEmbeddedRegistry = fmt.Errorf(
				"failed to parse embedded registry: %w",
				errEmbeddedRegistry,
			)
		}
	})
	return embeddedRegistry, errEmbeddedRegistry
}

// GetPlugin looks up a plugin by name in the embedded registry and returns its RegistryEntry.
// The name parameter is the plugin identifier to look up.
// It returns a pointer to the RegistryEntry if the plugin exists.
// If parsing the embedded registry fails, an error from GetEmbeddedRegistry is returned.
// If the plugin is not present in the registry, an error indicating the plugin was not found is returned.
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

// ListPluginsFromRegistry returns the names of all plugins contained in the embedded registry.
// It loads the embedded registry and collects each plugin key into a slice.
// The returned error is non-nil if the embedded registry cannot be loaded or parsed.
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

// GetAllPluginEntries retrieves all plugin entries present in the embedded registry catalog.
// It returns a slice containing every RegistryEntry found in the embedded catalog.
// If the embedded registry cannot be loaded or parsed, an error is returned.
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
