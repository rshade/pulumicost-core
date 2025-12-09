---
title: Error Codes Reference
description: PulumiCost error codes and troubleshooting guidance
layout: default
---

PulumiCost provides clear error messages and codes to help diagnose and resolve issues.
This reference categorizes common errors and suggests solutions.

## Engine Errors

These errors typically occur during the cost calculation and aggregation process within
the PulumiCost engine.

| Code | Message | Description | Solution |
| --- | --- | --- | --- |
| `ErrNoCostData` | "no cost data available" | No pricing information found | Ensure plugins are installed |
| `ErrMixedCurrencies` | "mixed currencies not supported" | Different currencies in aggregation | Filter by currency first |
| `ErrInvalidGroupBy` | "invalid groupBy type" | Unsupported grouping type | Use `daily` or `monthly` |
| `ErrEmptyResults` | "empty results" | No data to aggregate | Check input file and filters |
| `ErrInvalidDateRange` | "end date must be after start" | Invalid date range | Ensure `--to` > `--from` |
| `ErrResourceValidation` | "resource validation failed" | Internal validation error | Report bug with details |

## Configuration Errors

These errors relate to parsing or validating the PulumiCost configuration file
(`~/.pulumicost/config.yaml`).

| Code | Message | Description | Solution |
| --- | --- | --- | --- |
| `ErrConfigCorrupted` | "configuration file corrupted" | Cannot parse config file | Fix YAML or delete config |

## CLI Errors

Errors related to command-line argument parsing or general CLI usage.

| Message | Description | Solution |
| --- | --- | --- |
| "invalid date format" | Date not recognized | Use `YYYY-MM-DD` or RFC3339 |
| "date range cannot exceed 366 days" | Date range too wide | Reduce range to max 366 days |
| "plugin not found" | Plugin not in registry | Verify plugin name |
| "plugin already installed" | Plugin exists | Use `update` or `--force` |
