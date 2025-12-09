---
title: Pulumi Analyzer Integration
description: Architecture and protocol details for PulumiCost Pulumi Analyzer integration
layout: default
---

## Overview

PulumiCost integrates with the Pulumi CLI as an Analyzer, providing "Zero-Click" cost
estimation during `pulumi preview`. This delivers instant feedback on infrastructure
change costs directly within your Pulumi workflow.

## Architecture

The Pulumi Analyzer integration operates as a gRPC plugin. When `pulumicost analyzer
serve` is configured in your `Pulumi.yaml`, the Pulumi CLI communicates with the
PulumiCost analyzer via a local gRPC connection.

1. **Server Startup and Port Handshake**: Pulumi CLI launches the `pulumicost analyzer
   serve` process. PulumiCost's analyzer listens on a dynamic port and communicates it
   back to Pulumi via stdout.
2. **Resource Mapping**: The analyzer receives resource descriptors from Pulumi, maps
   them to an internal format compatible with PulumiCost's pricing engine.
3. **Cost Calculation**: The internal pricing engine, leveraging configured pricing
   plugins, calculates the estimated costs for each resource.
4. **Diagnostic Generation**: Estimated costs are returned to the Pulumi CLI as
   diagnostics with `ADVISORY` severity, appearing in the `pulumi preview` output.

## Configuration

To enable the PulumiCost Analyzer, add the following to your Pulumi project's
`Pulumi.yaml` file:

```yaml
name: my-project
runtime: go # or your chosen runtime
description: A Pulumi project with cost analysis
plugins:
  - path: pulumicost
    args: ["analyzer", "serve"]
```

The analyzer's behavior can be further configured via `~/.pulumicost/config.yaml`
within the `analyzer` section. Refer to the
[Configuration Reference](/reference/config-reference.md) for details on setting
timeouts and plugin-specific options for the analyzer.

## Protocol Details

The PulumiCost Analyzer implements the Pulumi Analyzer gRPC protocol. Key aspects
include:

- **gRPC RPCs**: The analyzer responds to standard Pulumi Analyzer RPCs for resource
  analysis.
- **Stdout/Stderr Behavior**: The initial port handshake occurs over stdout. Standard
  logging output is directed to stderr or configured log files.
- **Graceful Shutdown**: The analyzer is designed to shut down gracefully when the
  Pulumi CLI terminates its connection.
