---
title: Pulumi Analyzer Integration
description: Technical details of the PulumiCost analyzer gRPC integration
layout: default
---

PulumiCost integrates with Pulumi's analyzer framework to provide real-time cost
estimates during `pulumi preview` operations.

## Architecture

The analyzer is a gRPC server that the Pulumi engine invokes during the resource
lifecycle. It provides "advisory" diagnostics that do not block deployment but
inform the user of potential costs.

For setup instructions, see the [Analyzer Setup Guide](getting-started/analyzer-setup.md).

## How It Works

The analyzer is invoked automatically by Pulumi during preview. It:

1. Receives resource information via gRPC
2. Calculates costs using the pricing engine and plugins
3. Returns cost diagnostics that appear in the Pulumi output

## Key Technical Details

### Binary Naming Convention

Pulumi looks for `pulumi-analyzer-policy-<runtime>` on PATH or in the specified policy pack directory. Since we use
`runtime: pulumicost` in `PulumiPolicy.yaml`, the binary must be named:

```text
pulumi-analyzer-policy-pulumicost
```

### Handshake Protocol

When Pulumi starts the analyzer:

1. Pulumi executes the binary.
2. The binary detects it is being run as an analyzer (by checking `os.Args[0]`).
3. The binary starts a gRPC server on a random available port.
4. Binary prints the port number to `stdout` (ONLY the port, nothing else).
5. Pulumi connects to that port via gRPC.
6. **Important**: All logging must go to `stderr` to avoid breaking the handshake.

### RPC Methods

The analyzer implements these Pulumi Analyzer gRPC methods:

| Method           | Purpose                                      |
| ---------------- | -------------------------------------------- |
| `Handshake`      | Acknowledge connection from Pulumi engine    |
| `GetAnalyzerInfo`| Return analyzer metadata and policy info     |
| `GetPluginInfo`  | Return plugin version                        |
| `ConfigureStack` | Receive stack context before analysis        |
| `Analyze`        | Analyze single resource, return diagnostics  |
| `AnalyzeStack`   | Called at end, return summary diagnostic     |
| `Cancel`         | Handle graceful shutdown                     |

### Diagnostic Workflow

1. `ConfigureStack` is called once at start (clears cost cache).
2. `Analyze` is called for each resource (returns per-resource costs, caches them).
3. `AnalyzeStack` is called once at end (returns summary using cached costs).

This prevents duplicate diagnostics in the output and ensures the summary is
accurate based on the resources analyzed in the current run.

### Enforcement Level

All diagnostics use `ADVISORY` enforcement, meaning they never block deployments.
Costs are informational only.

## Internal Types

Pulumi internal resources (Stack, providers) are handled specially:

- Type prefix: `pulumi:`
- Cost: $0.00
- Message: "Internal Pulumi resource (no cloud cost)"

## See Also

- [Analyzer Setup Guide](getting-started/analyzer-setup.md)
- [CLI Commands Reference](reference/cli-commands.md)
