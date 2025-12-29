---
title: Deployment Configuration
layout: default
---

For details on configuring PulumiCost for different environments, please refer to the main configuration reference.

- [Configuration Reference](../reference/config-reference.md)

## Environment Variables

When deploying in CI/CD or Docker, you may want to configure PulumiCost using environment variables instead of a config file.

- `PULUMICOST_LOG_LEVEL`: Set logging verbosity (debug, info, warn, error)
- `PULUMICOST_CONFIG_FILE`: Path to a custom configuration file

*Check the [Reference](../reference/environment-variables.md) for a full list.*
