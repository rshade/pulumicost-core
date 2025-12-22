---
title: Analyzer Setup
description: Setting up PulumiCost as a Pulumi Analyzer
layout: default
---

PulumiCost can run as a Pulumi Analyzer, providing cost estimates directly
within your `pulumi preview` and `pulumi up` workflow.

## Prerequisites

- Pulumi CLI installed
- `pulumicost` binary available in your PATH

## Configuration

To enable the analyzer, add the following to your `Pulumi.yaml` file:

```yaml
analyzers:
  - name: pulumicost
    version: v1.0.0 # Match your installed version
```

## Usage

Once configured, run `pulumi preview` as usual. You will see cost estimates
in the diagnostics output.

```bash
pulumi preview
```

## Troubleshooting

If you don't see cost estimates:

1. Ensure `pulumicost` is in your PATH.
2. Check `Pulumi.yaml` syntax.
3. Run `pulumi preview --debug` to see if the analyzer is being loaded.
