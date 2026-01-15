# CLI Contract: Pulumi Tool Plugin

## Invocation Signals

The CLI behavior changes based on how it is invoked.

### 1. Plugin Mode Trigger

**Condition**:
- Binary Name (base) == `pulumi-tool-cost` (case-insensitive, ignores extension)
- OR Env `FINFOCUS_PLUGIN_MODE` == `true` (or `1`)

**Effect**:
- Root Command `Use`: `pulumi plugin run tool cost`
- Help Text Examples: Prefix `pulumi plugin run tool cost` instead of `finfocus`

### 2. Standard Mode

**Condition**:
- Triggers above are NOT met.

**Effect**:
- Root Command `Use`: `finfocus`
- Help Text Examples: Prefix `finfocus`

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `FINFOCUS_PLUGIN_MODE` | No | Forces plugin mode if set to `true`. |
| `PULUMI_HOME` | No | If set, modifies config search path to `$PULUMI_HOME/finfocus/`. |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General Error (Config, API, logic) |
