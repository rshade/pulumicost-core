# Debugging Summary: Analyzer E2E Tests

**Status**: RESOLVED (2025-12-09)

This document contains historical debugging notes. For current integration
guidance, see:

- **User Guide**: `/docs/analyzer-integration.md`
- **Test Notes**: `integration_testing_notes.md` (this directory)

## Summary of Solution

The analyzer integration works with these requirements:

1. **Binary naming**: `pulumi-analyzer-policy-pulumicost`
2. **PulumiPolicy.yaml**: Required in policy pack directory
3. **--policy-pack flag**: Required to activate the analyzer
4. **PATH**: Binary must be on PATH or in policy pack directory

## Key Technical Findings

### plugins.analyzers Does NOT Load Analyzers

The `plugins.analyzers` section in Pulumi.yaml only provides path hints for
plugin discovery. It does NOT automatically load or invoke the analyzer.

From Pulumi source (`sdk/go/common/resource/plugin/host.go`):

```go
for _, analyzerOpts := range plugins.Analyzers {
    info, err := parsePluginOpts(ctx.Root, analyzerOpts, apitype.AnalyzerPlugin)
    // This only adds to projectPlugins for path lookup
}
```

### Policy Pack Loading Requires PulumiPolicy.yaml

From Pulumi source (`sdk/go/common/resource/plugin/analyzer_plugin.go`):

```go
func NewPolicyAnalyzer(...) (Analyzer, error) {
    projPath := filepath.Join(policyPackPath, "PulumiPolicy.yaml")
    proj, err := workspace.LoadPolicyPack(projPath)
    // ...
}
```

### Runtime-Based Binary Naming

For custom runtimes without a language plugin, Pulumi looks for:
`pulumi-analyzer-policy-<runtime>`

Where `<runtime>` comes from PulumiPolicy.yaml.

## Verified Working Test (2025-12-09)

```bash
# Setup
mkdir -p /tmp/policy-pack
cp bin/pulumicost /tmp/policy-pack/pulumi-analyzer-policy-pulumicost
cat > /tmp/policy-pack/PulumiPolicy.yaml << 'EOF'
runtime: pulumicost
name: pulumicost
version: 0.0.0-dev
EOF
export PATH="/tmp/policy-pack:$PATH"

# Run
pulumi preview --policy-pack /tmp/policy-pack
```

Output confirmed:

```text
Policies:
    pulumicost@v0.0.0-dev (local: /tmp/policy-pack)
        - [advisory] cost-estimate (aws:ec2/instance:Instance: test-instance)
          Estimated Monthly Cost: $7.59 USD (source: pulumicost-plugin-aws-public)
        - [advisory] stack-cost-summary (pulumi:pulumi:Stack: ...)
          Total Estimated Monthly Cost: $7.59 USD (1 resources analyzed)
```
