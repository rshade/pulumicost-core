---
title: Analyzer Architecture
description: Architecture of the FinFocus Analyzer integration
layout: default
---

FinFocus integrates with the Pulumi engine via the Analyzer interface.
This allows FinFocus to intercept resource changes during `pulumi preview`
and `pulumi up` to provide cost estimates and policy enforcement.

## Overview

The Analyzer runs as a gRPC service that Pulumi connects to. When you run
a Pulumi command, if the analyzer is configured, Pulumi sends resource
definitions to FinFocus.

## Protocol

FinFocus implements the `pulumirpc.Analyzer` service.

- **Analyze**: Receives a resource and its properties. FinFocus calculates
  the cost.
- **AnalyzeStack**: Receives the entire stack state.

## Configuration

The analyzer is configured via `Pulumi.yaml` using the `analyzers` key:

```yaml
analyzers:
  - name: finfocus
    version: v1.0.0
```

## Diagnostics

If a resource violates a cost policy (e.g., exceeds budget), FinFocus
returns a diagnostic error or warning, which Pulumi displays in the CLI.
