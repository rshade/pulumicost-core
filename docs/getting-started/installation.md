---
layout: default
title: Installation Guide
description: Detailed installation instructions for PulumiCost
---

Detailed steps to install PulumiCost on your system.

## Prerequisites

- Pulumi CLI installed
- Go 1.25.5+ (if building from source)
- Git (if building from source)
- 5-10 minutes

## Installation Methods

### Option 1: Build from Source (Recommended)

**Step 1: Clone the repository**

```bash
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core
```

**Step 2: Build**

```bash
make build
```

Binary will be created at: `bin/pulumicost`

**Step 3: Add to PATH (optional)**

```bash
# macOS/Linux
export PATH="$PWD/bin:$PATH"

# Or copy to system path
sudo cp bin/pulumicost /usr/local/bin/
```

**Step 4: Verify**

```bash
pulumicost --version
pulumicost --help
```

### Option 2: Download Prebuilt Binary (Coming Soon)

Prebuilt binaries for Linux, macOS, and Windows will be available in GitHub releases.

```bash
# Download
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-arm64 -o pulumicost

# Make executable
chmod +x pulumicost

# Move to PATH
sudo mv pulumicost /usr/local/bin/
```

### Option 3: Docker

```bash
docker run ghcr.io/rshade/pulumicost:latest cost projected --help
```

## Verification

```bash
# Check version
pulumicost --version

# Test with example plan
pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json
```

## Next Steps

- [Quick Start Guide](quickstart.md)
- [User Guide](../guides/user-guide.md)
- [Plugin Setup](../plugins/vantage/setup.md)
