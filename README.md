# pulumicost-core
PulumiCost Core is the CLI and plugin host for the PulumiCost ecosystem.
It discovers and runs cost data plugins, parses Pulumi plan output, and calculates projected and actual costs for your infrastructure.

This is the Pulumi-agnostic core — it does not require a Pulumi fork, and works against JSON output from pulumi preview --json or pulumi stack export.

## Features
Plugin-based architecture
Supports vendor and custom plugins via the PulumiCost Spec costsource.proto gRPC interface.

## Projected & Actual costs

Projected: Uses PricingSpec from local files or plugin APIs.

Actual: Pulls historical costs from vendor APIs (e.g., Kubecost, Flexera, Cloudability).

Multiple output formats
Table, JSON, or newline-delimited JSON (ndjson) for CI pipelines.

No CUR parsing required
Plugins work with preprocessed or vendor-provided cost APIs, not raw cloud billing exports.

## Installation

```bash
# Clone & build
git clone https://github.com/yourorg/pulumicost-core
cd pulumicost-core
make build

# Binary will be in ./bin
./bin/pulumicost --help
```

## Directory Layout

```bash
cmd/pulumicost/        # CLI entrypoint
internal/cli/          # Cobra commands
internal/pluginhost/   # gRPC plugin launcher
internal/registry/     # Plugin registry & manifest parsing
internal/engine/       # Orchestration logic for actual/projected costs
internal/spec/         # Spec loader & validation
internal/ingest/       # Pulumi plan JSON parsing
pkg/version/           # CLI version info
examples/              # Sample Pulumi plans & specs
testdata/              # Unit test fixtures
```

## Proposed Layout

```bash
pulumicost-core/
├─ README.md
├─ LICENSE
├─ go.mod
├─ Makefile
├─ cmd/
│  └─ pulumicost/
│     └─ main.go
├─ internal/
│  ├─ cli/
│  │  ├─ root.go
│  │  ├─ cost_actual.go
│  │  ├─ cost_projected.go
│  │  ├─ plugin_validate.go
│  │  └─ plugin_list.go
│  ├─ pluginhost/
│  │  ├─ host.go
│  │  ├─ process.go
│  │  └─ stdio.go
│  ├─ registry/
│  │  ├─ registry.go
│  │  └─ manifest.go
│  ├─ engine/
│  │  ├─ engine.go
│  │  ├─ project.go
│  │  └─ types.go
│  ├─ spec/
│  │  ├─ loader.go
│  │  └─ validate.go
│  ├─ ingest/
│  │  ├─ pulumi_plan.go
│  │  └─ map_resource.go
│  ├─ config/
│  │  └─ config.go
│  └─ util/
│     ├─ time.go
│     └─ json.go
├─ pkg/
│  └─ version/
│     └─ version.go
├─ examples/
│  ├─ plans/                       # example outputs of `pulumi preview --json`
│  │  └─ aws-simple-plan.json
│  └─ specs/
│     └─ aws-ec2-t3-micro.yaml
└─ testdata/
   ├─ plan.json
   └─ pricing_spec.yaml
```

## Key Responsibilities

Key responsibilities
* CLI: pulumicost cost actual|projected, pulumicost plugin validate|list
* Plugin host: discover, launch, and talk to gRPC plugins (Kubecost, Cloudability, etc.)
* Registry: read ~/.pulumicost/plugins/<name>/<version>/<binary> * plugin.manifest.json
* Spec: load/validate PricingSpec YAML/JSON (optional override)
* Ingest: parse pulumi preview --json / stack export → []ResourceDescriptor
* Engine: orchestrate “for each resource → call plugin(s) → aggregate result”

## CLI UX (MVP)

```bash
# Projected cost from spec/vendor plugins for a Pulumi plan
pulumicost cost projected --pulumi-json ./plan.json --spec-dir ./specs --output table

# Actual cost for a time window
pulumicost cost actual --pulumi-json ./plan.json --from 2025-07-01 --to 2025-07-31 --output json

# Validate installed plugins
pulumicost plugin validate

# List plugins and capabilities
pulumicost plugin list
```

## Usage

1. Generate a Pulumi plan

```bash
pulumi preview --json > plan.json
```

2. Compute projected costs

```bash
pulumicost cost projected \
  --pulumi-json ./plan.json \
  --spec-dir ./specs \
  --output table
```

Example output:

```bash
RESOURCE                          ADAPTER     MONTHLY   CURRENCY  NOTES
aws:ec2/instance:Instance         aws-spec    7.50      USD       On-demand Linux t3.micro
aws:s3/bucket:Bucket               aws-spec    2.30      USD       Standard storage 100GB
```

3. Fetch actual historical costs

```bash
pulumicost cost actual \
  --pulumi-json ./plan.json \
  --from 2025-07-01 \
  --to 2025-07-31 \
  --adapter kubecost \
  --output json
```

## Plugins

PulumiCost plugins are standalone binaries implementing the CostSource gRPC service.
They are installed under:

```bash
~/.pulumicost/plugins/<name>/<version>/<binary>
Example:

```bash
~/.pulumicost/plugins/kubecost/1.0.0/pulumicost-kubecost
```

Listing plugins

```bash
pulumicost plugin list
```

Validating plugins

```bash
pulumicost plugin validate
```

## Configuration
CLI flags:

--pulumi-json: Path to pulumi preview --json or stack export output

--spec-dir: Directory with PricingSpec YAML/JSON files

--from, --to: Date range for actual cost (YYYY-MM-DD or RFC3339)

--adapter: Restrict to a specific plugin

--output: table, json, or ndjson

## Development
Prerequisites
Go 1.22+

pulumicost-spec (for proto & types)

cobra

gRPC

### Build

```bash
make build
```

### Run

```bash
bin/pulumicost --help
```

### Test

```bash
make test
```

## License
Apache-2.0 (recommended for core + spec)

