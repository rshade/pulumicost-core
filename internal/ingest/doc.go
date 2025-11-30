// Package ingest handles parsing of Pulumi infrastructure definitions.
//
// The primary function is converting "pulumi preview --json" output into
// resource descriptors that can be processed by the cost calculation engine.
//
// # Pulumi Plan Parsing
//
// The package extracts from Pulumi preview JSON:
//   - Resource types (e.g., "aws:ec2:Instance")
//   - Provider information
//   - Resource properties and configurations
//   - Resource dependencies and relationships
//
// # Resource Descriptors
//
// Output is a normalized set of ResourceDescriptor objects that provide
// a provider-agnostic view of infrastructure resources for cost analysis.
package ingest
