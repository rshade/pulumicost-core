# Quickstart: Filtering Actual Costs

This guide demonstrates how to use the `--filter` flag with the `finfocus cost actual` command.

## Prerequisites

-   FinFocus CLI installed
-   A Pulumi preview JSON file (e.g., `plan.json`)
-   Configured cloud provider credentials (for fetching actual costs)

## Usage

### Filter by Tag

Filter costs to show only resources with a specific tag.

```bash
finfocus cost actual --pulumi-json plan.json --from 2025-01-01 --filter "tag:Environment=production"
```

### Filter by Resource Type

Filter costs to show only a specific resource type.

```bash
finfocus cost actual --pulumi-json plan.json --from 2025-01-01 --filter "type=aws:s3/bucket"
```

### Combine Filter and Grouping

Filter the dataset first, then group the results. For example, show daily costs for production resources.

```bash
finfocus cost actual --pulumi-json plan.json --from 2025-01-01 --filter "tag:Environment=production" --group-by daily
```

### Multiple Filters

Combine multiple filters (AND logic).

```bash
finfocus cost actual --pulumi-json plan.json --from 2025-01-01 --filter "tag:Environment=production" --filter "type=aws:ec2/instance"
```
