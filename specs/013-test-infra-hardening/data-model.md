# Data Model - Test Infrastructure Hardening

## 1. Overview

This data model defines the structures required for the new test infrastructure, specifically for performance benchmarking and fuzz testing. No changes to the core application data model are required.

## 2. Performance Benchmark Entities

These entities represent the synthetic infrastructure plans used for stress testing.

### 2.1. BenchmarkConfig

Configuration for the synthetic data generator.

| Field | Type | Description |
|-------|------|-------------|
| `ResourceCount` | `int` | Total number of resources to generate. |
| `MaxDepth` | `int` | Maximum nesting level for child resources/properties. |
| `DependencyRatio` | `float64` | Probability (0.0-1.0) of a resource having a dependency. |
| `Seed` | `int64` | Random seed for deterministic generation. |

### 2.2. SyntheticResource

A simplified representation of a generic infrastructure resource for testing.

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `string` | Resource type (e.g., "aws:ec2:Instance"). |
| `Name` | `string` | Unique resource name. |
| `Properties` | `map[string]interface{}` | Arbitrary properties, potentially nested. |
| `DependsOn` | `[]string` | List of resource names this resource depends on. |

### 2.3. SyntheticPlan

The top-level container for the generated dataset.

| Field | Type | Description |
|-------|------|-------------|
| `Resources` | `[]SyntheticResource` | List of all generated resources. |
| `Variables` | `map[string]interface{}` | Global variables (optional). |

## 3. Fuzzing Entities

These entities define the input structure for fuzz targets.

### 3.1. FuzzInput

The byte slice input provided by the Go fuzzing engine.

| Field | Type | Description |
|-------|------|-------------|
| `data` | `[]byte` | Raw byte stream to be fed into parsers. |

## 4. Validation Rules

-   **ResourceCount**: Must be > 0.
-   **MaxDepth**: Must be >= 0.
-   **DependencyRatio**: Must be between 0.0 and 1.0.
-   **Seed**: Any valid int64.

## 5. Benchmark Configuration Presets

These presets define standard configurations for benchmark scenarios:

| Preset | ResourceCount | MaxDepth | DependencyRatio | Use Case |
|--------|---------------|----------|-----------------|----------|
| Small | 1,000 | 3 | 0.2 | Quick regression tests |
| Medium | 10,000 | 5 | 0.3 | Standard benchmarks |
| Large (Stress) | 100,000 | 5 | 0.3 | Stress testing (<5 min target) |
| Deep Nesting | 1,000 | 10 | 0.5 | Depth complexity testing |

**"Moderately complex"** refers to the **Medium** preset: `MaxDepth=5, DependencyRatio=0.3`. This simulates realistic enterprise infrastructure with:

- 5 levels of resource nesting (e.g., VPC → Subnet → Security Group → Instance → Volume)
- 30% of resources having at least one dependency

## 6. JSON Schema (Example Output)

```json
{
  "resources": [
    {
      "type": "aws:s3:Bucket",
      "name": "bucket-0",
      "properties": {
        "acl": "private",
        "tags": {
          "Environment": "test"
        }
      },
      "dependsOn": []
    },
    {
      "type": "aws:s3:BucketObject",
      "name": "object-0",
      "properties": {
        "source": "file.txt"
      },
      "dependsOn": ["bucket-0"]
    }
  ]
}
```
