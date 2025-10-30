// Package registry implements plugin discovery and lifecycle management for PulumiCost.
// It scans the filesystem for installed plugins following the ~/.pulumicost/plugins/<name>/<version>/
// directory structure and manages plugin connections via gRPC.
package registry

import (
	"encoding/json"
	"os"
)

// Manifest represents the optional plugin.manifest.json metadata file.
// It provides additional plugin information and validation data.
type Manifest struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Providers   []string          `json:"providers"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// LoadManifest loads and parses a plugin manifest JSON file from the specified path.
// It returns an error if the file doesn't exist or contains invalid JSON.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if unmarshalErr := json.Unmarshal(data, &manifest); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &manifest, nil
}
