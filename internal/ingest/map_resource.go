package ingest

import (
	"fmt"
	"strings"

	"github.com/rshade/pulumicost-core/internal/engine"
)

const unknownProvider = "unknown"

func MapResource(pulumiResource PulumiResource) (engine.ResourceDescriptor, error) {
	provider := extractProvider(pulumiResource.Type)

	return engine.ResourceDescriptor{
		Type:       pulumiResource.Type,
		ID:         pulumiResource.URN,
		Provider:   provider,
		Properties: pulumiResource.Inputs,
	}, nil
}

func extractProvider(resourceType string) string {
	parts := strings.Split(resourceType, ":")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return unknownProvider
}

func MapResources(resources []PulumiResource) ([]engine.ResourceDescriptor, error) {
	var descriptors []engine.ResourceDescriptor
	for _, r := range resources {
		desc, err := MapResource(r)
		if err != nil {
			return nil, fmt.Errorf("mapping resource %s: %w", r.URN, err)
		}
		descriptors = append(descriptors, desc)
	}
	return descriptors, nil
}
