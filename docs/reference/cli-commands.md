---
layout: default
title: CLI Commands Reference
description: Complete reference for all PulumiCost CLI commands
---

Complete command reference for PulumiCost.

## Commands Overview

```bash
pulumicost                 # Help
pulumicost cost            # Cost commands
pulumicost cost projected  # Estimate costs from plan
pulumicost cost actual     # Get actual historical costs
pulumicost plugin          # Plugin commands
pulumicost plugin init     # Initialize a new plugin
pulumicost plugin install  # Install a plugin
pulumicost plugin update   # Update a plugin
pulumicost plugin remove   # Remove a plugin
pulumicost plugin list     # List installed plugins
pulumicost plugin validate # Validate plugin setup
pulumicost analyzer        # Analyzer commands
pulumicost analyzer serve  # Start the analyzer gRPC server
```

## cost projected

Calculate estimated costs from Pulumi plan.

### Usage

```bash
pulumicost cost projected --pulumi-json <file> [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--pulumi-json` | Path to Pulumi preview JSON | Required |
| `--filter` | Filter resources (tag:key=value, type=*) | None |
| `--output` | Output format: table, json, ndjson | table |
| `--help` | Show help | |

### Examples

```bash
# Basic usage
pulumicost cost projected --pulumi-json plan.json

# JSON output
pulumicost cost projected --pulumi-json plan.json --output json

# Filter by type
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2*"

# NDJSON for pipelines
pulumicost cost projected --pulumi-json plan.json --output ndjson
```

## cost actual

Get actual historical costs from plugins.

### Usage

```bash
pulumicost cost actual [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--from` | Start date (YYYY-MM-DD or RFC3339) | 7 days ago |
| `--to` | End date (YYYY-MM-DD or RFC3339) | Today |
| `--filter` | Filter resources (tag:key=value, type=*) | None |
| `--group-by` | Group results (resource, type, provider, daily, monthly) | resource |
| `--output` | Output format: table, json, ndjson | table |
| `--help` | Show help | |

### Examples

```bash
# Last 7 days
pulumicost cost actual

# Specific date range
pulumicost cost actual --from 2024-01-01 --to 2024-01-31

# By day
pulumicost cost actual --group-by daily --from 2024-01-01 --to 2024-01-31

# By provider
pulumicost cost actual --group-by provider

# Filter by tag
pulumicost cost actual --filter "tag:env=prod"

# JSON output
pulumicost cost actual --output json --from 2024-01-01
```

## plugin init

Initialize a new PulumiCost plugin project.

### Usage

```bash
pulumicost plugin init <plugin-name> --author <name> --providers <list> [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--author` | Author name for the plugin | (required) |
| `--providers` | Comma-separated list of cloud providers | (required) |
| `--help` | Show help | |

### Examples

```bash
# Initialize a new AWS plugin
pulumicost plugin init my-aws-plugin --author "Your Name" --providers aws
```

## plugin install

Install a PulumiCost plugin from a registry or URL.

### Usage

```bash
pulumicost plugin install <plugin-name> [--version <version>] [--url <url>] [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--version` | Specify plugin version to install | latest |
| `--url` | URL to plugin binary (for custom installs) | (registry lookup) |
| `--force` | Force overwrite existing plugin installation | false |
| `--help` | Show help | |

### Examples

```bash
# Install the latest Vantage plugin
pulumicost plugin install vantage

# Install a specific version of a plugin
pulumicost plugin install kubecost --version 0.2.0

# Install from a custom URL
pulumicost plugin install my-plugin --url https://example.com/my-plugin-0.1.0.tar.gz
```

## plugin update

Update an installed PulumiCost plugin.

### Usage

```bash
pulumicost plugin update <plugin-name> [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--version` | Specify target version (defaults to latest) | latest |
| `--all` | Update all installed plugins | false |
| `--help` | Show help | |

### Examples

```bash
# Update the Vantage plugin to the latest version
pulumicost plugin update vantage

# Update all installed plugins
pulumicost plugin update --all
```

## plugin remove

Remove an installed PulumiCost plugin.

### Usage

```bash
pulumicost plugin remove <plugin-name> [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--all` | Remove all installed plugins | false |
| `--help` | Show help | |

### Examples

```bash
# Remove the Vantage plugin
pulumicost plugin remove vantage

# Remove all installed plugins
pulumicost plugin remove --all
```

## plugin list

List installed plugins.

### Usage

```bash
pulumicost plugin list [options]
```

### Options

| Flag | Description |
| --- | --- |
| `--help` | Show help |

### Examples

```bash
# List all plugins
pulumicost plugin list

# Output:
# NAME      VERSION   STATUS
# vantage   0.1.0     installed
# kubecost  0.2.0     installed
```

## plugin validate

Validate plugin installations.

### Usage

```bash
pulumicost plugin validate [options]
```

### Options

| Flag | Description |
| --- | --- |
| `--help` | Show help |

### Examples

```bash
# Validate all plugins
pulumicost plugin validate

# Output:
# vantage (0.1.0): OK
# kubecost (0.2.0): OK
```

## analyzer serve

Starts the PulumiCost analyzer gRPC server. This command is intended to be run by
the Pulumi CLI as part of the `pulumi preview` workflow, typically configured in
`Pulumi.yaml`.

### Usage

```bash
pulumicost analyzer serve [options]
```

### Options

| Flag | Description | Default |
| --- | --- | --- |
| `--logtostderr` | Log messages to stderr rather than log files | false |
| `--v` | Log level for V-logging (verbose logging) | 0 |
| `--pulumilogfile` | Pulumi log file name (internal use) | (generated) |
| `--help` | Show help | |

### Examples

```bash
# This command is typically not run directly by users.
# It's configured in Pulumi.yaml for zero-click cost estimation:
#
# plugins:
#   - path: pulumicost
#     args: ["analyzer", "serve"]
```

## Global Options

```bash
pulumicost [global options] command [command options]
```

| Option | Description |
| --- | --- |
| `--help` | Show help |
| `--version` | Show version |
| `--debug` | Enable debug logging |

## Date Formats

### Accepted Formats

```bash
# ISO 8601 (YYYY-MM-DD)
pulumicost cost actual --from 2024-01-01

# RFC3339 (full timestamp)
pulumicost cost actual --from 2024-01-01T00:00:00Z

# Relative (future)
pulumicost cost actual --from "7 days ago"
```

## Output Formats

### Table (Default)

Human-readable table format:

```text
RESOURCE    TYPE       MONTHLY   CURRENCY
Instance1   ec2        $7.50     USD
Bucket1     s3         $0.50     USD
──────────────────────────────
Total                  $8.00     USD
```

### JSON

Machine-readable JSON format:

```json
{
  "summary": {"totalMonthly": 8.00, "currency": "USD"},
  "resources": [
    {"name": "Instance1", "type": "ec2", "cost": 7.50}
  ]
}
```

### NDJSON

Newline-delimited JSON (one per line):

```text
{"name":"Instance1","type":"ec2","cost":7.50}
{"name":"Bucket1","type":"s3","cost":0.50}
```

## Exit Codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |

---

See [User Guide](../guides/user-guide.md) for workflow examples.
