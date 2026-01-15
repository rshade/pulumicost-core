# Data Model: Pulumi Tool Plugin Integration

## Configuration Entities

### Config Resolution Context

The logic for determining where to load configuration from.

| Field | Type | Description | Source |
|-------|------|-------------|--------|
| `PulumiHome` | `string` | Path to Pulumi configuration directory | Env: `PULUMI_HOME` |
| `PluginMode` | `boolean` | Whether the app is running as a plugin | Binary Name or Env: `FINFOCUS_PLUGIN_MODE` |
| `ConfigDir` | `string` | The resolved directory to load `config.yaml` from | Derived |

### Derived Logic (ConfigDir)

```pseudo
if PluginMode AND PulumiHome != "":
    return Join(PulumiHome, "finfocus")
else:
    return UserConfigDir() // XDG or Home
```

## Plugin Context (Injected)

These variables are available to the process when run by Pulumi, though currently `finfocus` only actively uses `PULUMI_HOME`.

| Env Variable | Description | Usage in Feature |
|--------------|-------------|------------------|
| `PULUMI_HOME` | Pulumi config root | Used to anchor `finfocus` config |
| `PULUMI_API` | Cloud API URL | Ignored (Future Use) |
| `PULUMI_ACCESS_TOKEN` | User Token | Ignored (Future Use) |
| `PULUMI_RPC_TARGET` | gRPC Address | Ignored (Future Use) |
