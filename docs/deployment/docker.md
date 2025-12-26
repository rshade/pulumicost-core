---
title: Docker Deployment
layout: default
---

This guide describes the Docker configuration for building and running
PulumiCost in a containerized environment.

## Quick Start

### Pull from GitHub Container Registry

```bash
# Pull the latest release
docker pull ghcr.io/rshade/pulumicost-core:latest

# Pull a specific version
docker pull ghcr.io/rshade/pulumicost-core:v1.0.0
```

### Build Locally

```bash
# Build the image
docker build -f docker/Dockerfile -t pulumicost:local .

# Run the help command
docker run --rm pulumicost:local --help
```

## Usage Examples

### Basic Commands

```bash
# Show help
docker run --rm ghcr.io/rshade/pulumicost-core:latest --help

# List plugins
docker run --rm ghcr.io/rshade/pulumicost-core:latest plugin list

# Validate plugins
docker run --rm ghcr.io/rshade/pulumicost-core:latest plugin validate
```

### Cost Analysis with Volume Mounts

```bash
# Calculate projected costs from a local Pulumi plan
docker run --rm \
  -v $(pwd):/workspace \
  ghcr.io/rshade/pulumicost-core:latest \
  cost projected --pulumi-json /workspace/plan.json

# Get actual costs with configuration
docker run --rm \
  -v $(pwd):/workspace \
  -v ~/.pulumicost:/home/pulumicost/.pulumicost \
  ghcr.io/rshade/pulumicost-core:latest \
  cost actual --from 2024-01-01 --to 2024-01-31
```

### Plugin Management

```bash
# Mount plugin directory to persist plugins
docker run --rm \
  -v ~/.pulumicost/plugins:/home/pulumicost/.pulumicost/plugins \
  ghcr.io/rshade/pulumicost-core:latest \
  plugin list
```

## Image Details

- **Base Image**: Alpine Linux (latest)
- **Go Version**: 1.25.5 (golang:1.25.5-alpine)
- **User**: Non-root user `pulumicost` (UID: 1001, GID: 1001)
- **Working Directory**: `/home/pulumicost`
- **Plugin Directory**: `/home/pulumicost/.pulumicost/plugins`
- **Specs Directory**: `/home/pulumicost/.pulumicost/specs`

## Security Features

- Runs as non-root user for enhanced security
- Multi-stage build to minimize image size
- Health check included for container monitoring
- SBOM (Software Bill of Materials) generated during CI builds
- Vulnerability scanning with Trivy

## Environment Variables

The container respects the following environment variables:

- `HOME`: Set to `/home/pulumicost`
- `PATH`: Includes `/usr/local/bin` for the pulumicost binary

## Persistent Data

To persist plugins and configuration between container runs:

```bash
# Create local directories
mkdir -p ~/.pulumicost/{plugins,specs}

# Run with persistent volumes
docker run --rm \
  -v ~/.pulumicost:/home/pulumicost/.pulumicost \
  -v $(pwd):/workspace \
  ghcr.io/rshade/pulumicost-core:latest \
  cost projected --pulumi-json /workspace/plan.json
```

## Development

### Build Arguments

The Dockerfile supports the following build-time variables:

- Git version information is automatically embedded during build
- Build date and commit information included in the binary

### Health Check

The image includes a health check that runs `pulumicost --help`:

<!-- markdownlint-disable MD031 -->
{% raw %}

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' <container_id>
```

{% endraw %}
<!-- markdownlint-enable MD031 -->

## Troubleshooting

### Permission Issues

If you encounter permission issues with volume mounts:

```bash
# Ensure proper ownership of plugin directories
sudo chown -R 1001:1001 ~/.pulumicost
```

### Plugin Installation

Currently, plugins must be manually installed in the `~/.pulumicost/plugins`
directory. Future versions will include automated plugin downloading capabilities.
