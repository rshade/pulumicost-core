---
title: Analyzer Architecture
description: Architecture of the Pulumicost Analyzer integration
layout: default
---

Pulumicost integrates with the Pulumi engine via the Analyzer interface.
This allows Pulumicost to intercept resource changes during `pulumi preview`
and `pulumi up` to provide cost estimates and policy enforcement.

## Overview

The Analyzer runs as a gRPC service that Pulumi connects to. When you run
a Pulumi command, if the analyzer is configured, Pulumi sends resource
definitions to Pulumicost.

## Protocol

Pulumicost implements the `pulumirpc.Analyzer` service.

- **Analyze**: Receives a resource and its properties. Pulumicost calculates
  the cost.
- **AnalyzeStack**: Receives the entire stack state.

## Configuration

The analyzer is configured via `Pulumi.yaml` using the `analyzers` key:

```yaml
analyzers:
  - name: pulumicost
    version: v1.0.0
```

## Diagnostics

If a resource violates a cost policy (e.g., exceeds budget), Pulumicost
returns a diagnostic error or warning, which Pulumi displays in the CLI.
