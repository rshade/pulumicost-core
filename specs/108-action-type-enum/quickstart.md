# Quickstart: Extended RecommendationActionType Enum Support

**Feature**: 108-action-type-enum
**Date**: 2025-12-29

## Overview

This feature adds the `cost recommendations` CLI command and full support for
all 11 `RecommendationActionType` enum values in the CLI filter parser, TUI
display, and JSON output.

## Usage Examples

### Basic Recommendations Command

```bash
# Get all cost optimization recommendations for resources in a Pulumi plan
pulumicost cost recommendations --pulumi-json plan.json

# Output recommendations as JSON
pulumicost cost recommendations --pulumi-json plan.json --output json

# Use a specific adapter/plugin
pulumicost cost recommendations --pulumi-json plan.json --adapter kubecost
```

### Filtering by Action Type

```bash
# Filter recommendations by a single action type
pulumicost cost recommendations --pulumi-json plan.json --filter "action=MIGRATE"

# Filter by multiple action types (comma-separated)
pulumicost cost recommendations --pulumi-json plan.json --filter "action=RIGHTSIZE,TERMINATE"

# Case-insensitive matching works
pulumicost cost recommendations --pulumi-json plan.json --filter "action=migrate,consolidate"
```

### Available Action Types for Filtering

<!-- markdownlint-disable MD060 -->

| Type | Description |
| ---- | ----------- |
| RIGHTSIZE | Resize resources to match actual usage |
| TERMINATE | Stop or delete idle resources |
| PURCHASE_COMMITMENT | Reserved instances/savings plans |
| ADJUST_REQUESTS | Kubernetes request/limit tuning |
| MODIFY | General configuration changes |
| DELETE_UNUSED | Remove orphaned/unused resources |
| MIGRATE | Move workloads to different regions/zones/SKUs |
| CONSOLIDATE | Combine multiple resources into fewer, larger ones |
| SCHEDULE | Start/stop resources on schedule (dev/test) |
| REFACTOR | Architectural changes (e.g., move to serverless) |
| OTHER | Provider-specific recommendations |

<!-- markdownlint-enable MD060 -->

### TUI Display

When viewing recommendations in TUI mode, each action type displays with a
human-readable label:

- "Migrate" instead of "MIGRATE"
- "Purchase Commitment" instead of "PURCHASE_COMMITMENT"
- "Adjust Requests" instead of "ADJUST_REQUESTS"

### JSON Output

JSON output preserves the canonical enum string names:

```json
{
  "recommendations": [
    {
      "action_type": "MIGRATE",
      "description": "Move EC2 instance to us-west-2 region",
      "estimated_savings": 45.00
    },
    {
      "action_type": "SCHEDULE",
      "description": "Stop dev instance outside business hours",
      "estimated_savings": 120.00
    }
  ]
}
```

### Error Handling

Invalid action type filter values produce helpful error messages:

```bash
$ pulumicost cost recommendations --pulumi-json plan.json --filter "action=INVALID"
Error: invalid action type "INVALID". Valid types: RIGHTSIZE, TERMINATE,
PURCHASE_COMMITMENT, ADJUST_REQUESTS, MODIFY, DELETE_UNUSED, MIGRATE,
CONSOLIDATE, SCHEDULE, REFACTOR, OTHER
```

## Developer Integration

### Using Action Type Utilities

```go
import "github.com/rshade/pulumicost-core/internal/proto"

// Get display label for TUI
label := proto.ActionTypeLabel(pbc.RecommendationActionType_RECOMMENDATION_ACTION_TYPE_MIGRATE)
// Returns: "Migrate"

// Parse filter string
actionType, err := proto.ParseActionType("migrate")
// Returns: pbc.RecommendationActionType_RECOMMENDATION_ACTION_TYPE_MIGRATE, nil

// Parse multiple action types from filter
types, err := proto.ParseActionTypeFilter("migrate,consolidate,schedule")
// Returns: []pbc.RecommendationActionType{MIGRATE, CONSOLIDATE, SCHEDULE}, nil

// Get valid types for help text
validTypes := proto.ValidActionTypes()
// Returns: []string{"RIGHTSIZE", "TERMINATE", ..., "OTHER"}
```

## Backward Compatibility

- Existing filters using original 6 action types continue to work
- No changes to JSON output format
- Proto enum values maintain their integer mappings
