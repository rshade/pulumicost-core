---
layout: default
title: Troubleshooting Guide
description: Common issues and solutions
---

# Troubleshooting Guide

## Installation Issues

**"pulumicost: command not found"**

Solution:
```bash
# Build from source
make build

# Add to PATH
export PATH="$PWD/bin:$PATH"

# Or verify installation
ls -la bin/pulumicost
```

**"make: command not found"**

Solution:
```bash
# Install make
# macOS
brew install make

# Ubuntu/Debian
sudo apt-get install make

# Check
make --version
```

## Cost Calculation Issues

**"No cost data available"**

Causes:
1. Plugin not installed
2. Plugin not configured
3. Resource type not supported
4. API credentials invalid

Solution:
```bash
# Check plugins
pulumicost plugin list
pulumicost plugin validate

# See plugin documentation
# docs/plugins/vantage/setup.md
```

**"Invalid date format"**

Solution:
```bash
# Use YYYY-MM-DD
pulumicost cost actual --from 2024-01-01

# Or RFC3339
pulumicost cost actual --from 2024-01-01T00:00:00Z
```

**"Filter not working"**

Solution:
```bash
# Check filter syntax
pulumicost cost actual --filter "tag:env=prod"

# Multiple conditions
pulumicost cost actual --filter "tag:env=prod AND tag:team=platform"
```

## Plugin Issues

**"Plugin validation failed"**

```bash
# Debug plugin
pulumicost plugin validate

# Check plugin binary
ls -la ~/.pulumicost/plugins/*/*/
chmod +x ~/.pulumicost/plugins/*/*/pulumicost-*
```

**"Plugin timeout"**

Solution:
1. Check network connectivity
2. Verify API credentials
3. Check plugin logs
4. Increase timeout (if available)

**"API authentication failed"**

Solution:
```bash
# Verify credentials
echo $VANTAGE_API_KEY  # For Vantage

# Or check config
cat ~/.pulumicost/config.yaml
```

## Performance Issues

**"CLI is slow"**

Solution:
```bash
# Filter to smaller dataset
pulumicost cost actual --from 2024-01-31 --to 2024-01-31

# Use NDJSON for large output
pulumicost cost actual --output ndjson
```

**"Plugin is timing out"**

Solution:
1. Check plugin is running
2. Verify network connectivity
3. Check plugin logs
4. Try again (may be API issue)

## File Issues

**"Pulumi JSON file not found"**

Solution:
```bash
# Generate plan
pulumi preview --json > plan.json

# Verify
cat plan.json | head
```

**"Permission denied"**

Solution:
```bash
# Fix permissions
chmod +x ~/.pulumicost/plugins/*/*/pulumicost-*

# Or rebuild
make build
chmod +x bin/pulumicost
```

## Configuration Issues

**"Configuration not found"**

Solution:
```bash
# Vantage requires config
# See: docs/plugins/vantage/setup.md

# Or use local specs
mkdir -p ~/.pulumicost/specs
# Add YAML spec files
```

**"Default currency error"**

Solution:
```bash
# Ensure specs have currency
cat ~/.pulumicost/specs/*.yaml | grep -i currency

# See spec documentation
# docs/deployment/configuration.md
```

## Getting More Help

1. **Check logs:**
   ```bash
   pulumicost cost actual --debug
   ```

2. **Read relevant guide:**
   - [User Guide](../guides/user-guide.md)
   - [Plugin Documentation](../plugins/)
   - [Configuration Guide](../deployment/configuration.md)

3. **Report issue:**
   - [GitHub Issues](https://github.com/rshade/pulumicost-core/issues)

