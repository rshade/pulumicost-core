# Quickstart: PulumiCost Analyzer Plugin

**Feature**: 009-analyzer-plugin
**Date**: 2025-12-05

## Overview

The PulumiCost Analyzer enables **zero-click cost estimation** directly within `pulumi preview`. After installation, cost estimates appear automatically for every resource in your stack.

```text
Previewing update (dev):
     Type                 Name         Plan       Info
 +   pulumi:pulumi:Stack  mystack      create
 +   └─ aws:ec2:Instance  webserver    create

Diagnostics:
  pulumicost:
    INFO: Estimated Monthly Cost: $8.45 USD (source: local-spec)

  pulumicost:stack-cost-summary:
    INFO: Total Estimated Monthly Cost: $8.45 USD (1 resources analyzed)
```

## Prerequisites

- **Pulumi CLI** v3.0.0 or later
- **Go** 1.24+ (for building from source)
- **Operating System**: Linux, macOS, or Windows

## Installation

### Option 1: Install from Release

```bash
# Download the latest release
curl -L -o pulumi-analyzer-cost.tar.gz \
    https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m).tar.gz

# Extract and install
tar -xzf pulumi-analyzer-cost.tar.gz
VERSION=$(./pulumicost version 2>&1 | grep -oP 'v[\d.]+' || echo "v0.1.0")
PLUGIN_DIR=~/.pulumi/plugins/analyzer-cost-${VERSION}
mkdir -p "$PLUGIN_DIR"
mv pulumicost "$PLUGIN_DIR/pulumi-analyzer-cost"
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/rshade/pulumicost-core.git
cd pulumicost-core

# Build the binary
make build

# Install as Pulumi plugin
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
PLUGIN_DIR=~/.pulumi/plugins/analyzer-cost-${VERSION}
mkdir -p "$PLUGIN_DIR"
cp bin/pulumicost "$PLUGIN_DIR/pulumi-analyzer-cost"
```

### Verify Installation

```bash
# List installed Pulumi plugins
pulumi plugin ls

# Should show:
# NAME    KIND      VERSION
# cost    analyzer  v0.1.0
```

## Configuration

### Enable Analyzer in Pulumi.yaml

Add the `analyzers` section to your `Pulumi.yaml`:

```yaml
name: my-infrastructure
runtime: yaml
description: My cloud infrastructure

# Enable cost estimation
analyzers:
  - cost
```

### Optional: PulumiCost Configuration

Create `~/.pulumicost/config.yaml` for advanced settings:

```yaml
# Analyzer settings
analyzer:
  timeout:
    per_resource: 5s      # Per-resource timeout
    total: 60s            # Overall analysis timeout
    warn_threshold: 30s   # Log warning threshold

# Logging
logging:
  level: info             # debug, info, warn, error
  format: text            # text or json

# Output preferences
output:
  default_format: table
  precision: 2
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PULUMICOST_LOG_LEVEL` | Logging level | `info` |
| `PULUMICOST_LOG_FORMAT` | Log format (text/json) | `text` |
| `PULUMICOST_CONFIG` | Config file path | `~/.pulumicost/config.yaml` |

## Usage

### Basic Usage

Run `pulumi preview` as normal:

```bash
pulumi preview
```

Cost estimates appear in the diagnostics section for each resource.

### With Debug Logging

For troubleshooting:

```bash
PULUMICOST_LOG_LEVEL=debug pulumi preview 2>&1 | tee preview.log
```

### Supported Resources

The analyzer provides cost estimates for common cloud resources:

| Provider | Resource Types |
|----------|---------------|
| AWS | EC2 instances, RDS, S3, Lambda, ELB, etc. |
| Azure | VMs, Storage, App Service, etc. |
| GCP | Compute instances, Cloud Storage, etc. |
| Kubernetes | Pods, Deployments (via Kubecost plugin) |

## Example Output

### Simple AWS Stack

```text
$ pulumi preview

Previewing update (dev):
     Type                     Name          Plan       Info
 +   pulumi:pulumi:Stack      my-app-dev    create
 +   ├─ aws:ec2:Instance      webserver     create
 +   ├─ aws:rds:Instance      database      create
 +   └─ aws:s3:Bucket         assets        create

Diagnostics:
  pulumicost:cost-estimate (webserver):
    INFO: Estimated Monthly Cost: $8.45 USD (source: local-spec)
         Instance type: t3.micro

  pulumicost:cost-estimate (database):
    INFO: Estimated Monthly Cost: $25.00 USD (source: local-spec)
         Instance class: db.t3.micro

  pulumicost:cost-estimate (assets):
    INFO: Estimated Monthly Cost: $0.50 USD (source: local-spec)
         Storage class: Standard

  pulumicost:stack-cost-summary:
    INFO: Total Estimated Monthly Cost: $33.95 USD (3 resources analyzed)
```

## Troubleshooting

### Analyzer Not Running

**Symptom**: No cost diagnostics in preview output.

**Check 1**: Verify plugin is installed:

```bash
ls ~/.pulumi/plugins/ | grep analyzer-cost
```

**Check 2**: Verify analyzer is enabled in Pulumi.yaml:

```yaml
analyzers:
  - cost
```

**Check 3**: Test the analyzer binary directly:

```bash
~/.pulumi/plugins/analyzer-cost-v*/pulumi-analyzer-cost
# Should output a port number and wait
```

### "Unsupported resource type" Warnings

This is normal for resources without pricing data. The analyzer continues with other resources.

To add custom pricing specs:

```bash
mkdir -p ~/.pulumicost/specs
cat > ~/.pulumicost/specs/aws-myservice-default.yaml << EOF
name: aws-myservice-default
provider: aws
service: myservice
pricing:
  monthlyEstimate: 25.00
  currency: USD
EOF
```

### Timeouts

If preview hangs or times out:

1. Check network connectivity to pricing APIs
2. Reduce stack size or increase timeout:

```yaml
# ~/.pulumicost/config.yaml
analyzer:
  timeout:
    total: 120s
```

### Debug Logs

Enable debug logging to diagnose issues:

```bash
PULUMICOST_LOG_LEVEL=debug pulumi preview 2>debug.log
```

Logs go to stderr (required by Pulumi plugin protocol).

## Next Steps

- **[Developer Guide](../../docs/guides/developer-guide.md)**: Extend PulumiCost with custom plugins
- **[Architecture Guide](../../docs/guides/architect-guide.md)**: Understand the plugin architecture
- **[CLI Reference](../../docs/reference/cli.md)**: Full CLI documentation
