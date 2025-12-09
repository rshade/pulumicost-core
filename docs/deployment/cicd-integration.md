---
title: CI/CD Integration
description: Integrate PulumiCost into CI/CD pipelines for shift-left cost analysis
layout: default
---

Integrating PulumiCost into your CI/CD pipelines helps prevent unexpected cloud costs
and provides early feedback on infrastructure changes.

## Goals of CI/CD Integration

- **Shift-Left Cost Analysis**: Catch cost implications early in development.
- **Automated Cost Reporting**: Generate cost summaries for deployments.
- **Policy Enforcement**: Fail builds if cost policies are violated.

## Basic Integration with `pulumi preview` (Zero-Click)

The easiest way to integrate PulumiCost is using the Analyzer plugin with `pulumi
preview`, providing "zero-click" cost estimation in your Pulumi diagnostics.

1. **Configure Pulumi.yaml**:
   Ensure your `Pulumi.yaml` is configured to use the PulumiCost analyzer:

   ```yaml
   name: my-project
   runtime: go # or your chosen runtime
   description: A Pulumi project with cost analysis
   plugins:
     - path: pulumicost
       args: ["analyzer", "serve"]
   ```

2. **Run `pulumi preview` in CI**:
   Your CI pipeline should already be running `pulumi preview`. No extra steps are
   needed for PulumiCost once the analyzer is configured. The cost estimates will
   appear in the `pulumi preview` output as diagnostics.

   ```bash
   # Example GitHub Actions step
   - name: Run Pulumi Preview with Cost Analysis
     run: pulumi preview --diff --json > pulumi_preview.json
   ```

## Advanced Cost Reporting

For more detailed reports or to integrate with external systems, you can use the
`pulumicost cost` commands directly.

### Generating Daily/Monthly Cost Trends

```bash
# Example: Generate a monthly cost report for a specific stack
- name: Generate Monthly Cost Report
  run: |
    pulumi stack select my-prod-stack # Select your target stack
    pulumi preview --json > pulumi_preview.json
    pulumicost cost actual \
      --pulumi-json pulumi_preview.json \
      --from "$(date -I -d 'last month')" \
      --to "$(date -I)" \
      --group-by monthly \
      --output json > monthly_cost_report.json
```

### Passing Credentials via Environment Variables

It is highly recommended to pass sensitive plugin credentials (e.g., API keys, tokens)
to PulumiCost via environment variables in your CI/CD system. This avoids storing them
in configuration files within your repository.

Refer to the [Environment Variables Reference](../reference/environment-variables.md)
for a comprehensive list of variables and the `PULUMICOST_PLUGIN_<NAME>_<KEY>` pattern
for plugin-specific settings.

```bash
# Example GitHub Actions secrets
env:
  PULUMICOST_PLUGIN_VANTAGE_API_KEY: ${{ secrets.VANTAGE_API_KEY }}
  PULUMICOST_LOG_LEVEL: info # Or debug for more verbose CI logs
```

### Enforcing Cost Policies

You can use the output of `pulumicost` commands in JSON or NDJSON format to implement
custom cost policies in your CI/CD.

```bash
# Example: Fail if a projected cost exceeds a threshold
- name: Check Projected Cost Threshold
  id: cost_check
  run: |
    pulumi preview --json > pulumi_preview.json
    COST_CHANGE=$(pulumicost cost projected \
      --pulumi-json pulumi_preview.json --output json | jq '.summary.totalMonthly')
    THRESHOLD=100.00 # Example: $100 monthly increase threshold

    if (( $(echo "$COST_CHANGE > $THRESHOLD" | bc -l) )); then
      echo "::error title=Cost Policy Violation::Cost exceeds threshold."
      exit 1
    fi
```

For more advanced policy enforcement, consider using dedicated policy-as-code tools
that can parse Pulumi's plan output and integrate with PulumiCost's cost data.
