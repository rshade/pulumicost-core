# PulumiCost User Guide

Complete guide to using PulumiCost Core for cloud infrastructure cost analysis.

## Table of Contents

- [Overview](#overview)
- [Basic Workflow](#basic-workflow)
- [Projected Cost Analysis](#projected-cost-analysis)
- [Actual Cost Analysis](#actual-cost-analysis)
- [Advanced Features](#advanced-features)
- [Output Formats](#output-formats)
- [Configuration](#configuration)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

PulumiCost Core analyzes your Pulumi infrastructure definitions to provide:

- **Projected Costs**: Estimates based on resource specifications
- **Actual Costs**: Historical spending data from cloud providers
- **Cost Analytics**: Detailed breakdowns and trend analysis

The tool works by parsing Pulumi plan JSON output and querying cost data through plugins or local pricing specifications.

## Basic Workflow

### 1. Generate Pulumi Plan

First, export your infrastructure plan to JSON format:

```bash
cd your-pulumi-project

# For new resources (preview)
pulumi preview --json > plan.json

# For existing resources (current state)
pulumi stack export > current-state.json
```

### 2. Analyze Costs

Use the generated JSON file with PulumiCost commands:

```bash
# Quick projected cost estimate
pulumicost cost projected --pulumi-json plan.json

# Historical cost analysis
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01
```

## Projected Cost Analysis

Projected costs estimate monthly spending based on resource specifications and pricing data.

### Basic Usage

```bash
# Default table output
pulumicost cost projected --pulumi-json plan.json

# Specify output format
pulumicost cost projected --pulumi-json plan.json --output json
```

### Resource Filtering

Filter resources to focus on specific components:

```bash
# Filter by resource type
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2/instance"

# Filter by multiple criteria (comma-separated)
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2/instance,aws:rds/instance"
```

### Using Custom Specs

Override default pricing with custom specifications:

```bash
# Use custom spec directory
pulumicost cost projected --pulumi-json plan.json --spec-dir ./custom-pricing

# Use specific plugin only
pulumicost cost projected --pulumi-json plan.json --adapter aws-pricing-plugin
```

### Sample Output

```
RESOURCE                          ADAPTER     MONTHLY   CURRENCY  NOTES
aws:ec2/instance:Instance         aws-spec    $7.50     USD       t3.micro Linux on-demand
aws:s3/bucket:Bucket             aws-spec    $2.30     USD       Standard storage 100GB
aws:rds/instance:Instance        aws-spec    $15.70    USD       db.t3.micro PostgreSQL
                                                        -------
                                  TOTAL       $25.50    USD
```

## Actual Cost Analysis

Actual costs retrieve historical spending data from cloud provider APIs through plugins.

### Basic Usage

```bash
# Last 7 days (to defaults to now)
pulumicost cost actual --pulumi-json plan.json --from 2025-01-07

# Specific date range
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-01-31

# Use specific plugin
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --adapter kubecost
```

### Date Format Support

PulumiCost supports multiple date formats:

```bash
# Simple date format (YYYY-MM-DD)
--from 2025-01-01 --to 2025-01-31

# RFC3339 timestamps
--from 2025-01-01T00:00:00Z --to 2025-01-31T23:59:59Z

# Partial timestamps (time defaults to start/end of day)
--from 2025-01-01T14:30:00Z
```

### Cost Aggregation and Grouping

Group costs by different dimensions:

```bash
# Group by resource
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by resource

# Group by resource type
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by type

# Group by cloud provider
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by provider

# Group by date (time series)
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by date
```

### Tag-Based Filtering

Filter costs by resource tags:

```bash
# Filter by specific tag value
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by "tag:Environment=prod"

# Filter by team
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by "tag:Team=backend"
```

### Sample Output

```json
{
  "summary": {
    "totalMonthly": 156.78,
    "currency": "USD",
    "byProvider": {
      "aws": 156.78
    },
    "byService": {
      "ec2": 89.45,
      "s3": 23.12,
      "rds": 44.21
    },
    "byAdapter": {
      "kubecost": 156.78
    }
  },
  "resources": [
    {
      "resourceType": "aws:ec2/instance:Instance",
      "resourceId": "web-server",
      "adapter": "kubecost",
      "currency": "USD",
      "totalCost": 89.45,
      "dailyCosts": [2.88, 2.92, 2.85, ...],
      "costPeriod": "2025-01-01 to 2025-01-31",
      "startDate": "2025-01-01T00:00:00Z",
      "endDate": "2025-01-31T23:59:59Z"
    }
  ]
}
```

## Advanced Features

### Multiple Time Ranges

Compare costs across different periods:

```bash
# Q4 2024 vs Q1 2025
pulumicost cost actual --pulumi-json plan.json --from 2024-10-01 --to 2024-12-31 > q4-2024.json
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-03-31 > q1-2025.json
```

### Complex Filtering

Combine multiple filter types:

```bash
# EC2 instances in production environment
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2/instance" | \
  jq '.resources[] | select(.notes | contains("prod"))'

# Resources with specific tags and types
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by "tag:Environment=prod" | \
  jq '.resources[] | select(.resourceType | contains("ec2"))'
```

### Batch Processing

Process multiple Pulumi projects:

```bash
#!/bin/bash
for project in project-a project-b project-c; do
  echo "Analyzing $project..."
  pulumi -C $project preview --json > $project-plan.json
  pulumicost cost projected --pulumi-json $project-plan.json --output json > $project-costs.json
done
```

## Output Formats

### Table Format (Default)

Human-readable output with aligned columns:

```bash
pulumicost cost projected --pulumi-json plan.json --output table
```

Best for: Interactive use, quick overviews

### JSON Format

Structured data for programmatic use:

```bash
pulumicost cost projected --pulumi-json plan.json --output json
```

Best for: API integration, detailed analysis, reporting

### NDJSON Format

Newline-delimited JSON for streaming:

```bash
pulumicost cost projected --pulumi-json plan.json --output ndjson
```

Best for: Pipeline processing, log analysis, time series data

## Configuration

### Directory Structure

PulumiCost uses the following directory structure:

```
~/.pulumicost/
├── plugins/                     # Plugin binaries
│   ├── kubecost/
│   │   └── 1.0.0/
│   │       └── pulumicost-kubecost
│   └── aws-plugin/
│       └── 0.1.0/
│           └── pulumicost-aws
└── specs/                       # Local pricing specs
    ├── aws-ec2-t3-micro.yaml
    ├── aws-s3-standard.yaml
    └── azure-vm-b2s.yaml
```

### Local Pricing Specs

Create YAML files for custom pricing:

```yaml
# ~/.pulumicost/specs/aws-ec2-t3-micro.yaml
provider: aws
service: ec2
sku: t3.micro
currency: USD
pricing:
  instanceType: t3.micro
  onDemandHourly: 0.0104
  monthlyEstimate: 7.59
  vcpu: 2
  memory: 1
metadata:
  region: us-west-2
  operatingSystem: linux
  tenancy: shared
```

### Plugin Configuration

Plugins may require additional configuration:

```bash
# Set plugin-specific environment variables
export KUBECOST_API_URL="http://kubecost.example.com:9090"
export AWS_PRICING_API_KEY="your-api-key"

# Run with plugin configuration
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --adapter kubecost
```

## Best Practices

### Cost Analysis Workflow

1. **Start with Projected Costs**: Estimate costs before deployment
2. **Monitor Actual Costs**: Track spending after deployment
3. **Compare and Optimize**: Identify discrepancies and optimization opportunities
4. **Regular Reviews**: Set up periodic cost analysis

### Resource Tagging Strategy

Implement consistent tagging for better cost attribution:

```typescript
// Pulumi TypeScript example
const webServer = new aws.ec2.Instance("web-server", {
    // ... other properties
    tags: {
        Environment: "production",
        Team: "backend",
        Project: "web-app",
        CostCenter: "engineering"
    }
});
```

### Automation

Integrate PulumiCost into your CI/CD pipeline:

```yaml
# GitHub Actions example
- name: Cost Analysis
  run: |
    pulumi preview --json > plan.json
    pulumicost cost projected --pulumi-json plan.json --output json > projected-costs.json
    
    # Set cost threshold
    if [ $(jq '.summary.totalMonthly' projected-costs.json) -gt 1000 ]; then
      echo "Cost threshold exceeded!"
      exit 1
    fi
```

### Performance Tips

- Use `--adapter` to restrict to specific plugins for faster queries
- Filter resources early with `--filter` to reduce processing time  
- Use NDJSON output for large datasets
- Cache plugin responses when possible

## Examples

### Cost Optimization Analysis

Compare different instance types:

```bash
# Current plan with t3.micro
pulumicost cost projected --pulumi-json current-plan.json > current-costs.txt

# Modified plan with t3.small  
sed 's/t3.micro/t3.small/g' current-plan.json > optimized-plan.json
pulumicost cost projected --pulumi-json optimized-plan.json > optimized-costs.txt

# Compare results
diff current-costs.txt optimized-costs.txt
```

### Multi-Environment Cost Tracking

Track costs across environments:

```bash
#!/bin/bash
for env in dev staging prod; do
  echo "=== $env Environment ==="
  pulumicost cost actual \
    --pulumi-json plans/$env-plan.json \
    --from 2025-01-01 \
    --group-by "tag:Environment=$env" \
    --output json > costs/$env-costs.json
done
```

### Cost Attribution Report

Generate detailed cost attribution:

```bash
# By team
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by "tag:Team=backend" --output json | \
  jq '.summary.byService'

# By project  
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by "tag:Project=web-app" --output json | \
  jq '.resources[] | {type: .resourceType, cost: .totalCost}'
```

### Time Series Analysis

Track cost trends over time:

```bash
# Generate daily cost data
pulumicost cost actual \
  --pulumi-json plan.json \
  --from 2025-01-01 \
  --to 2025-01-31 \
  --group-by date \
  --output ndjson | \
  jq -r '[.startDate, .totalCost] | @csv' > daily-costs.csv
```

## Next Steps

- [Installation Guide](installation.md) - Detailed setup instructions
- [Cost Calculations](cost-calculations.md) - Deep dive into cost methodologies  
- [Plugin System](plugin-system.md) - Plugin development and management
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

For more examples, see the [examples](../examples/) directory in the repository.