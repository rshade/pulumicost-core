# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CLI Package Overview

The `internal/cli` package implements the Cobra-based command-line interface for PulumiCost. It follows a hierarchical command structure with two main command groups: `cost` (for cost calculations) and `plugin` (for plugin management).

## Command Architecture

### Command Hierarchy

```text
pulumicost
├── cost
│   ├── projected  # Calculate estimated costs from Pulumi plan
│   └── actual     # Fetch historical costs from cloud providers
└── plugin
    ├── list       # List installed plugins
    └── validate   # Validate plugin installations
```

### Command Implementation Pattern

Each command follows this consistent pattern:

1. **Constructor Function**: `NewXxxCmd()` returns a `*cobra.Command`
2. **Flag Variables**: Declared as local variables in the constructor
3. **RunE Function**: Contains the main command logic
4. **Flag Registration**: Uses `cmd.Flags().StringVar()` for configuration
5. **Required Flags**: Marked with `cmd.MarkFlagRequired()`

### Core Dependencies Flow

```text
CLI Command → Ingest (Pulumi JSON) → Registry (Plugin Discovery) → Engine (Cost Calculation) → Output
```

Key imports:

- `internal/ingest`: Loads and maps Pulumi resources
- `internal/registry`: Discovers and opens plugin connections
- `internal/engine`: Orchestrates cost calculation
- `internal/config`: Manages configuration paths
- `internal/spec`: Loads local pricing specifications

## Testing Commands

Run tests for the CLI package:

```bash
go test ./internal/cli/...
go test -v ./internal/cli/... -run TestCLIIntegration  # Integration tests
go test -race ./internal/cli/...                        # With race detection
```

## Common Development Tasks

### Adding a New Command

1. Create new file: `internal/cli/your_command.go`
2. Follow the constructor pattern with `NewYourCmd() *cobra.Command`
3. Use `RunE` for error handling (not `Run`)
4. Always use `cmd *cobra.Command` as first parameter when command needs access to output streams
5. Add command to parent in `root.go` or appropriate parent command

### Resource Processing Pipeline

All cost commands follow this pipeline:

1. Load Pulumi plan: `ingest.LoadPulumiPlan(planPath)`
2. Map resources: `ingest.MapResources(pulumiResources)`
3. Open plugins: `registry.Open(ctx, adapter)`
4. Calculate costs: `engine.GetProjectedCost()` or `engine.GetActualCost()`
5. Render output: `engine.RenderResults(outputFormat, results)`

### Plugin Management Pattern

Plugin commands interact with the registry:

1. Check plugin directory exists: `os.Stat(cfg.PluginDir)`
2. List plugins: `registry.ListPlugins()`
3. Validate plugins: Custom validation logic in `ValidatePlugin()`
4. Display results: Use `tabwriter` for formatted output

## Critical Patterns and Conventions

### Error Handling

- Use `RunE` not `Run` for commands
- Wrap errors with context: `fmt.Errorf("action: %w", err)`
- Return early on errors, don't continue processing

### Resource Cleanup

- Always defer cleanup functions immediately after obtaining resources:

  ```go
  clients, cleanup, err := reg.Open(ctx, adapter)
  if err != nil {
      return fmt.Errorf("opening plugins: %w", err)
  }
  defer cleanup()
  ```

### Command Output

- Use `cmd.Printf()` and `cmd.Println()` for command output (not `fmt.Printf`)
- Use `tabwriter` for table formatting with consistent padding
- Support multiple output formats: table, json, ndjson

### Date/Time Handling

- Support multiple formats: "2006-01-02", RFC3339
- Default `--to` to `time.Now()` if not provided
- Validate date ranges (to must be after from)

### Testing Patterns

- Use `cli_test` package for black-box testing
- Create temporary directories with `t.TempDir()`
- Mock Pulumi plans as JSON structures
- Test both success and error cases
- Use `cmd.SetOut()` and `cmd.SetErr()` to capture output

### Flag Conventions

- Required flags: `--pulumi-json` for all cost commands
- Optional adapter override: `--adapter`
- Output format: `--output` (table|json|ndjson)
- Global debug flag: `--debug` on root command

## Common Gotchas

1. **Plugin Directory**: Commands gracefully handle missing plugin directory
2. **Empty Plugin List**: Don't error when no plugins installed, show informative message
3. **Date Parsing**: Try multiple formats before failing
4. **Executable Check**: Use `info.Mode()&0111` to check if file is executable
5. **Manifest Validation**: Optional but validated if present
6. **Output Streams**: Always use `cmd.Printf()` not `fmt.Printf()` for testability

## Integration Points

- **Config Package**: Uses `config.New()` for default paths (~/.pulumicost/specs, ~/.pulumicost/plugins)
- **Registry Package**: `registry.NewDefault()` creates standard plugin registry
- **Engine Package**: Handles both projected and actual cost calculations
- **Proto Adapter**: Real protobuf definitions from pulumicost-spec repository

