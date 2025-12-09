---
title: Analyzer Setup Quickstart
description: Quick setup guide for PulumiCost Pulumi Analyzer integration
layout: default
---

## Zero-Click Cost Estimation

PulumiCost integrates directly with the Pulumi CLI as an analyzer, providing cost
estimates automatically during `pulumi preview`.

### Prerequisites

- Pulumi CLI installed
- `pulumicost` installed and in your PATH

### Configuration

Add the analyzer to your project's `Pulumi.yaml`:

```yaml
name: my-project
runtime: go # or your chosen runtime
description: A Pulumi project with cost analysis
plugins:
  - path: pulumicost
    args: ["analyzer", "serve"]
```

### Usage

Run a preview as normal. Cost estimates will appear in the diagnostics output:

```text
Diagnostics:
  pulumicost:
    Type: aws:ec2/instance:Instance
    ID:   web-server
    Cost: $7.59/month (est.)
```

### Troubleshooting

If you don't see cost estimates:

1. **Check Logs**: Run with debug logging enabled:

   ```bash
   PULUMICOST_LOG_LEVEL=debug pulumi preview
   ```

   Check `~/.pulumicost/logs/pulumicost.log`.

2. **Verify Port**: The analyzer runs on a dynamic port. Ensure no firewalls are
   blocking localhost traffic.

3. **Strict Mode**: If config issues are suspected, enable strict mode:

   ```bash
   PULUMICOST_CONFIG_STRICT=true pulumi preview
   ```
