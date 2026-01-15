# Research: Pulumi Tool Plugin Integration

**Feature**: 110-pulumi-tool-plugin
**Date**: 2025-12-30

## Research Questions

### 1. Does Pulumi require specific exit codes for tool plugins?

**Finding**:
Pulumi's plugin system generally treats any non-zero exit code as a failure. There is no strict public specification reserving specific codes (like 10-15) for specific internal errors for *tool* plugins (unlike Language Hosts which have a tighter contract).

**Decision**:
Use standard POSIX exit codes:
- `0`: Success
- `1`: General Error (Runtime, Config, API)
- `127`: Command not found (if applicable to sub-commands)

**Rationale**:
Keeps integration simple and predictable. Pulumi CLI will simply bubble up the non-zero exit status to the user.

### 2. PULUMI_HOME vs XDG Standards

**Finding**:
Pulumi uses `PULUMI_HOME` as its primary override for the configuration directory. If unset, it defaults to `~/.pulumi`. While some modern tools prefer `~/.config` (XDG), acting as a "Pulumi Tool" implies we should respect the Pulumi ecosystem's conventions first.

**Decision**:
Precedence Order:
1. `PULUMI_HOME/finfocus/` (if `PULUMI_HOME` is set)
2. `XDG_CONFIG_HOME/finfocus/` (if `FINFOCUS_HOME` not set, standard Linux fallback)
3. `~/.finfocus` (Legacy/Default fallback)

**Rationale**:
This fulfills the requirement to be a "good citizen" in the Pulumi ecosystem (Scenario 3) while maintaining backward compatibility and standard Linux behavior when running standalone.

## Alternatives Considered

- **Strict Pulumi Only**: Ignoring XDG entirely. Rejected because `finfocus` is still a standalone tool and should behave like one when not acting as a plugin.
- **Hidden Config (`.finfocus`) inside `PULUMI_HOME`**: Rejected in favor of non-hidden `finfocus` folder inside `PULUMI_HOME` as that directory is already a "config/state" root, so hiding files inside it is unnecessary (similar to how `~/.config/appname` is not hidden).

## Implementation Details

- **Binary Name Detection**: `filepath.Base(os.Args[0])` matches `pulumi-tool-cost` (ignoring case/extension).
- **Env Var**: `FINFOCUS_PLUGIN_MODE=true` forces the behavior.
