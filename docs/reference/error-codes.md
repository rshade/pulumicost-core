---
title: Error Codes
description: Reference for Pulumicost error codes
layout: default
---

This page lists common error codes and messages returned by Pulumicost.

## CLI Errors

### ERR-001: Config File Error

**Message**: "Failed to load configuration file"
**Cause**: The configuration file at `~/.pulumicost/config.yaml` is invalid
or unreadable.
**Fix**: Check permissions and YAML syntax.

### ERR-002: Plugin Not Found

**Message**: "Plugin [name] not found"
**Cause**: The requested plugin is not installed in `~/.pulumicost/plugins`.
**Fix**: Run `pulumicost plugin install [name]`.

## Engine Errors

### ENG-001: Pricing Lookup Failed

**Message**: "No pricing data found for resource"
**Cause**: The resource type or SKU is not supported by the pricing provider.
**Fix**: Check if the resource is supported or add a local override in
`~/.pulumicost/specs/`.

### ENG-002: Plugin Timeout

**Message**: "Plugin request timed out"
**Cause**: The plugin took too long to respond (default 10s).
**Fix**: Check network connectivity or increase timeout via environment
variable.
