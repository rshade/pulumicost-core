# PulumiCost Core

**Cloud cost analysis for Pulumi infrastructure** - Calculate projected and actual infrastructure costs without modifying your Pulumi programs.

PulumiCost Core is a CLI tool that analyzes Pulumi infrastructure definitions to provide accurate cost estimates and historical cost tracking through a flexible plugin-based architecture.

## Key Features

- **📊 Projected Costs**: Estimate monthly costs before deploying infrastructure
- **💰 Actual Costs**: Track historical spending with detailed breakdowns  
- **🔌 Plugin-Based**: Extensible architecture supporting multiple cost data sources
- **📈 Advanced Analytics**: Resource grouping, filtering, and aggregation
- **📱 Multiple Formats**: Table, JSON, and NDJSON output options
- **🔍 Smart Filtering**: Filter by resource type, tags, or custom expressions
- **⏰ Time Range Queries**: Flexible date range support for cost analysis
- **🏗️ No Code Changes**: Works with existing Pulumi projects via JSON output

## Quick Start

### 1. Installation

Download the latest release or build from source:

```bash
# Download latest release (coming soon)
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost

# Or build from source
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core
make build
./bin/pulumicost --help
```

### 2. Generate Pulumi Plan

Export your infrastructure plan to JSON:

```bash
cd your-pulumi-project
pulumi preview --json > plan.json
```

### 3. Calculate Costs

**Projected Costs** - Estimate costs before deployment:
```bash
pulumicost cost projected --pulumi-json plan.json
```

**Actual Costs** - View historical spending (requires plugins):
```bash
# Last 7 days
pulumicost cost actual --pulumi-json plan.json --from 2025-01-07

# Specific date range  
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --to 2025-01-31
```

## Example Output

### Projected Cost Analysis
```bash
$ pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json

RESOURCE                          ADAPTER     MONTHLY   CURRENCY  NOTES
aws:ec2/instance:Instance         aws-spec    $7.50     USD       t3.micro Linux on-demand
aws:s3/bucket:Bucket             none        $0.00     USD       No pricing information available  
aws:rds/instance:Instance        none        $0.00     USD       No pricing information available
```

### Actual Cost Analysis
```bash
$ pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by type --output json
{
  "summary": {
    "totalMonthly": 45.67,
    "currency": "USD",
    "byProvider": {"aws": 45.67},
    "byService": {"ec2": 23.45, "s3": 12.22, "rds": 10.00}
  },
  "resources": [...]
}
```

## Core Concepts

### Resource Analysis Flow
1. **Export** - Generate Pulumi plan JSON with `pulumi preview --json`
2. **Parse** - Extract resource definitions and properties  
3. **Query** - Fetch cost data via plugins or local specifications
4. **Aggregate** - Calculate totals with grouping and filtering options
5. **Output** - Present results in table, JSON, or NDJSON format

### Plugin Architecture
PulumiCost uses plugins to fetch cost data from various sources:

- **Cost Plugins**: Query cloud provider APIs (Kubecost, Vantage, AWS Cost Explorer, etc.)
- **Spec Files**: Local YAML/JSON pricing specifications as fallback
- **Plugin Discovery**: Automatic detection from `~/.pulumicost/plugins/`

## Advanced Usage

### Resource Filtering
```bash
# Filter by resource type
pulumicost cost projected --pulumi-json plan.json --filter "type=aws:ec2/instance"

# Filter by tag
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by "tag:Environment=prod"
```

### Cost Aggregation  
```bash
# Group by provider
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by provider

# Group by resource type
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by type

# Group by date for time series
pulumicost cost actual --pulumi-json plan.json --from 2025-01-01 --group-by date
```

### Output Formats
```bash
# Table format (default)
pulumicost cost projected --pulumi-json plan.json --output table

# JSON for API integration
pulumicost cost projected --pulumi-json plan.json --output json

# NDJSON for streaming/pipeline processing
pulumicost cost projected --pulumi-json plan.json --output ndjson
```

## Plugin Management

### List Available Plugins
```bash
pulumicost plugin list
```

### Validate Plugin Installation  
```bash
pulumicost plugin validate
```

### Plugin Directory Structure
```
~/.pulumicost/plugins/
├── kubecost/
│   └── 1.0.0/
│       └── pulumicost-kubecost
├── vantage/
│   └── 1.0.0/
│       └── pulumicost-vantage
├── aws-plugin/
│   └── 0.1.0/
│       └── pulumicost-aws
```

## Documentation

Complete documentation is available in the [docs/](docs/) directory with guides for every audience:

- **👤 End Users**: [User Guide](docs/guides/user-guide.md) - How to install and use PulumiCost
- **🛠️ Engineers**: [Developer Guide](docs/guides/developer-guide.md) - How to extend and contribute
- **🏗️ Architects**: [Architect Guide](docs/guides/architect-guide.md) - System design and integration
- **💼 Business/CEO**: [Business Value](docs/guides/business-value.md) - ROI and competitive advantage

**Quick Links:**
- [🚀 5-Minute Quickstart](docs/getting-started/quickstart.md)
- [📖 Full Documentation Index](docs/README.md)
- [🔌 Available Plugins](docs/plugins/) - Vantage, Kubecost, and more
- [🛠️ Plugin Development](docs/plugins/plugin-development.md)
- [🏗️ System Architecture](docs/architecture/system-overview.md)
- [💬 FAQ & Support](docs/support/faq.md)

## Use Cases

- **💡 Pre-deployment Planning**: Estimate costs before infrastructure changes
- **📊 Cost Optimization**: Identify expensive resources and right-size instances  
- **🔍 Cost Attribution**: Track spending by team, environment, or project
- **📈 Trend Analysis**: Monitor cost changes over time
- **🚨 Budget Monitoring**: Set up alerts for cost thresholds
- **📋 Financial Reporting**: Generate cost reports for stakeholders

## Architecture

PulumiCost Core is designed as a plugin-agnostic orchestrator:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Pulumi JSON   │    │  PulumiCost     │    │    Plugins      │
│     Output      │───▶│     Core        │───▶│  (Kubecost,     │
│                 │    │                 │    │   Vantage, ...) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Cost Analysis  │
                       │   & Reporting   │
                       └─────────────────┘
```

## Contributing

We welcome contributions! See our development documentation:

- [CONTRIBUTING.md](CONTRIBUTING.md) - Development setup and guidelines  
- [CLAUDE.md](CLAUDE.md) - AI assistant development context
- [Architecture Documentation](internal/) - Internal package documentation

## License

Apache-2.0 - See [LICENSE](LICENSE) for details.

## Related Projects

- [pulumicost-spec](https://github.com/rshade/pulumicost-spec) - Protocol definitions and schemas
- [pulumicost-plugin-kubecost](https://github.com/rshade/pulumicost-plugin-kubecost) - Kubecost integration plugin
- [pulumicost-plugin-vantage](https://github.com/rshade/pulumicost-plugin-vantage) - Vantage cost intelligence plugin

---

**Getting Started**: Try the [examples](examples/) directory for sample Pulumi plans and pricing specifications.