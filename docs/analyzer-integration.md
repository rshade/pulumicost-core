# Pulumi Analyzer Integration

PulumiCost integrates with Pulumi's analyzer framework to provide real-time cost
estimates during `pulumi preview` operations.

## How It Works

The analyzer is invoked automatically by Pulumi during preview. It:

1. Receives resource information via gRPC
2. Calculates costs using the pricing engine and plugins
3. Returns cost diagnostics that appear in the Pulumi output

## Quick Start

### 1. Build the Binary

```bash
make build
```

### 2. Create a Policy Pack Directory

```bash
mkdir -p /path/to/policy-pack

# Copy binary with correct name
cp bin/pulumicost /path/to/policy-pack/pulumi-analyzer-policy-pulumicost
chmod +x /path/to/policy-pack/pulumi-analyzer-policy-pulumicost

# Create PulumiPolicy.yaml
cat > /path/to/policy-pack/PulumiPolicy.yaml << 'EOF'
runtime: pulumicost
name: pulumicost
version: 0.0.0-dev
EOF
```

### 3. Run Preview with Policy Pack

```bash
# Add policy pack directory to PATH
export PATH="/path/to/policy-pack:$PATH"

# Run preview
pulumi preview --policy-pack /path/to/policy-pack
```

## Example Output

```text
Policies:
    pulumicost@v0.0.0-dev (local: /path/to/policy-pack)
        - [advisory] [severity: medium]  cost-estimate  (pulumi:providers:aws: default)
          Internal Pulumi resource (no cloud cost)
        - [advisory] [severity: low]  cost-estimate  (aws:ec2/instance:Instance: my-instance)
          Estimated Monthly Cost: $7.59 USD (source: pulumicost-plugin-aws-public)
        - [advisory] [severity: low]  stack-cost-summary  (pulumi:pulumi:Stack: my-stack)
          Total Estimated Monthly Cost: $7.59 USD (1 resources analyzed)
```

## Key Technical Details

### Binary Naming Convention

Pulumi looks for `pulumi-analyzer-policy-<runtime>` on PATH. Since we use
`runtime: pulumicost` in PulumiPolicy.yaml, the binary must be named:

```text
pulumi-analyzer-policy-pulumicost
```

### Handshake Protocol

When Pulumi starts the analyzer:

1. Pulumi executes the binary
2. Binary prints port number to stdout (ONLY the port, nothing else)
3. Pulumi connects to that port via gRPC
4. All logging must go to stderr to avoid breaking the handshake

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

1. `ConfigureStack` is called once at start (clears cost cache)
2. `Analyze` is called for each resource (returns per-resource costs, caches them)
3. `AnalyzeStack` is called once at end (returns summary using cached costs)

This prevents duplicate diagnostics in the output.

### Enforcement Level

All diagnostics use `ADVISORY` enforcement, meaning they never block deployments.
Costs are informational only.

## Environment Variables

| Variable              | Description                           | Default |
| --------------------- | ------------------------------------- | ------- |
| `PULUMICOST_LOG_LEVEL`| Logging level (debug, info, warn)     | `info`  |

## Troubleshooting

### "could not start policy pack"

Ensure the binary is named correctly and on PATH:

```bash
which pulumi-analyzer-policy-pulumicost
```

### No diagnostics appearing

1. Verify the policy pack directory has both files:
   - `pulumi-analyzer-policy-pulumicost` (executable)
   - `PulumiPolicy.yaml`

2. Check the runtime in PulumiPolicy.yaml matches the binary suffix

3. Run with debug logging:

   ```bash
   PULUMICOST_LOG_LEVEL=debug pulumi preview --policy-pack /path
   ```

### Empty port on stdout

The binary detects its name and only enters analyzer mode when named
`pulumi-analyzer-policy-pulumicost` or `pulumi-analyzer-pulumicost`.

Test manually:

```bash
cp bin/pulumicost /tmp/pulumi-analyzer-policy-pulumicost
/tmp/pulumi-analyzer-policy-pulumicost
# Should print a port number like "41234"
```

## Internal Types

Pulumi internal resources (Stack, providers) are handled specially:

- Type prefix: `pulumi:`
- Cost: $0.00
- Message: "Internal Pulumi resource (no cloud cost)"

## See Also

- [Plugin Development](plugins/plugin-development.md)
- [CLI Reference](reference/cli-reference.md)
