# Quickstart: Testing Pulumi Tool Plugin

This guide explains how to build and test the `finfocus` binary as a Pulumi Tool Plugin locally.

## Prerequisites

- Go 1.22+
- Pulumi CLI installed
- Existing `finfocus` config (optional)

## 1. Build the Plugin Binary

Compile the binary with the specific name required by Pulumi:

```bash
go build -o pulumi-tool-cost ./cmd/finfocus
```

## 2. Install "Locally" (Simulated)

Pulumi looks for plugins in `~/.pulumi/plugins`. Create a directory structure mimicking a real installed plugin:

```bash
mkdir -p ~/.pulumi/plugins/tool-cost-v0.0.0-dev/
cp pulumi-tool-cost ~/.pulumi/plugins/tool-cost-v0.0.0-dev/
```

## 3. Verify Installation

Check if Pulumi recognizes the plugin:

```bash
pulumi plugin ls
# Should see:
# cost  tool  v0.0.0-dev  ...
```

## 4. Run via Pulumi

Execute commands through the Pulumi CLI:

```bash
pulumi plugin run tool cost -- help
```

**Expected Output**:
The help text should show `Usage: pulumi plugin run tool cost ...`

## 5. Test Configuration Isolation (Optional)

1. Set `export PULUMI_HOME=/tmp/fake-pulumi`
2. Run `pulumi plugin run tool cost -- configure ...`
3. Verify `config.yaml` is created inside `/tmp/fake-pulumi/finfocus/`.
