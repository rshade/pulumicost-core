# Plan: Pulumi Analyzer Plugin Integration

## Objective
Integrate `pulumicost` as a Pulumi **Analyzer Plugin**. This enables cost estimation and policy enforcement to run automatically during the `pulumi preview` and `pulumi up` lifecycle events.

## Product Vision

**User Story**
> "As a Cloud Infrastructure Engineer, I want to see the estimated cost impact of my changes directly within the `pulumi preview` output so that I can optimize spend *before* resources are provisioned, without context switching to a separate tool or exporting files manually."

**Strategic Value (Why this is the Golden Path)**

1.  **Frictionless Adoption (The "Zero-Click" Experience)**
    *   **Current State:** External CLI tools require manual steps (exporting JSON, running commands), leading to forgetfulness and low adoption.
    *   **Future State:** The Analyzer runs automatically with every `pulumi preview`. Cost visibility becomes a default, unavoidable part of the workflow.

2.  **"Shifting Left" (Proactive vs. Reactive)**
    *   **Impact:** Moves cost management from monthly reporting (reactive) to the engineering design phase (proactive). Seeing `+ $45/mo` alongside a resource creation prompt directly influences developer behavior in the moment of decision.

3.  **The "Policy as Code" Moat**
    *   **Competitive Advantage:** Unlike a standalone CLI tool, an Analyzer integrates with Pulumi's policy engine. This allows us to eventually *gate* deployments (e.g., "Fail build if monthly cost increase > $100"), transforming `pulumicost` from a passive calculator into an active compliance guardrail.

## Overview
An Analyzer Plugin allows external tools to inspect Pulumi resources during deployment. By implementing the `pulumirpc.Analyzer` gRPC interface, `pulumicost` can receive the resource graph from the Pulumi engine, calculate costs, and return "diagnostics" (info, warnings, or errors) displayed directly in the Pulumi CLI output.

## Architecture

### gRPC Interface
The plugin must serve the `Analyzer` service defined in `pulumi/sdk/proto/go/pulumirpc`.

**Key Methods:**
*   `Analyze(ctx, request)`: Called for individual resources. (Likely skipped for cost, as cost is often aggregate).
*   `AnalyzeStack(ctx, request)`: Called with the full list of resources in the stack. This is the primary integration point.
*   `GetPluginInfo(ctx, void)`: Returns plugin metadata.

### Workflow
1.  **Initialization**: Pulumi starts `pulumi-analyzer-cost` and waits for it to print its listening port to stdout.
2.  **Handshake**: Pulumi connects to the port via gRPC.
3.  **Analysis**: During a preview/update, Pulumi sends the fully resolved resource graph directly to `AnalyzeStack` via gRPC. **Note:** This eliminates the need for the user to manually generate or pass a `plan.json` file, as the resource data is streamed in memory.
4.  **Calculation**: `pulumicost` converts these in-memory gRPC resources to its internal `ResourceDescriptor` format and runs the pricing engine.
5.  **Reporting**: `pulumicost` returns `AnalyzeDiagnostic` messages containing cost estimates.

## Implementation Steps

### 1. Dependencies
Add the Pulumi SDK dependency:
```bash
go get github.com/pulumi/pulumi/sdk/v3
```

### 2. Server Implementation
Create a new package `internal/pulumiapi` containing the gRPC server.

**`internal/pulumiapi/analyzer.go` (Conceptual):**
```go
type CostAnalyzer struct {
    pulumirpc.UnimplementedAnalyzerServer
}

func (a *CostAnalyzer) AnalyzeStack(ctx context.Context, req *pulumirpc.AnalyzeStackRequest) (*pulumirpc.AnalyzeResponse, error) {
    // 1. Map req.Resources to []engine.Resource
    // 2. Call engine.CalculateProjectedCost()
    // 3. Format result.TotalCost into a diagnostic message
    return &pulumirpc.AnalyzeResponse{
        Diagnostics: []*pulumirpc.AnalyzeDiagnostic{
            {
                PolicyName: "cost-estimation",
                Message:    fmt.Sprintf("Estimated Monthly Cost: $%.2f", cost),
                Severity:   pulumirpc.AnalyzeDiagnostic_INFO,
            },
        },
    }, nil
}
```

### 3. Entry Point Modification
Modify `cmd/pulumicost/main.go` to support a "serve" mode.

*   **Flag**: Add hidden flag `--serve` or distinct subcommand.
*   **Behavior**:
    *   Start TCP listener on random port.
    *   Register `AnalyzerServer`.
    *   `fmt.Printf("%d\n", port)` (Critical: this is how Pulumi finds the server).
    *   Block indefinitely.

**Configuration & Context:**
Similar to the Tool plugin, the Analyzer must respect the context it runs in.
*   **Config Path**: Should not default to `~/.pulumicost`. It should likely use the project's context or a temporary directory.
*   **Logging**: Logs should be directed to a file or `stderr` (carefully), as `stdout` is reserved for the port handshake. Using Pulumi's RPC logging interface is preferred once connected.

### 4. Resource Mapping
Implement a mapper to convert `pulumirpc.Resource` (Protobuf-based, weakly typed map) into `pulumicost`'s `ResourceDescriptor`. This logic likely already exists partially in the `ingest` package for JSON parsing and will need adaptation for Protobuf structs.

## Configuration & Usage

### Installation
Binary must be named `pulumi-analyzer-cost` and placed in `~/.pulumi/plugins/analyzer-cost-vX.Y.Z/`.

### Project Configuration (`Pulumi.yaml`)
Users enable the analyzer in their project file:

```yaml
name: my-project
runtime: nodejs
description: A project with cost analysis
analyzers:
  - cost
```

### Output
When the user runs `pulumi preview`:

```text
Previewing update (dev)

     Type                 Name           Plan
 +   pulumi:pulumi:Stack  my-project-dev create
 +   aws:s3:Bucket        my-bucket      create

Diagnostics:
  cost:cost-estimation:
    info: Estimated Monthly Cost: $24.50
```

