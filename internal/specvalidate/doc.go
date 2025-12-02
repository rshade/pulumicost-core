// Package specvalidate provides validation for pricing specifications.
//
// This package validates YAML pricing specifications to ensure they
// conform to the expected schema and contain valid pricing data.
//
// # Validation Checks
//
// The validator ensures:
//   - Required fields are present (resource_type, pricing)
//   - Pricing values are valid numbers
//   - Currency codes are recognized
//   - Resource types follow expected patterns
package specvalidate
