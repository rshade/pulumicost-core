---
layout: default
title: Frequently Asked Questions
description: Common questions and answers about PulumiCost
---

# Frequently Asked Questions

## Installation & Setup

**Q: How do I install PulumiCost?**

A: See the [Installation Guide](../getting-started/installation.md).

**Q: Does PulumiCost require a specific version of Pulumi?**

A: PulumiCost works with Pulumi 3.0+. We recommend the latest version.

**Q: Can I use PulumiCost with my cloud provider?**

A: For projected costs, yes - works with all Pulumi-supported clouds.
For actual costs, it depends on available plugins (currently Vantage).

## Usage

**Q: Why are some resources showing $0 cost?**

A: Some resources don't have pricing data available. This is normal for:
- S3 buckets (storage costs apply only if data exists)
- Databases (pricing depends on actual usage)
- VPCs, subnets (no direct costs)

**Q: How accurate are the projected costs?**

A: 95%+ accurate for most resources. Actual costs depend on:
- Real usage patterns
- Discounts and reserved instances
- Data transfer costs

**Q: Can I filter by custom tags?**

A: Yes! Use `--filter "tag:key=value"`. See [User Guide](../guides/user-guide.md).

## Plugins

**Q: What plugins are available?**

A: Currently: Vantage (in progress)
Soon: Kubecost (planned)
Future: Flexera, Cloudability

**Q: Do I need a plugin?**

A: For projected costs: No, uses local specs
For actual costs: Yes, need a plugin

**Q: How do I install a plugin?**

A: See [Plugin Documentation](../plugins/).

## Troubleshooting

**Q: "Plugin not found" error**

A: Install the plugin first. See [Plugin Setup](../plugins/vantage/setup.md).

**Q: "No cost data available"**

A: Some resources don't have pricing. Check:
1. Is plugin/spec configured?
2. Are resource types supported?
3. Does plugin have credentials?

**Q: How do I reset configuration?**

A: Delete config directory: `rm -rf ~/.pulumicost`

## Data & Privacy

**Q: Does PulumiCost send my data anywhere?**

A: Only to plugins you configure (e.g., Vantage).
Local specs don't send any data.

**Q: Is my infrastructure data secure?**

A: Pulumi JSON contains resource details - treat as sensitive.
PulumiCost is a local CLI tool - data stays on your machine.

## Performance

**Q: How long does cost calculation take?**

A: <1 second for local specs
1-5 seconds with plugins (depends on API response)

**Q: Can I use PulumiCost with large infrastructure (1000+ resources)?**

A: Yes. Use `--output ndjson` for streaming.

## Support

**Q: Where can I get help?**

A: 
- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Issues](https://github.com/rshade/pulumicost-core/issues)
- [GitHub Discussions](https://github.com/rshade/pulumicost-core/discussions)

**Q: How do I report a bug?**

A: [Open a GitHub Issue](https://github.com/rshade/pulumicost-core/issues/new)

**Q: Can I contribute?**

A: Yes! See [Contributing Guide](contributing.md).

