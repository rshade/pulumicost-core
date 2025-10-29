---
layout: default
title: User Guide
description: Complete guide for end users - install, configure, and use PulumiCost
---

# PulumiCost User Guide

This guide is for anyone who wants to **use PulumiCost** to see costs for their Pulumi infrastructure.

## Table of Contents

1. [What is PulumiCost?](#what-is-pulumicost)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Cost Types](#cost-types)
5. [Common Workflows](#common-workflows)
6. [Configuration](#configuration)
7. [Output Formats](#output-formats)
8. [Filtering and Grouping](#filtering-and-grouping)
9. [Troubleshooting](#troubleshooting)

---

## What is PulumiCost?

PulumiCost is a command-line tool that calculates cloud infrastructure costs from your Pulumi infrastructure definitions.

**Key Features:**
- 📊 **Projected Costs** - Estimate costs before deploying
- 💰 **Actual Costs** - See what you're actually paying
- 🔌 **Multiple Cost Sources** - Works with Vantage, local specs, and more
- 🎯 **Flexible Filtering** - Filter by resource type, tags, or custom criteria
- 📈 **Cost Aggregation** - Group costs by provider, type, date, or tags
- 📱 **Multiple Formats** - Table, JSON, or NDJSON output

---

## Installation

### Prerequisites

- **Pulumi CLI** installed and working
- **Go 1.24+** (if building from source)
- **Cloud credentials** configured (AWS, Azure, GCP, etc.)

### Option 1: Download Binary (Recommended)

Coming soon - prebuilt binaries for Linux, macOS, and Windows.

### Option 2: Build from Source

```bash
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core
make build
./bin/pulumicost --help
```

### Verify Installation

```bash
pulumicost --version
pulumicost --help
```

---

## Quick Start

### 1. Generate Pulumi Plan

```bash
cd your-pulumi-project
pulumi preview --json > plan.json
```

### 2. View Projected Costs

```bash
pulumicost cost projected --pulumi-json plan.json
```

**Output:**
```
RESOURCE                          TYPE                MONTHLY   CURRENCY
aws:ec2/instance:Instance         aws:ec2:Instance    $7.50     USD
aws:s3/bucket:Bucket              aws:s3:Bucket       $0.00     USD
aws:rds/instance:Instance         aws:rds:Instance    $0.00     USD

Total: $7.50 USD
```

### 3. (Optional) View Actual Costs

Requires plugin configuration. See [Configuration](#configuration).

```bash
pulumicost cost actual --pulumi-json plan.json --from 2024-01-01
```

---

## Cost Types

### Projected Costs

**What it is:** Estimated costs based on your infrastructure definitions

**When to use:**
- Before deploying infrastructure
- During planning and design phases
- Comparing different infrastructure options

**Command:**
```bash
pulumicost cost projected --pulumi-json plan.json
```

### Actual Costs

**What it is:** Real costs from your cloud provider's billing system

**When to use:**
- After infrastructure is deployed and running
- Cost optimization and analysis
- Budget tracking and reporting

**Command:**
```bash
pulumicost cost actual --pulumi-json plan.json --from 2024-01-01 --to 2024-01-31
```

**Note:** Requires plugin setup (Vantage, Kubecost, etc.)

---

## Common Workflows

### 1. Check Cost Before Deploying

```bash
# Generate plan
pulumi preview --json > plan.json

# Check projected costs
pulumicost cost projected --pulumi-json plan.json

# Review output and make decisions
```

### 2. Compare Costs of Different Configurations

```bash
# Try one configuration
pulumi preview --json > config1.json
pulumicost cost projected --pulumi-json config1.json

# Switch configuration
# ... modify Pulumi code ...

# Try another configuration
pulumi preview --json > config2.json
pulumicost cost projected --pulumi-json config2.json

# Compare outputs
```

### 3. Track Historical Spending

```bash
# View last 7 days
pulumicost cost actual --from 2024-01-24

# View last month
pulumicost cost actual --from 2024-01-01 --to 2024-01-31

# View by day
pulumicost cost actual --from 2024-01-01 --to 2024-01-31 --group-by daily
```

### 4. Find Expensive Resources

```bash
# Sort by cost (output shows highest first)
pulumicost cost projected --pulumi-json plan.json --output json | jq '.resources | sort_by(.estimatedCost) | reverse'

# Or filter to specific resource type
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:rds*"
```

### 5. Cost by Environment

```bash
# Assuming resources are tagged with 'env' tag
pulumicost cost actual --filter "tag:env=prod" --from 2024-01-01

pulumicost cost actual --filter "tag:env=dev" --from 2024-01-01
```

---

## Configuration

### Using Vantage Plugin

Vantage provides unified cost data from multiple cloud providers.

**Setup:**
1. Get Vantage API key from https://vantage.sh
2. Configure plugin (see [Vantage Plugin Setup](../plugins/vantage/setup.md))
3. Run commands with Vantage data

**Commands:**
```bash
pulumicost cost actual --from 2024-01-01 --to 2024-01-31
```

### Using Local Pricing Specs

Use local YAML files for cost estimates without external services.

**Setup:**
1. Create YAML spec file: `~/.pulumicost/specs/my-specs.yaml`
2. Add resource pricing definitions
3. PulumiCost automatically uses them

**Example spec file:**
```yaml
---
resources:
  aws:ec2/instance:Instance:
    t3.micro:
      monthly: 7.50
      currency: USD
      notes: Linux on-demand
    t3.small:
      monthly: 15.00
      currency: USD
```

---

## Output Formats

### Table (Default)

```bash
pulumicost cost projected --pulumi-json plan.json
```

**Output:**
```
RESOURCE                      TYPE              MONTHLY   CURRENCY
aws:ec2/instance:Instance     aws:ec2:Instance  $7.50     USD
aws:s3/bucket:Bucket          aws:s3:Bucket     $0.00     USD
```

### JSON

```bash
pulumicost cost projected --pulumi-json plan.json --output json
```

**Output:**
```json
{
  "summary": {
    "totalMonthly": 7.50,
    "currency": "USD"
  },
  "resources": [
    {
      "type": "aws:ec2:Instance",
      "estimatedCost": 7.50,
      "currency": "USD"
    }
  ]
}
```

### NDJSON (Newline-Delimited JSON)

Useful for streaming and pipeline processing.

```bash
pulumicost cost projected --pulumi-json plan.json --output ndjson
```

**Output:**
```
{"type": "aws:ec2:Instance", "estimatedCost": 7.50}
{"type": "aws:s3:Bucket", "estimatedCost": 0.00}
```

---

## Filtering and Grouping

### Filtering by Resource Type

```bash
# EC2 instances only
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2*"

# RDS databases
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:rds*"
```

### Filtering by Tags

```bash
# Production resources
pulumicost cost actual --filter "tag:env=prod" --from 2024-01-01

# Team resources
pulumicost cost actual --filter "tag:team=platform" --from 2024-01-01

# Multiple conditions
pulumicost cost actual --filter "tag:env=prod AND tag:team=platform" --from 2024-01-01
```

### Grouping by Dimension

```bash
# By provider (AWS, Azure, GCP)
pulumicost cost actual --group-by provider --from 2024-01-01

# By resource type
pulumicost cost actual --group-by type --from 2024-01-01

# By date (daily breakdown)
pulumicost cost actual --group-by daily --from 2024-01-01 --to 2024-01-31

# By tag
pulumicost cost actual --group-by "tag:env" --from 2024-01-01
```

---

## Troubleshooting

### "No cost data available"

**Problem:** No pricing information found for resources

**Solutions:**
- Check if plugin is configured correctly
- Verify API credentials are valid
- Some resources may not have pricing data - this is normal
- Check troubleshooting guide: [Troubleshooting](../support/troubleshooting.md)

### "Invalid date format"

**Problem:** Date format not recognized

**Solutions:**
- Use format: `YYYY-MM-DD` (e.g., `2024-01-01`)
- Or RFC3339: `2024-01-01T00:00:00Z`
- Example: `--from 2024-01-01 --to 2024-01-31`

### "Plugin not found"

**Problem:** Cost source plugin not installed

**Solutions:**
```bash
# List installed plugins
pulumicost plugin list

# Validate installations
pulumicost plugin validate

# See plugin setup guide for your cost source
# - Vantage: docs/plugins/vantage/setup.md
```

### Getting Help

- **FAQ:** [Frequently Asked Questions](../support/faq.md)
- **Troubleshooting:** [Detailed Troubleshooting Guide](../support/troubleshooting.md)
- **Report Issue:** [GitHub Issues](https://github.com/rshade/pulumicost-core/issues)

---

## Next Steps

- **Quick Start:** [5-Minute Quickstart](../getting-started/quickstart.md)
- **Installation:** [Detailed Installation Guide](../getting-started/installation.md)
- **Vantage Setup:** [Setting up Vantage Plugin](../plugins/vantage/setup.md)
- **CLI Reference:** [Complete CLI Commands](../reference/cli-commands.md)
- **Examples:** [Practical Examples](../getting-started/examples/)

---

**Last Updated:** 2025-10-29
