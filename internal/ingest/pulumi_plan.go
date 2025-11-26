// Package ingest provides Pulumi plan parsing and resource mapping functionality.
// It converts Pulumi preview JSON output into resource descriptors for cost calculation.
package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rshade/pulumicost-core/internal/logging"
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
	return LoadPulumiPlanWithContext(context.Background(), path)
}

// LoadPulumiPlanWithContext loads and parses a Pulumi plan JSON file with logging context.
func LoadPulumiPlanWithContext(ctx context.Context, path string) (*PulumiPlan, error) {
	log := logging.FromContext(ctx)
	log.Debug().
		Ctx(ctx).
		Str("component", "ingest").
		Str("operation", "load_plan").
		Str("plan_path", path).
		Msg("loading Pulumi plan")

	data, err := os.ReadFile(path)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("component", "ingest").
			Err(err).
			Str("plan_path", path).
			Msg("failed to read plan file")
		return nil, fmt.Errorf("reading plan file: %w", err)
	}

	log.Debug().
		Ctx(ctx).
		Str("component", "ingest").
		Int("file_size_bytes", len(data)).
		Msg("plan file read successfully")

	var plan PulumiPlan
	if unmarshalErr := json.Unmarshal(data, &plan); unmarshalErr != nil {
		log.Error().
			Ctx(ctx).
			Str("component", "ingest").
			Err(unmarshalErr).
			Str("plan_path", path).
			Msg("failed to parse plan JSON")
		return nil, fmt.Errorf("parsing plan JSON: %w", unmarshalErr)
	}

	log.Debug().
		Ctx(ctx).
		Str("component", "ingest").
		Int("step_count", len(plan.Steps)).
		Msg("plan parsed successfully")

	return &plan, nil
}

// GetResources extracts all resources from the Pulumi plan steps.
func (p *PulumiPlan) GetResources() []PulumiResource {
	return p.GetResourcesWithContext(context.Background())
}

// GetResourcesWithContext extracts all resources from the Pulumi plan steps with logging context.
func (p *PulumiPlan) GetResourcesWithContext(ctx context.Context) []PulumiResource {
	log := logging.FromContext(ctx)
	var resources []PulumiResource
	var skippedOps []string

	for _, step := range p.Steps {
		if step.Op == "create" || step.Op == "update" || step.Op == "same" {
			resources = append(resources, PulumiResource{
				Type:     step.Type,
				URN:      step.URN,
				Provider: extractProviderFromURN(step.URN),
				Inputs:   step.Inputs,
			})
			log.Debug().
				Ctx(ctx).
				Str("component", "ingest").
				Str("resource_type", step.Type).
				Str("operation", step.Op).
				Str("urn", step.URN).
				Msg("extracted resource from plan")
		} else {
			skippedOps = append(skippedOps, step.Op)
		}
	}

	log.Debug().
		Ctx(ctx).
		Str("component", "ingest").
		Int("total_steps", len(p.Steps)).
		Int("extracted_resources", len(resources)).
		Int("skipped_operations", len(skippedOps)).
		Msg("resource extraction complete")

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
