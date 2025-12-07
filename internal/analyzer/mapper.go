package analyzer

import (
	"errors"
	"fmt"
	"strings"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/rshade/pulumicost-core/internal/engine"
	"google.golang.org/protobuf/types/known/structpb"
)

// URN parsing constants.
const (
	// minURNParts is the minimum number of parts in a URN to extract the resource name.
	// URN format: urn:pulumi:stack::project::type::name (6 parts minimum, we need at least 2).
	minURNParts = 2

	// minProviderTypeParts is the minimum number of parts in a provider type.
	// Format: pulumi:providers:aws (3 parts).
	minProviderTypeParts = 3
)

// MappingError represents an error that occurred during resource mapping.
type MappingError struct {
	Index   int    // Position in original slice
	URN     string // Resource URN if available
	Type    string // Resource type if available
	Message string // Error description
	Err     error  // Underlying error
}

// Error implements the error interface.
func (e *MappingError) Error() string {
	if e.URN != "" {
		return "mapping " + e.URN + ": " + e.Message
	}
	return fmt.Sprintf("mapping resource at index %d: %s", e.Index, e.Message)
}

// Unwrap returns the underlying error.
func (e *MappingError) Unwrap() error {
	return e.Err
}

// ErrNilResource is returned when a nil resource is encountered.
var ErrNilResource = errors.New("nil resource")

// MappingResult contains the results of mapping resources with error tracking.
type MappingResult struct {
	Resources []engine.ResourceDescriptor // Successfully mapped resources
	Errors    []MappingError              // Errors encountered during mapping
	Skipped   int                         // Count of skipped resources
}

// MapResource converts a pulumirpc.AnalyzerResource to an engine.ResourceDescriptor.
//
// The mapping extracts cost-relevant fields from the Pulumi resource representation
// and normalizes them to the internal format used by the cost calculation engine.
//
// Field mappings:
//   - Type: Direct copy from r.Type
//   - ID: Extracted from URN (last :: segment)
//   - Provider: Extracted from provider resource type or resource type prefix
// MapResource converts a Pulumi analyzer resource into an engine.ResourceDescriptor.
// The returned ResourceDescriptor has Type set from the resource type, ID extracted from the resource URN, Provider derived from the resource provider or type, and Properties converted from the resource's protobuf Struct.
// The input r must be non-nil.
func MapResource(r *pulumirpc.AnalyzerResource) engine.ResourceDescriptor {
	return engine.ResourceDescriptor{
		Type:       r.GetType(),
		ID:         extractResourceID(r.GetUrn()),
		Provider:   extractProvider(r),
		Properties: structToMap(r.GetProperties()),
	}
}

// MapResources converts a slice of AnalyzerResource to ResourceDescriptors.
//
// This is the primary entry point for batch resource mapping. All resources
// are converted, and any individual mapping failures are handled gracefully
// (the resource is included with best-effort field extraction).
//
// Note: This function skips nil resources silently. Use MapResourcesWithErrors
// MapResources converts a slice of Pulumi analyzer resources into a slice of engine.ResourceDescriptor.
// It maps each non-nil resource in the input in order and skips any nil entries.
// If the input slice is empty, MapResources returns nil.
func MapResources(resources []*pulumirpc.AnalyzerResource) []engine.ResourceDescriptor {
	if len(resources) == 0 {
		return nil
	}

	result := make([]engine.ResourceDescriptor, 0, len(resources))
	for _, r := range resources {
		if r == nil {
			continue
		}
		result = append(result, MapResource(r))
	}
	return result
}

// MapResourcesWithErrors converts resources with explicit error tracking.
//
// This function provides detailed error information for each resource that
// fails to map. Use this when you need visibility into mapping failures
// for diagnostics or debugging.
//
// Graceful degradation: nil resources are skipped and counted, valid
// MapResourcesWithErrors maps a slice of AnalyzerResource values into a MappingResult containing successfully mapped
// resource descriptors and per-resource mapping errors.
// MapResourcesWithErrors treats nil slice entries as skipped: each nil entry increments Skipped and appends a
// MappingError (with Err set to ErrNilResource and Index set to the entry position). Non-nil entries are converted
// via MapResource and appended to Resources. The Errors slice records any per-item issues with their input index.
func MapResourcesWithErrors(resources []*pulumirpc.AnalyzerResource) MappingResult {
	result := MappingResult{
		Resources: make([]engine.ResourceDescriptor, 0, len(resources)),
		Errors:    make([]MappingError, 0),
	}

	for i, r := range resources {
		if r == nil {
			result.Skipped++
			result.Errors = append(result.Errors, MappingError{
				Index:   i,
				Message: "nil resource pointer",
				Err:     ErrNilResource,
			})
			continue
		}

		result.Resources = append(result.Resources, MapResource(r))
	}

	return result
}

// extractResourceID extracts the resource name from a Pulumi URN.
//
// URN format: urn:pulumi:stack::project::type::name
//
// Examples:
//   - "urn:pulumi:dev::myapp::aws:ec2/instance:Instance::webserver" → "webserver"
//   - "urn:pulumi:prod::api::azure:compute/vm:VM::api-server-01" → "api-server-01"
//   - "" → ""
// extractResourceID extracts the resource name portion from a Pulumi URN.
// If urn is empty it returns an empty string. If urn contains "::" separators,
// it returns the last segment after the final "::"; otherwise it returns the
// original urn.
func extractResourceID(urn string) string {
	if urn == "" {
		return ""
	}

	parts := strings.Split(urn, "::")
	if len(parts) >= minURNParts {
		return parts[len(parts)-1] // Last part is the resource name
	}
	return urn
}

// extractProvider extracts the provider name from the resource.
//
// The function uses a two-tier strategy:
//  1. First, try to extract from the provider resource's type field
//     (format: "pulumi:providers:aws" → "aws")
//  2. Fall back to extracting from the resource type prefix
//     (format: "aws:ec2/instance:Instance" → "aws")
//
// extractProvider extracts the provider name for the given AnalyzerResource.
// It first attempts to read the resource's Provider.Type and, if formatted like "pulumi:providers:aws",
// returns the third colon-separated segment ("aws").
// If that fails, it falls back to the resource type prefix (the substring before the first ":" in the resource type,
// e.g. "aws" from "aws:ec2/instance:Instance").
// If neither method yields a provider, it returns "unknown".
// r is the AnalyzerResource to inspect.
func extractProvider(r *pulumirpc.AnalyzerResource) string {
	// Try provider resource first
	if p := r.GetProvider(); p != nil {
		if providerType := p.GetType(); providerType != "" {
			// Format: pulumi:providers:aws
			parts := strings.Split(providerType, ":")
			if len(parts) >= minProviderTypeParts {
				return parts[2]
			}
		}
	}

	// Fall back to resource type prefix
	// Format: aws:ec2/instance:Instance
	resourceType := r.GetType()
	if resourceType == "" {
		return "unknown"
	}

	parts := strings.Split(resourceType, ":")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}

	return "unknown"
}

// structToMap converts a protobuf Struct to a Go map.
//
// This function uses the standard protobuf AsMap() conversion which handles
// all protobuf Value types:
//   - NullValue → nil
//   - BoolValue → bool
//   - NumberValue → float64
//   - StringValue → string
//   - ListValue → []interface{}
//   - Struct → map[string]interface{}
//
// structToMap converts a protobuf Struct into a native Go map[string]interface{}.
// If s is nil, it returns an empty map.
// The returned map contains native Go representations of protobuf values (nil as nil, booleans, numbers, strings, lists as []interface{}, and nested structs as map[string]interface{}).
func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return make(map[string]interface{})
	}
	return s.AsMap()
}