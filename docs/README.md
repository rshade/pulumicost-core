# FinFocus Documentation

Welcome to the FinFocus documentation hub. Whether you're a user, engineer, architect,
or business stakeholder, you'll find comprehensive guides to help you succeed.

## Quick Navigation

### I want to... (pick your path)

<table>
  <tr>
    <th>👤 End User</th>
    <th>🛠️ Engineer/Developer</th>
    <th>🏗️ Software Architect</th>
    <th>💼 Business/CEO</th>
  </tr>
  <tr>
    <td>
      <strong>See costs in 5 minutes</strong><br>
      <a href="getting-started/quickstart.md">Quickstart</a><br>
      <a href="getting-started/installation.md">Install</a><br>
      <a href="reference/cli-commands.md">CLI Commands</a><br>
      <a href="support/faq.md">FAQ</a>
    </td>
    <td>
      <strong>Build a plugin</strong><br>
      <a href="guides/developer-guide.md">Developer Guide</a><br>
      <a href="plugins/plugin-development.md">Plugin Dev</a><br>
      <a href="plugins/plugin-sdk.md">SDK Reference</a><br>
      <a href="support/contributing.md">Contributing</a>
    </td>
    <td>
      <strong>Integrate with our system</strong><br>
      <a href="guides/architect-guide.md">Architect Guide</a><br>
      <a href="architecture/system-overview.md">Architecture</a><br>
      <a href="deployment/deployment.md">Deployment</a><br>
      <a href="deployment/security.md">Security</a>
    </td>
    <td>
      <strong>Understand the value</strong><br>
      <a href="guides/business-value.md">Value Prop</a><br>
      <a href="guides/comparison.md">Competitive</a><br>
      <a href="architecture/roadmap.md">Roadmap</a><br>
      <a href="getting-started/quickstart.md">Demo (5m)</a>
    </td>
  </tr>
</table>

---

## Documentation Overview

### 📚 Comprehensive Guides

- **[User Guide](guides/user-guide.md)** - Complete guide for end users
- **[Developer Guide](guides/developer-guide.md)** - Complete guide for plugin developers
- **[Architect Guide](guides/architect-guide.md)** - Complete guide for software architects
- **[Business Value](guides/business-value.md)** - Value proposition and ROI

### 2. View Projected Costs

```bash
finfocus cost projected --pulumi-json plan.json
```

**Output:**

```text
RESOURCE                          TYPE                MONTHLY   CURRENCY
aws:ec2/instance:Instance         aws:ec2:Instance    $7.50     USD
aws:s3/bucket:Bucket              aws:s3:Bucket       $0.00     USD
aws:rds/instance:Instance         aws:rds:Instance    $0.00     USD

Total: $7.50 USD
```

### 3. Zero-Click Cost Estimation

Integrate FinFocus directly into your `pulumi preview` workflow.

```yaml
# Configure analyzer in Pulumi.yaml
plugins:
  - path: finfocus
    args: ['analyzer', 'serve']
```

Then run:

```bash
pulumi preview
```

### 4. (Optional) View Actual Costs

- **[5-Minute Quickstart](getting-started/quickstart.md)** - Get costs in 5 minutes
- **[Installation Guide](getting-started/installation.md)** - Step-by-step installation
- **[Prerequisites](getting-started/prerequisites.md)** - System requirements
- **[Examples](getting-started/examples/)** - Practical examples with Vantage, local pricing, and multiple plugins

### 🏗️ Architecture & Design

- **[System Overview](architecture/system-overview.md)** - High-level architecture with diagrams
- **[Core Concepts](architecture/core-concepts.md)** - Key concepts explained
- **[Plugin Protocol](architecture/plugin-protocol.md)** - gRPC plugin protocol specification
- **[Cost Calculation](architecture/cost-calculation.md)** - How costs are calculated and aggregated

### 🔌 Plugin Documentation

#### For Plugin Developers

- **[Plugin Development Guide](plugins/plugin-development.md)** - How to build a FinFocus plugin
- **[Plugin SDK Reference](plugins/plugin-sdk.md)** - API and SDK documentation
- **[Plugin Examples](plugins/plugin-examples.md)** - Code patterns and examples
- **[Plugin Checklist](plugins/plugin-checklist.md)** - Ensure your plugin is complete

#### Vantage Plugin (IN PROGRESS)

- **[Setup Guide](plugins/vantage/setup.md)** - Get started with Vantage
- **[Authentication](plugins/vantage/authentication.md)** - API key management
- **[Features](plugins/vantage/features.md)** - What's supported
- **[Troubleshooting](plugins/vantage/troubleshooting.md)** - Common issues

#### Future Plugins (PLANNED)

- **[Kubecost](plugins/kubecost/coming-soon.md)** - Kubernetes cost allocation
- **[Flexera](plugins/flexera/coming-soon.md)** - Multi-cloud cost management
- **[Cloudability](plugins/cloudability/coming-soon.md)** - Enterprise cost visibility

### 📖 Reference Documentation

- **[CLI Commands](reference/cli-commands.md)** - Complete command reference
- **[Configuration](reference/config-reference.md)** - Configuration options
- **[API Reference](reference/api-reference.md)** - gRPC API documentation
- **[Error Codes](reference/error-codes.md)** - Error codes and solutions

### 🚢 Deployment & Operations

- **[Installation Guide](deployment/installation.md)** - Detailed installation procedures
- **[Configuration Guide](deployment/configuration.md)** - How to configure
- **[Docker](deployment/docker.md)** - Docker deployment
- **[CI/CD Integration](deployment/cicd-integration.md)** - Pipeline integration
- **[Security](deployment/security.md)** - Security best practices
- **[Troubleshooting](deployment/troubleshooting.md)** - Operational issues

### 💬 Support & Community

- **[FAQ](support/faq.md)** - Frequently asked questions
- **[Troubleshooting Guide](support/troubleshooting.md)** - Problem-solving by symptom
- **[Contributing](support/contributing.md)** - How to contribute to FinFocus
- **[Code of Conduct](support/code-of-conduct.md)** - Community guidelines
- **[Support Channels](support/support-channels.md)** - Where to get help

---

## Key Concepts

### What is FinFocus?

**Key Features:**

- 🔍 **Pulumi Analyzer Integration**: Zero-click cost estimation during `pulumi preview`
- 📊 **Cross-Provider Aggregation**: Daily/monthly cost trends across AWS, Azure, GCP
- 🏗️ **Plugin Development Kit**: Initialize, install, update, and manage plugins
- 📊 **Projected Costs** - Estimate costs before deploying
- 💰 **Actual Costs** - See what you're actually paying
- 🔌 **Multiple Cost Sources** - Works with Vantage, local specs, and more
- 🎯 **Flexible Filtering** - Filter by resource type, tags, or custom criteria
- 📈 **Cost Aggregation** - Group costs by provider, type, date, or tags
- 📱 **Multiple Formats** - Table, JSON, or NDJSON output

### How It Works

```text
1. You define infrastructure with Pulumi
2. FinFocus reads your Pulumi definitions
3. Plugins fetch pricing and cost data
4. FinFocus calculates and displays results
```

### Plugin-Based Architecture

FinFocus uses a plugin system to support multiple cost providers:

- **Vantage** (IN PROGRESS) - Multi-cloud cost aggregation
- **Kubecost** (PLANNED) - Kubernetes cost allocation
- **Flexera** (FUTURE) - Enterprise cost management
- **Cloudability** (FUTURE) - Cloud cost visibility
- **Local Specs** (ALWAYS) - No external service needed

---

## Documentation Structure

See [plan.md](plan.md) for complete documentation architecture, maintenance strategy, and implementation timeline.

---

## Finding What You Need

### By Role

- **DevOps/Infrastructure Engineer** → Start with [Installation](getting-started/installation.md) → [Quickstart](getting-started/quickstart.md)
- **FinOps Engineer** → Start with [Business Value](guides/business-value.md) → [Architecture](architecture/system-overview.md)
- **Plugin Developer** → Start with [Developer Guide](guides/developer-guide.md) → [Plugin Development](plugins/plugin-development.md)
- **Platform Architect** → Start with [Architect Guide](guides/architect-guide.md) → [Architecture](architecture/system-overview.md)
- **Executive/CEO** → Start with [Business Value](guides/business-value.md) → [Roadmap](architecture/roadmap.md)

### By Use Case

- **I just installed FinFocus** → [Quickstart](getting-started/quickstart.md)
- **I want to integrate with Vantage** → [Vantage Setup](plugins/vantage/setup.md)
- **I'm building a custom plugin** → [Plugin Development](plugins/plugin-development.md)
- **I'm integrating with CI/CD** → [CI/CD Integration](deployment/cicd-integration.md)
- **Something's not working** → [Troubleshooting](support/troubleshooting.md)

### By Problem

- **"How do I install FinFocus?"** → [Installation Guide](getting-started/installation.md)
- **"How do I configure it?"** → [Configuration Guide](deployment/configuration.md)
- **"How do I build a plugin?"** → [Plugin Development](plugins/plugin-development.md)
- **"What's the cost calculation?"** → [Cost Calculation](architecture/cost-calculation.md)
- **"Where's the CLI reference?"** → [CLI Commands](reference/cli-commands.md)

---

## Site Map

```text
📄 Getting Started
  ├─ Quickstart (5 minutes)
  ├─ Installation
  ├─ Prerequisites
  └─ Examples
      ├─ Vantage Setup
      ├─ Local Pricing
      └─ Multi-Plugin

📘 Guides (by audience)
  ├─ User Guide (what it does)
  ├─ Developer Guide (how to extend)
  ├─ Architect Guide (how it's designed)
  └─ Business Value (why it matters)

🏗️ Architecture
  ├─ System Overview
  ├─ Core Concepts
  ├─ Plugin Protocol
  ├─ Cost Calculation
  ├─ Roadmap
  └─ Diagrams (visual)

🔌 Plugins
  ├─ Plugin Development
  ├─ SDK Reference
  ├─ Examples
  ├─ Checklist
  └─ Per-Plugin Documentation
      ├─ Vantage (IN PROGRESS)
      ├─ Kubecost (PLANNED)
      ├─ Flexera (FUTURE)
      └─ Cloudability (FUTURE)

📋 Reference
  ├─ CLI Commands
  ├─ Configuration
  ├─ API Reference
  └─ Error Codes

🚢 Deployment
  ├─ Installation
  ├─ Configuration
  ├─ Docker
  ├─ Kubernetes (FUTURE)
  ├─ CI/CD Integration
  ├─ Security
  └─ Troubleshooting

💬 Support
  ├─ FAQ
  ├─ Troubleshooting
  ├─ Contributing
  ├─ Code of Conduct
  └─ Support Channels
```

---

## Latest Updates

- **2025-10-29** - Documentation architecture planning complete
- Coming soon: Core guides and plugin documentation

---

## Contributing to Documentation

Found an error? Want to improve a guide? The documentation is open source!

See [Contributing Guide](support/contributing.md) for how to:

- Report documentation issues
- Submit improvements
- Add new examples
- Translate documentation

---

## For LLM/AI Tools

If you're an AI assistant helping someone with FinFocus, see [llms.txt](llms.txt) for a
machine-readable index of all documentation.

---

**Complete Table of Contents:** [View Full TOC](TABLE-OF-CONTENTS.md)

**Last Updated:** 2025-10-29
**Version:** 1.0 (Complete Documentation)
**Status:** ✅ Core documentation complete | 🟡 Architecture diagrams in progress
