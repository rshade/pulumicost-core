# Data Model: Plugin Info and DryRun Discovery

## Entities

### PluginMetadata
Represents the metadata returned by a plugin via `GetPluginInfo`.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Unique name of the plugin |
| Version | string | Semantic version of the plugin implementation |
| SpecVersion | string | Semantic version of the PulumiCost spec implemented |
| SupportedProviders | []string | List of cloud providers the plugin supports |
| Metadata | map[string]string | Additional provider-specific metadata |

### FieldMapping
Represents a FOCUS field support status returned by `DryRun`.

| Field | Type | Description |
|-------|------|-------------|
| FieldName | string | Name of the FOCUS field (e.g., "ServiceCategory") |
| Status | string | One of: SUPPORTED, UNSUPPORTED, CONDITIONAL, DYNAMIC |
| Condition | string | Human-readable explanation of the condition |
| ExpectedType | string | Data type (string, double, etc.) |

## Relationships

- A `PluginClient` (in `pluginhost`) will now hold a `PluginMetadata` object after successful initialization.
- The `Engine` will be able to query `FieldMapping` through the `PluginClient`.

## Validation Rules

- `spec_version` must be a valid SemVer string.
- `name` and `version` must be non-empty.
- `DryRun` requests must include a valid `ResourceDescriptor`.
