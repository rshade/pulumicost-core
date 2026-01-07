# Quickstart: Verify Plugin Version Selection

## Prerequisites

- PulumiCost Core installed or built locally.
- Access to `~/.pulumicost/plugins` directory.

## Verification Steps

### 1. Setup Test Plugins

Create a dummy plugin with multiple versions.

```bash
# Create directories
mkdir -p ~/.pulumicost/plugins/test-plugin/v1.0.0
mkdir -p ~/.pulumicost/plugins/test-plugin/v2.0.0

# Create dummy binaries
touch ~/.pulumicost/plugins/test-plugin/v1.0.0/pulumicost-plugin-test-plugin
chmod +x ~/.pulumicost/plugins/test-plugin/v1.0.0/pulumicost-plugin-test-plugin

touch ~/.pulumicost/plugins/test-plugin/v2.0.0/pulumicost-plugin-test-plugin
chmod +x ~/.pulumicost/plugins/test-plugin/v2.0.0/pulumicost-plugin-test-plugin
```

### 2. Verify List (All Versions)

If a command exists to list all (e.g. `pulumicost plugin list --all`), verify it shows both.

### 3. Verify Selection (Latest)

Run a command that uses plugins (e.g., `pulumicost cost ...` or `pulumicost plugin list` if it defaults to latest).

*Note: As of this plan, the specific CLI command output verification depends on how `plugin list` is implemented. The core logic verification is done via unit tests.*

### 4. Run Unit Tests

To verify the logic without CLI:

```bash
go test -v ./internal/registry/ -run TestListLatestPlugins
```