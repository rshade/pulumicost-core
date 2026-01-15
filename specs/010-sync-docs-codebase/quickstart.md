# Quickstart: Analyzer Setup

## Zero-Click Cost Estimation

FinFocus integrates directly with the Pulumi CLI as an analyzer, providing cost estimates automatically during `pulumi preview`.

### Prerequisites

- Pulumi CLI installed
- `finfocus` installed and in your PATH

### Configuration

Add the analyzer to your project's `Pulumi.yaml`:

```yaml
name: my-project
runtime: go
description: A Go Pulumi program
plugins:
  - path: finfocus
    args: ["analyzer", "serve"]
```

### Usage

Run a preview as normal. Cost estimates will appear in the diagnostics output:

```bash
pulumi preview
```

**Example Output:**

```text
Diagnostics:
  finfocus:
    Type: aws:ec2/instance:Instance
    ID:   web-server
    Cost: $7.59/month (est.)
```

### Troubleshooting

If you don't see cost estimates:

1. **Check Logs**: Run with debug logging enabled:
   ```bash
   FINFOCUS_LOG_LEVEL=debug pulumi preview
   ```
   Check `~/.finfocus/logs/finfocus.log`.

2. **Verify Port**: The analyzer runs on a dynamic port. Ensure no firewalls are blocking localhost traffic.

3. **Strict Mode**: If config issues are suspected, enable strict mode:
   ```bash
   FINFOCUS_CONFIG_STRICT=true pulumi preview
   ```
