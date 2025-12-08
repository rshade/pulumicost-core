# pulumicost-core Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-22

## Active Technologies

- Local Pulumi state (ephemeral), no persistent DB. (010-e2e-cost-testing)
- N/A (Stateless operation) (008-analyzer-plugin)

- Go 1.25.5
- `github.com/rshade/pulumicost-spec`
- `google.golang.org/grpc` (002-implement-supports-handler)

## Project Structure

```text
cmd/
internal/
pkg/
test/
testdata/
```

## Commands

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Run
make run
```

## Code Style

Go 1.25.5: Follow standard conventions

<!-- MANUAL ADDITIONS START -->

## Workflow Restrictions

- **NEVER COMMIT**: Do not execute `git commit`. Always stop after `git add` and ask the user to review/commit.
<!-- MANUAL ADDITIONS END -->

## Integration & Testing

- **E2E Testing**: Uses `pulumi preview --json` against real AWS infrastructure (ephemeral stacks).
- **Pulumi Plan JSON**: The structure of `pulumi preview --json` output nests resource state under `newState` (for creates/updates). Ingest logic MUST check `newState` to correctly extract `Inputs` and `Type`.
- **Resource Type Compatibility**: Plugins must handle Pulumi-style resource types (e.g., `aws:ec2/instance:Instance`) or Core must normalize them. Currently, plugins are expected to handle the mapping.
- **Property Extraction**: Core (`adapter.go`) relies on populated `Inputs` to extract SKU and Region. If `Inputs` are empty (due to ingest issues), pricing lookup fails.

## Recent Changes

- 010-e2e-cost-testing: Fixed E2E test failure by parsing `newState` in Pulumi plan JSON.
- 010-e2e-cost-testing: Patched `aws-public` plugin to support `aws:ec2/instance:Instance` resource type.
- 010-e2e-cost-testing: Verified cost calculation accuracy ($7.59/month for t3.micro).
- 010-e2e-cost-testing: Added Go 1.25
- 010-e2e-cost-testing: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
- 010-e2e-cost-testing: Added Go 1.25
- 008-analyzer-plugin: Added Go 1.25 + `github.com/pulumi/pulumi/sdk/v3` (for `pulumirpc`), `google.golang.org/grpc`, `github.com/spf13/cobra`
- 008-analyzer-plugin: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]

## Session Analysis - Recommended Updates

Based on recent development sessions, consider adding:

### Go Version Management

- **Version Consistency**: When updating Go versions, update both `go.mod` and ALL markdown files simultaneously
- **Search Pattern**: Use `grep "Go.*1\." --include="*.md"` to find all version references in documentation
- **Files to Check**: go.mod, all .md files in docs/, specs/, examples/, and root level documentation
- **Docker Images**: Update Docker base images (e.g., `golang:1.24` → `golang:1.25.5`) in documentation examples

### Systematic Version Updates

- **Process**: 1) Update go.mod first, 2) Find all references with grep, 3) Update each file systematically, 4) Verify with final grep search
- **Common Patterns**: Update both specific versions (1.24.10 → 1.25.5) and minimum requirements (Go 1.24+ → Go 1.25.5+)
- **CI Workflows**: Update GitHub Actions go-version parameters in documentation examples

This ensures complete version consistency across the entire codebase and documentation.

## AI Agent File Maintenance

This file (GEMINI.md) provides guidance for Gemini AI assistants. To maintain its effectiveness:

### Update Requirements:

- **Review regularly** when significant codebase changes occur
- **Update version information** immediately when technologies change
- **Document new active technologies** as they are introduced
- **Update workflow restrictions** if development processes change
- **Maintain current project structure** documentation
- **Keep build and test commands** accurate and functional

### When to Update:

- New technologies are adopted
- Build processes change
- Project structure evolves
- Workflow restrictions change
- New dependencies are added
- Testing frameworks change

### Integration with GitHub Copilot:

- This file is automatically read by GitHub Copilot via `.github/instructions/ai-agent-files.instructions.md`
- Use it as reference for Gemini AI assistants
- Follow the documented workflow restrictions
- Keep information current for consistent AI assistance

### Maintenance Checklist:

- [ ] Active technologies list is current
- [ ] Project structure reflects reality
- [ ] Build commands work as documented
- [ ] Workflow restrictions are accurate
- [ ] Integration and testing information is up to date
- [ ] Recent changes section is maintained
