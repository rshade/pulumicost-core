# Resource Mapping Contract

**Feature**: 009-analyzer-plugin
**Date**: 2025-12-05
**Location**: `internal/analyzer/mapper.go`

## Overview

This document defines the contract for mapping Pulumi's `pulumirpc.AnalyzerResource` to PulumiCost's `engine.ResourceDescriptor`. The mapping must be lossless for cost-relevant fields while handling edge cases gracefully.

## Interface Definition

```go
package analyzer

import (
    "github.com/pulumi/pulumi/sdk/v3/proto/go/pulumirpc"
    "github.com/rshade/pulumicost-core/internal/engine"
)

// ResourceMapper defines the interface for resource mapping.
type ResourceMapper interface {
    // MapResource converts a single AnalyzerResource to ResourceDescriptor.
    MapResource(r *pulumirpc.AnalyzerResource) (engine.ResourceDescriptor, error)

    // MapResources converts a slice of resources.
    MapResources(resources []*pulumirpc.AnalyzerResource) ([]engine.ResourceDescriptor, []MappingError)
}

// MappingError captures a resource that failed to map.
type MappingError struct {
    URN     string // Original URN
    Type    string // Original type
    Error   error  // Mapping error
    Skipped bool   // True if resource was skipped
}
```

## Field Mapping Specification

### Type Field

| Source | Target | Transformation |
|--------|--------|----------------|
| `r.Type` | `ResourceDescriptor.Type` | Direct copy |

**Examples**:

| Input | Output |
|-------|--------|
| `aws:ec2/instance:Instance` | `aws:ec2/instance:Instance` |
| `azure:compute/virtualMachine:VirtualMachine` | `azure:compute/virtualMachine:VirtualMachine` |
| `kubernetes:core/v1:Pod` | `kubernetes:core/v1:Pod` |
| `pulumi:pulumi:Stack` | `pulumi:pulumi:Stack` |

**Validation**:

- Type MUST NOT be empty
- Type length MUST NOT exceed 256 bytes
- Invalid types return `MappingError`

### ID Field

| Source | Target | Transformation |
|--------|--------|----------------|
| `r.Urn` | `ResourceDescriptor.ID` | Extract last `::` segment |

**URN Format**:

```text
urn:pulumi:stack::project::type::name
          │      │        │     │
          └stack └project └type └name (this becomes ID)
```

**Examples**:

| Input URN | Output ID |
|-----------|-----------|
| `urn:pulumi:dev::myapp::aws:ec2/instance:Instance::webserver` | `webserver` |
| `urn:pulumi:prod::api::azure:storage/account:Account::main` | `main` |
| `urn:pulumi:staging::k8s::kubernetes:core/v1:Pod::nginx` | `nginx` |

**Edge Cases**:

| Scenario | Behavior |
|----------|----------|
| Empty URN | Use empty string |
| Malformed URN (no `::`) | Use entire URN |
| URN with trailing `::` | Use empty string |

### Provider Field

| Source | Target | Transformation |
|--------|--------|----------------|
| `r.Provider.Type` | `ResourceDescriptor.Provider` | Extract provider name |
| `r.Type` (fallback) | `ResourceDescriptor.Provider` | Extract first `:` segment |

**Provider URN Format**:

```text
pulumi:providers:aws
                 │
                 └provider name (this becomes Provider)
```

**Examples**:

| Provider Type | Output |
|---------------|--------|
| `pulumi:providers:aws` | `aws` |
| `pulumi:providers:azure` | `azure` |
| `pulumi:providers:gcp` | `gcp` |
| `pulumi:providers:kubernetes` | `kubernetes` |

**Fallback from Resource Type**:

| Resource Type | Extracted Provider |
|---------------|-------------------|
| `aws:ec2/instance:Instance` | `aws` |
| `azure:compute/vm:VM` | `azure` |
| `gcp:compute/instance:Instance` | `gcp` |
| `pulumi:pulumi:Stack` | `pulumi` |

**Edge Cases**:

| Scenario | Behavior |
|----------|----------|
| No provider resource | Use type prefix |
| Empty provider type | Use type prefix |
| Unknown format | Return "unknown" |

### Properties Field

| Source | Target | Transformation |
|--------|--------|----------------|
| `r.Properties` | `ResourceDescriptor.Properties` | `structpb.AsMap()` |

**Type Conversions**:

| Protobuf Type | Go Type |
|---------------|---------|
| `NullValue` | `nil` |
| `BoolValue` | `bool` |
| `NumberValue` | `float64` |
| `StringValue` | `string` |
| `ListValue` | `[]interface{}` |
| `Struct` | `map[string]interface{}` |

**Cost-Relevant Properties** (extracted for SKU/region):

| Provider | Property Keys |
|----------|---------------|
| AWS EC2 | `instanceType`, `ami`, `availabilityZone` |
| AWS RDS | `instanceClass`, `engine`, `allocatedStorage` |
| AWS S3 | `region`, `acl` |
| Azure VM | `vmSize`, `location` |
| Azure Storage | `accountReplicationType`, `accountTier` |
| GCP Compute | `machineType`, `zone` |
| Kubernetes | `resources.requests`, `resources.limits` |

**Property Validation**:

- Max 100 properties per resource
- Key max 128 bytes
- Value max 10KB (serialized)
- Invalid properties logged as warning, not fatal

## Error Handling

### Mapping Errors

```go
type MappingError struct {
    URN     string
    Type    string
    Error   error
    Skipped bool
}
```

**Error Categories**:

| Error | Cause | Behavior |
|-------|-------|----------|
| `ErrEmptyType` | Empty resource type | Skip resource |
| `ErrTypeTooLong` | Type exceeds 256 bytes | Skip resource |
| `ErrPropertyOverflow` | Too many properties | Truncate properties |
| `ErrPropertyInvalid` | Invalid property format | Skip property |
| `ErrURNMalformed` | Invalid URN format | Use URN as ID |

### Graceful Degradation

The mapper continues processing even when individual resources fail:

```go
func MapResources(resources []*pulumirpc.AnalyzerResource) ([]engine.ResourceDescriptor, []MappingError) {
    var results []engine.ResourceDescriptor
    var errors []MappingError

    for _, r := range resources {
        desc, err := MapResource(r)
        if err != nil {
            errors = append(errors, MappingError{
                URN:     r.GetUrn(),
                Type:    r.GetType(),
                Error:   err,
                Skipped: true,
            })
            continue
        }
        results = append(results, desc)
    }

    return results, errors
}
```

## Test Cases

### Happy Path

| Test | Input | Expected Output |
|------|-------|-----------------|
| AWS EC2 | `type: aws:ec2/instance:Instance`, `urn: ...::webserver` | `Type: aws:ec2...`, `ID: webserver`, `Provider: aws` |
| Azure VM | `type: azure:compute/vm:VirtualMachine`, provider set | `Provider: azure` |
| With properties | `properties: {instanceType: t3.micro}` | `Properties: map[instanceType:t3.micro]` |

### Edge Cases

| Test | Input | Expected |
|------|-------|----------|
| Empty URN | `urn: ""` | `ID: ""` |
| No provider | `provider: nil` | Provider from type prefix |
| Nil properties | `properties: nil` | `Properties: map{}` |
| Pulumi Stack | `type: pulumi:pulumi:Stack` | `Provider: pulumi` |

### Error Cases

| Test | Input | Expected |
|------|-------|----------|
| Empty type | `type: ""` | `MappingError{Skipped: true}` |
| Long type | `type: (257 chars)` | `MappingError{Skipped: true}` |
| Too many props | 101 properties | Truncated, warning logged |

## Performance Requirements

| Metric | Target |
|--------|--------|
| Single resource mapping | < 1ms |
| 100 resources | < 50ms |
| 1000 resources | < 500ms |
| Memory per resource | < 1KB overhead |
