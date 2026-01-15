// Package spec handles local pricing specifications.
//
// When plugins don't provide pricing for a resource type, the engine
// falls back to local YAML-based pricing specifications.
//
// # Specification Location
//
// Specs are stored in ~/.finfocus/specs/ as YAML files.
//
// # Specification Format
//
// Example pricing spec (aws-ec2-t3-micro.yaml):
//
//	resource_type: aws:ec2:Instance
//	instance_type: t3.micro
//	pricing:
//	  hourly: 0.0104
//	  currency: USD
//
// # Usage
//
// Specs provide fallback pricing when:
//   - No plugin is available for a resource type
//   - Plugin returns no pricing data
//   - Plugin query fails
package spec
