package registry

import (
	"encoding/json"
	"os"
)

type Manifest struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Providers   []string          `json:"providers"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

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
