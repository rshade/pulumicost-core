package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PulumiPlan struct {
	Steps []PulumiStep `json:"steps"`
}

type PulumiStep struct {
	Op       string                 `json:"op"`
	URN      string                 `json:"urn"`
	Type     string                 `json:"type"`
	Provider string                 `json:"provider"`
	Inputs   map[string]interface{} `json:"inputs"`
	Outputs  map[string]interface{} `json:"outputs"`
}

type PulumiResource struct {
	Type     string
	URN      string
	Provider string
	Inputs   map[string]interface{}
}

func LoadPulumiPlan(path string) (*PulumiPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading plan file: %w", err)
	}

	var plan PulumiPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parsing plan JSON: %w", err)
	}

	return &plan, nil
}

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
	if len(parts) >= 3 {
		providerParts := strings.Split(parts[2], ":")
		if len(providerParts) > 0 {
			return providerParts[0]
		}
	}
	return "unknown"
}
