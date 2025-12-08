# Plan: Pulumi Tool Plugin Integration

## Objective
Enable `pulumicost` to be executed as a Pulumi Tool Plugin via the command `pulumi plugin run cost`. This allows for seamless integration into Pulumi workflows and leverages Pulumi's authentication and context.

## Overview
Pulumi "Tool" plugins are auxiliary binaries that can be managed and executed by the Pulumi CLI. By packaging `pulumicost` as a tool plugin, users can invoke it directly through Pulumi, and the binary gains access to the Pulumi environment (API tokens, workspace context, and schema services).

## Architecture

### Execution Flow
1.  User runs: `pulumi plugin run tool cost -- [args]`
2.  Pulumi CLI resolves the plugin path: `~/.pulumi/plugins/tool-cost-vX.Y.Z/pulumi-tool-cost`.
3.  Pulumi starts a local gRPC server (Plugin Host).
4.  Pulumi executes the `pulumi-tool-cost` binary.
    *   **Stdin/Stdout/Stderr**: Piped to the user's terminal.
    *   **Environment Variables**: Injected by Pulumi.

### Injected Context
The tool receives the following environment variables:
*   `PULUMI_RPC_TARGET`: Address of the Pulumi Plugin Host gRPC server (Schema Loader/Mapper).
*   `PULUMI_API`: The Pulumi Cloud API URL (if authenticated).
*   `PULUMI_ACCESS_TOKEN`: The user's current access token.
*   `PULUMI_HOME`: Path to the Pulumi configuration directory.

## Implementation Steps

### 1. Code Adjustments
While simply renaming the binary allows it to execute, a robust integration requires code changes to ensure a consistent user experience. The current codebase hardcodes the application name as "pulumicost", which affects help text, configuration paths, and environment variables.

**Required Changes:**
*   **Dynamic Name Detection**: The application should detect its binary name (`os.Args[0]`) at startup.
    *   If named `pulumi-tool-cost`, it should auto-configure for plugin mode.
*   **CLI Definition (`internal/cli/root.go`)**:
    *   Update `Use: "pulumicost"` to use the detected binary name.
    *   Update help text and examples to reflect the invocation method (`pulumi plugin run cost` vs `pulumicost`).
*   **Configuration (`internal/config`)**:
    *   **Env Vars**: The user has decided to strictly maintain `PULUMICOST_` prefixes to ensure a clear separation of concerns from the core Pulumi runtime.
    *   **Paths**: Avoid hardcoding `~/.pulumicost`. If running as a plugin, consider using `PULUMI_HOME` or a standard XDG path to avoid scattering config files.

### 2. Binary Packaging
Once the code supports dynamic naming:

*   **Binary Name**: Must be `pulumi-tool-cost`.
*   **Build Command**:
    ```bash
    go build -o pulumi-tool-cost ./cmd/pulumicost
    ```

### 2. Installation Structure
Pulumi looks for plugins in `~/.pulumi/plugins/`. The directory structure must be:

```text
$HOME/.pulumi/plugins/
└── tool-cost-v<version>/
    └── pulumi-tool-cost  <-- The binary
```

### 3. Verification
Running `pulumi plugin ls` should list the plugin:

```text
NAME  KIND  VERSION  SIZE    INSTALLED   LAST USED
cost  tool  v0.4.3   25MB    <time>      <time>
```

### 4. Enhanced Integration (Future Work)
To fully utilize the "Tool" status, `pulumicost` can be updated to use the injected environment variables.

*   **Cloud Backend Integration**: Use `PULUMI_API` and `PULUMI_ACCESS_TOKEN` to fetch the stack's deployment state directly from Pulumi Cloud, eliminating the need for users to manually export JSON with `pulumi stack export`.
*   **Schema Resolution**: Connect to `PULUMI_RPC_TARGET` to query provider schemas. This allows `pulumicost` to dynamically discover which resource properties constitute "pricing dimensions" (e.g., detecting that `instanceType` is a key property for `aws:ec2/instance:Instance`) without hardcoded mapping logic.

## Usage Guide

**Basic Usage:**
```bash
pulumi plugin run tool cost -- cost projected --pulumi-json plan.json
```

**With Alias (User-side):**
Users can alias this command for brevity:
```bash
alias pulumicost="pulumi plugin run tool cost --"
pulumicost cost actual
```
