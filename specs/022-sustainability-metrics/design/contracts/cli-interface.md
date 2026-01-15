# CLI Interface Contract

**Feature**: Sustainability Metrics

## Inputs

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--utilization` | `float64` | `0.5` (or plugin default) | Assumed utilization rate (0.0 - 1.0) for estimation. |

### Usage

```bash
finfocus --utilization 0.75
```

## Outputs

### JSON Output (`--json`)

The `impactMetrics` field is added to the resource object.

```json
[
  {
    "urn": "urn:pulumi:dev::stack::aws:ec2/instance:Instance::my-server",
    "costTotal": 100.0,
    "currency": "USD",
    "impactMetrics": [
      {
        "kind": "METRIC_KIND_CARBON_FOOTPRINT",
        "value": 5000.0,
        "unit": "gCO2e"
      }
    ]
  }
]
```

### Table Output (TUI)

Columns are dynamic.

| Resource | Cost | ... | CO₂ |
|----------|------|-----|-----|
| my-server| $100 | ... | 5 kg|

If no metrics are present for ANY resource in the table, the `CO₂` column is hidden.
