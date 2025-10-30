// Package ingest provides Pulumi plan parsing and resource mapping functionality.
// It converts Pulumi preview JSON output into resource descriptors for cost calculation.
package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	minURNParts = 3
)

// PulumiPlan represents the top-level structure of a Pulumi preview JSON output.
type PulumiPlan struct {
	Steps []PulumiStep `json:"steps"`
}

// PulumiStep represents a single resource operation step in a Pulumi plan.
type PulumiStep struct {
	Op       string                 `json:"op"`
	URN      string                 `json:"urn"`
	Type     string                 `json:"type"`
	Provider string                 `json:"provider"`
	Inputs   map[string]interface{} `json:"inputs"`
	Outputs  map[string]interface{} `json:"outputs"`
}

// PulumiResource contains the detailed information about a resource in a Pulumi step.
type PulumiResource struct {
	Type     string
	URN      string
	Provider string
	Inputs   map[string]interface{}
}

// LoadPulumiPlan loads and parses a Pulumi plan JSON file from the specified path.
func LoadPulumiPlan(path string) (*PulumiPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading plan file: %w", err)
	}

	var plan PulumiPlan
	if unmarshalErr := json.Unmarshal(data, &plan); unmarshalErr != nil {
		return nil, fmt.Errorf("parsing plan JSON: %w", unmarshalErr)
	}

	return &plan, nil
}

// GetResources extracts all resources from the Pulumi plan steps.
func (p *PulumiPlan) GetResources() []PulumiResource {
	var resources []PulumiResource
	for _, step := range p.Steps {
		if step.Op == "create" || step.Op == "update" || step.Op == "same" {
			resources = append(resources, PulumiResource{
				Type:     step.Type,
				URN:      step.URN,
				Provider: extractProviderFromURN(step.URN),
				Inputs:   step.Inputs,
			})
		}
	}
	return resources
}

func extractProviderFromURN(urn string) string {
	parts := strings.Split(urn, "::")
	if len(parts) >= minURNParts {
		providerParts := strings.Split(parts[2], ":")
		if len(providerParts) > 0 && providerParts[0] != "" {
			return providerParts[0]
		}
	}
	return unknownProvider
}
