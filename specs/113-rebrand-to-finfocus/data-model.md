# Data Model: FinFocus

## File Structure

### Configuration Directory (`~/.finfocus`)

| File/Dir | Purpose | Format | Notes |
|----------|---------|--------|-------|
| `config.yaml` | User settings | YAML | Renamed from `config.json` preference in v0.1? No, core supports both. |
| `plugins/` | Plugin binaries | Binary | Executables named `finfocus-plugin-<name>` |
| `logs/` | Application logs | Text | `finfocus.log` |
| `cache/` | Pricing cache | JSON/Gob | Persisted pricing data |

## Configuration Entity (`config.yaml`)

```yaml
finfocus:
  logLevel: "info"      # formerly finfocus.logLevel
  outputFormat: "table" # formerly finfocus.outputFormat
  concurrency: 4        # formerly finfocus.concurrency
  
  # Plugin configuration
  plugins:
    aws-public:
      region: "us-west-2"
```

## Migration State

The application logic does not persist a "migration state" file, but rather infers it from the existence of directories:

- **Migrated**: `~/.finfocus` exists.
- **Legacy**: `~/.finfocus` exists AND `~/.finfocus` does not.
- **Fresh**: Neither exists.

## Environment Variables Mapping

| New Variable | Legacy Variable | Purpose |
|--------------|-----------------|---------|
| `FINFOCUS_HOME` | `FINFOCUS_HOME` | Override config directory |
| `FINFOCUS_LOG_LEVEL` | `FINFOCUS_LOG_LEVEL` | Logging verbosity |
| `FINFOCUS_API_KEY` | `FINFOCUS_API_KEY` | SaaS integration (future) |
