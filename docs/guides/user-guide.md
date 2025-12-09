---
layout: default
title: User Guide
description: Complete guide for end users - install, configure, and use PulumiCost
---

This guide is for anyone who wants to **use PulumiCost** to see costs for their Pulumi
infrastructure.

## Table of Contents

1. [What is PulumiCost?](#what-is-pulumicost)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Cost Types](#cost-types)
5. [Common Workflows](#common-workflows)
6. [Configuration](#configuration)
7. [Output Formats](#output-formats)
8. [Filtering and Grouping](#filtering-and-grouping)
9. [Debugging and Logging](#debugging-and-logging)
10. [Logging Configuration](#logging-configuration)
11. [Troubleshooting](#troubleshooting)

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
- **Go 1.25.5+** (if building from source)
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

```text
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

## Zero-Click Cost Estimation (Analyzer)

PulumiCost can integrate directly with the Pulumi CLI as an Analyzer, providing instant
cost estimates during `pulumi preview`. This eliminates the need for a separate
`pulumicost` command to see projected costs.

For detailed setup instructions, refer to the [Analyzer Setup Guide](../getting-started/analyzer-setup.md).

---

## Cross-Provider Aggregation

PulumiCost supports aggregating costs across multiple cloud providers and services,
allowing you to get a holistic view of your infrastructure spending. This feature is
particularly powerful when combining actual cost data from various plugins.

### Daily Cost Trends

View daily cost trends across all configured providers for a specific period:

```bash
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-01-31 --group-by daily
```

### Monthly Comparison

Generate a monthly cost comparison. You can output this as JSON for further processing:

```bash
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by monthly --output json
```

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

```text
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
    "totalMonthly": 7.5,
    "currency": "USD"
  },
  "resources": [
    {
      "type": "aws:ec2:Instance",
      "estimatedCost": 7.5,
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

```text
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

## Debugging and Logging

### Using Debug Mode

PulumiCost includes a `--debug` flag that enables verbose logging to help troubleshoot issues:

```bash
# Enable debug output for any command
pulumicost cost projected --debug --pulumi-json plan.json

# Debug output shows:
# - Command start/stop with duration
# - Resource ingestion details
# - Plugin lookup attempts
# - Cost calculation decisions
# - Fallback behavior when plugins don't return data
```

**Example Debug Output:**

```text
2025-01-15T10:30:45Z INF command started command=projected trace_id=01HQ7X2J3K4M5N6P7Q8R9S0T1U component=cli
2025-01-15T10:30:45Z DBG loading Pulumi plan plan_path=plan.json component=ingest
2025-01-15T10:30:45Z DBG extracted 3 resources from plan component=ingest
2025-01-15T10:30:45Z DBG querying plugin for projected cost resource_type=aws:ec2:Instance plugin=vantage component=engine
2025-01-15T10:30:46Z DBG plugin returned cost data monthly_cost=7.50 component=engine
2025-01-15T10:30:46Z INF projected cost calculation complete result_count=3 duration_ms=245 component=engine
```

### Environment Variables for Logging

Configure logging behavior via environment variables:

```bash
# Set log level (trace, debug, info, warn, error)
export PULUMICOST_LOG_LEVEL=debug

# Set log format (json, text, console)
export PULUMICOST_LOG_FORMAT=json

# Inject external trace ID for correlation with other systems
export PULUMICOST_TRACE_ID=my-pipeline-trace-12345

# Example: Debug with JSON format for log aggregation
PULUMICOST_LOG_LEVEL=debug PULUMICOST_LOG_FORMAT=json \
  pulumicost cost projected --pulumi-json plan.json 2> debug.log
```

### Configuration Precedence

Log settings are applied in this order (highest priority first):

1. **CLI flags** (`--debug`)
2. **Environment variables** (`PULUMICOST_LOG_LEVEL`)
3. **Config file** (`~/.pulumicost/config.yaml`)
4. **Defaults** (info level, text format)

### Trace ID for Debugging

Every command generates a unique trace ID that appears in all log entries.
This helps correlate log entries for a single operation:

```bash
# Use external trace ID for pipeline correlation
PULUMICOST_TRACE_ID=jenkins-build-123 pulumicost cost projected --debug --pulumi-json plan.json

# All logs will include: trace_id=jenkins-build-123
```

---

## Logging Configuration

PulumiCost provides comprehensive logging capabilities for debugging, monitoring, and auditing.

### Configuration File

Create or edit `~/.pulumicost/config.yaml` to configure logging:

```yaml
logging:
  # Log level: trace, debug, info, warn, error (default: info)
  level: info

  # Log format: json, text, console (default: console)
  format: json

  # Log to file (optional - defaults to stderr)
  file: /var/log/pulumicost/pulumicost.log

  # Audit logging for compliance (optional)
  audit:
    enabled: true
    file: /var/log/pulumicost/audit.log
```

### Log Output Locations

**Default Behavior:**

- Without configuration: logs go to stderr in console format
- With `--debug` flag: forces debug level, console format, and stderr output

**File Logging:**

When file logging is configured, PulumiCost displays the log location at startup:

```bash
$ pulumicost cost projected --pulumi-json plan.json
Logging to: /var/log/pulumicost/pulumicost.log
COST SUMMARY
============
...
```

**Fallback Behavior:**

If the configured log file cannot be written (permissions, disk full), PulumiCost:

1. Falls back to stderr
2. Displays a warning with the reason

```bash
$ pulumicost cost projected --pulumi-json plan.json
Warning: Could not write to log file, falling back to stderr (permission denied)
COST SUMMARY
============
...
```

### Audit Logging

Audit logging tracks all cost queries for compliance and analysis.

**Enable Audit Logging:**

```yaml
logging:
  audit:
    enabled: true
    file: /var/log/pulumicost/audit.log
```

**Audit Log Entry Example:**

```json
{
  "time": "2025-01-15T10:30:45Z",
  "level": "info",
  "audit": true,
  "command": "cost projected",
  "trace_id": "01HQ7X2J3K4M5N6P7Q8R9S0T1U",
  "duration_ms": 245,
  "success": true,
  "result_count": 3,
  "total_cost": 7.5,
  "parameters": {
    "pulumi_json": "plan.json",
    "output": "table"
  }
}
```

**Audit Entry Fields:**

| Field          | Description                                                  |
| -------------- | ------------------------------------------------------------ |
| `command`      | CLI command executed (e.g., "cost projected", "cost actual") |
| `trace_id`     | Unique request identifier for correlation                    |
| `duration_ms`  | Command execution time in milliseconds                       |
| `success`      | Whether the command completed successfully                   |
| `result_count` | Number of resources processed                                |
| `total_cost`   | Sum of all calculated costs                                  |
| `parameters`   | Command parameters (sensitive values redacted)               |

**Security:** Sensitive parameter values (API keys, passwords, tokens) are automatically redacted in audit logs.

### Log Rotation

PulumiCost does not perform log rotation internally. Use external tools:

**Linux (logrotate):**

```text
# /etc/logrotate.d/pulumicost
/var/log/pulumicost/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

**systemd journald:**

If running as a service, logs go to journald automatically:

```bash
journalctl -u pulumicost --since today
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
