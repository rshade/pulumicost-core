# Data Model: Analyzer Recommendations Display

**Feature**: 106-analyzer-recommendations
**Date**: 2025-12-25

## Entities

### Recommendation (NEW)

Represents a single cost optimization suggestion to display in analyzer
diagnostics.

| Field            | Type    | Required | Description                         |
| ---------------- | ------- | -------- | ----------------------------------- |
| Type             | string  | Yes      | Category (e.g., "Right-sizing")     |
| Description      | string  | Yes      | Actionable text                     |
| EstimatedSavings | float64 | No       | Monthly savings estimate (0 if N/A) |
| Currency         | string  | No       | ISO 4217 code (empty if no savings) |

**Validation Rules**:

- Type must not be empty
- Description must not be empty
- EstimatedSavings must be >= 0
- Currency should be valid ISO 4217 if EstimatedSavings > 0

**JSON Representation**:

```json
{
  "type": "Right-sizing",
  "description": "Switch to t3.small",
  "estimatedSavings": 15.00,
  "currency": "USD"
}
```

### CostResult (MODIFIED)

Extended to include optional recommendations.

| Field           | Type             | Req | Description               |
| --------------- | ---------------- | --- | ------------------------- |
| ...             | ...              | ... | (existing fields)         |
| Recommendations | []Recommendation | No  | Optimization suggestions  |

**JSON Representation (with recommendations)**:

```json
{
  "resourceType": "aws:ec2/instance:Instance",
  "resourceId": "webserver",
  "adapter": "aws-plugin",
  "currency": "USD",
  "monthly": 25.50,
  "hourly": 0.035,
  "notes": "",
  "breakdown": {},
  "sustainability": {},
  "recommendations": [
    {
      "type": "Right-sizing",
      "description": "Switch to t3.small",
      "estimatedSavings": 15.00,
      "currency": "USD"
    }
  ]
}
```

## Relationships

```text
CostResult 1---* Recommendation
    │
    │ (existing relationships unchanged)
    │
    └───* SustainabilityMetric
```

## State Transitions

N/A - This feature is stateless (display-only).

## Data Volume Assumptions

- Recommendations per resource: Typically 0-5, max ~10
- Total recommendations per stack: Typically 0-20, max ~100
- Message length impact: ~50-100 chars per recommendation
