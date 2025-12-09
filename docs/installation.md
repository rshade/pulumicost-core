# Installation Guide

Complete installation instructions for PulumiCost Core across different platforms and deployment scenarios.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
- [Platform-Specific Instructions](#platform-specific-instructions)
- [Docker Installation](#docker-installation)
- [Build from Source](#build-from-source)
- [Plugin Installation](#plugin-installation)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Architecture**: x86_64 (amd64) or ARM64
- **Memory**: Minimum 256MB RAM
- **Storage**: 50MB for binary, additional space for plugins and specs

### Software Dependencies

- **Pulumi CLI**: Required for generating plan JSON files
  - Install from: https://www.pulumi.com/docs/get-started/install/
  - Minimum version: 3.0.0

## Installation Methods

### 1. Download Pre-built Binaries (Recommended)

Download the latest release for your platform:

#### Linux (x86_64)

```bash
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/
```

#### Linux (ARM64)

```bash
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-arm64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/
```

#### macOS (Intel)

```bash
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-amd64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/
```

#### macOS (Apple Silicon)

```bash
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-arm64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/
```

#### Windows (x86_64)

```powershell
# Download using PowerShell
Invoke-WebRequest -Uri "https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-windows-amd64.exe" -OutFile "pulumicost.exe"

# Add to PATH (optional)
Move-Item pulumicost.exe $env:USERPROFILE\bin\
$env:PATH += ";$env:USERPROFILE\bin"
```

### 2. Package Managers

#### Homebrew (macOS/Linux)

```bash
# Coming soon - not yet available
brew install pulumicost
```

#### Chocolatey (Windows)

```powershell
# Coming soon - not yet available
choco install pulumicost
```

#### APT (Ubuntu/Debian)

```bash
# Coming soon - not yet available
sudo apt update
sudo apt install pulumicost
```

## Platform-Specific Instructions

### Linux

#### Ubuntu/Debian

```bash
# Install dependencies
sudo apt update
sudo apt install curl

# Download and install
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/

# Verify installation
pulumicost --version
```

#### CentOS/RHEL/Fedora

```bash
# Install dependencies
sudo yum install curl  # or dnf install curl

# Download and install
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/

# Verify installation
pulumicost --version
```

#### Alpine Linux

```bash
# Install dependencies
apk add curl

# Download and install
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost
mv pulumicost /usr/local/bin/

# Verify installation
pulumicost --version
```

### macOS

#### Using curl (Recommended)

```bash
# Download for your architecture
# Intel Macs
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-amd64 -o pulumicost

# Apple Silicon Macs
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-darwin-arm64 -o pulumicost

# Install
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/

# Verify installation
pulumicost --version
```

#### Bypass Gatekeeper (if needed)

```bash
# If macOS blocks the binary due to security settings
sudo spctl --add /usr/local/bin/pulumicost
sudo xattr -dr com.apple.quarantine /usr/local/bin/pulumicost
```

### Windows

#### PowerShell Installation

```powershell
# Create bin directory
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\bin

# Download binary
Invoke-WebRequest -Uri "https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-windows-amd64.exe" -OutFile "$env:USERPROFILE\bin\pulumicost.exe"

# Add to PATH (persistent)
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\bin", [System.EnvironmentVariableTarget]::User)

# Refresh current session
$env:PATH += ";$env:USERPROFILE\bin"

# Verify installation
pulumicost --version
```

#### Windows Subsystem for Linux (WSL)

```bash
# Use Linux installation instructions inside WSL
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o pulumicost
chmod +x pulumicost
sudo mv pulumicost /usr/local/bin/
```

## Docker Installation

### Using Official Docker Image

```bash
# Pull the latest image
docker pull rshade/pulumicost-core:latest

# Run with volume mounts
docker run --rm -v $(pwd):/workspace rshade/pulumicost-core:latest \
  cost projected --pulumi-json /workspace/plan.json

# Create an alias for easier usage
alias pulumicost='docker run --rm -v $(pwd):/workspace rshade/pulumicost-core:latest'
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  pulumicost:
    image: rshade/pulumicost-core:latest
    volumes:
      - ./plans:/workspace/plans:ro
      - ./specs:/workspace/specs:ro
      - ~/.pulumicost:/root/.pulumicost
    command: ['cost', 'projected', '--pulumi-json', '/workspace/plans/plan.json']
```

### Building Docker Image

```bash
# Clone repository
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core

# Build image
docker build -t pulumicost-core .

# Run
docker run --rm pulumicost-core --version
```

## Build from Source

### Prerequisites for Building

- **Go**: Version 1.25.5 or later
  - Install from: https://golang.org/dl/
- **Git**: For cloning the repository
- **Make**: For using build scripts (optional)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core

# Build using Make (recommended)
make build

# Or build directly with Go
go build -o bin/pulumicost ./cmd/pulumicost

# Install to system
sudo cp bin/pulumicost /usr/local/bin/

# Verify installation
pulumicost --version
```

### Development Build

```bash
# Clone and setup for development
git clone https://github.com/rshade/pulumicost-core
cd pulumicost-core

# Install dependencies
go mod download

# Run tests
make test

# Build and run
make dev

# Or run directly
go run ./cmd/pulumicost --help
```

### Cross-compilation

```bash
# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o pulumicost-linux-amd64 ./cmd/pulumicost
GOOS=darwin GOARCH=arm64 go build -o pulumicost-darwin-arm64 ./cmd/pulumicost
GOOS=windows GOARCH=amd64 go build -o pulumicost-windows-amd64.exe ./cmd/pulumicost
```

## Plugin Installation

### Plugin Directory Setup

Create the plugin directory structure:

```bash
# Create plugin directory
mkdir -p ~/.pulumicost/plugins

# Create specs directory
mkdir -p ~/.pulumicost/specs
```

### Installing Plugins

#### Kubecost Plugin Example

```bash
# Download plugin binary (example)
curl -L https://github.com/rshade/pulumicost-plugin-kubecost/releases/latest/download/pulumicost-kubecost-linux-amd64 \
  -o ~/.pulumicost/plugins/kubecost/1.0.0/pulumicost-kubecost

# Make executable
chmod +x ~/.pulumicost/plugins/kubecost/1.0.0/pulumicost-kubecost

# Create plugin manifest (optional)
cat > ~/.pulumicost/plugins/kubecost/1.0.0/plugin.manifest.json << EOF
{
  "name": "kubecost",
  "version": "1.0.0",
  "binary": "pulumicost-kubecost",
  "supports": ["actual_cost"]
}
EOF
```

#### Plugin Directory Structure

```
~/.pulumicost/plugins/
├── kubecost/
│   └── 1.0.0/
│       ├── pulumicost-kubecost          # Plugin binary
│       └── plugin.manifest.json        # Optional manifest
├── aws-plugin/
│   └── 0.1.0/
│       └── pulumicost-aws
└── azure-plugin/
    └── 0.2.0/
        └── pulumicost-azure
```

### Installing Pricing Specs

Create local pricing specifications:

```bash
# Example AWS EC2 pricing spec
cat > ~/.pulumicost/specs/aws-ec2-t3-micro.yaml << EOF
provider: aws
service: ec2
sku: t3.micro
currency: USD
pricing:
  instanceType: t3.micro
  onDemandHourly: 0.0104
  monthlyEstimate: 7.59
  vcpu: 2
  memory: 1
metadata:
  region: us-west-2
  operatingSystem: linux
EOF
```

## Verification

### Basic Verification

```bash
# Check version
pulumicost --version

# Show help
pulumicost --help

# List available commands
pulumicost help
```

### Plugin Verification

```bash
# List installed plugins
pulumicost plugin list

# Validate plugin installation
pulumicost plugin validate
```

### Test with Example

```bash
# Test with provided example
cd pulumicost-core  # if you cloned the repo
pulumicost cost projected --pulumi-json examples/plans/aws-simple-plan.json

# Expected output should show resource cost estimates
```

### Integration Test

```bash
# Test complete workflow
cd your-pulumi-project
pulumi preview --json > plan.json
pulumicost cost projected --pulumi-json plan.json

# Should display cost estimates for your resources
```

## Troubleshooting

### Common Issues

#### Binary Not Found

```bash
# Check if binary is in PATH
which pulumicost

# If not found, add directory to PATH
export PATH=$PATH:/usr/local/bin

# Or use full path
/usr/local/bin/pulumicost --version
```

#### Permission Denied (Linux/macOS)

```bash
# Make binary executable
chmod +x /usr/local/bin/pulumicost

# Check file permissions
ls -la /usr/local/bin/pulumicost
```

#### macOS Security Warning

```bash
# Allow the binary to run
sudo spctl --add /usr/local/bin/pulumicost
sudo xattr -dr com.apple.quarantine /usr/local/bin/pulumicost
```

#### Plugin Not Found

```bash
# Check plugin directory structure
ls -la ~/.pulumicost/plugins/

# Verify plugin permissions
chmod +x ~/.pulumicost/plugins/*/*/pulumicost-*
```

#### Network Issues

```bash
# Use proxy if needed
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# Or download manually and transfer
# Download on another machine and copy via SCP/USB
```

### Getting Help

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review the [GitHub Issues](https://github.com/rshade/pulumicost-core/issues)
3. Join the community discussion
4. File a bug report with:
   - Operating system and version
   - PulumiCost version (`pulumicost --version`)
   - Complete error message
   - Steps to reproduce

## Next Steps

After installation:

1. Read the [User Guide](user-guide.md) for detailed usage instructions
2. Review [Cost Calculations](cost-calculations.md) to understand methodologies
3. Set up [Plugin System](plugin-system.md) for actual cost tracking
4. Check out the [Examples](../examples/) directory

## Version Management

### Updating PulumiCost

```bash
# Check current version
pulumicost --version

# Download and replace binary with new version
curl -L https://github.com/rshade/pulumicost-core/releases/latest/download/pulumicost-linux-amd64 -o /tmp/pulumicost
chmod +x /tmp/pulumicost
sudo mv /tmp/pulumicost /usr/local/bin/pulumicost

# Verify update
pulumicost --version
```

### Multiple Versions

```bash
# Install specific version
curl -L https://github.com/rshade/pulumicost-core/releases/download/v0.1.0/pulumicost-linux-amd64 -o pulumicost-v0.1.0
chmod +x pulumicost-v0.1.0

# Use specific version
./pulumicost-v0.1.0 --version
```
