---
title: Environment Variables
description: Environment variables for Pulumicost Core
layout: default
---

Pulumicost supports configuration via environment variables.

## Global

| Variable                 | Description                              | Default                      |
| ------------------------ | ---------------------------------------- | ---------------------------- |
| `PULUMICOST_LOG_LEVEL`   | Log verbosity (debug, info, warn, error) | info                         |
| `PULUMICOST_CONFIG_FILE` | Path to configuration file               | `~/.pulumicost/config.yaml`  |
| `PULUMICOST_PLUGIN_DIR`  | Directory for plugins                    | `~/.pulumicost/plugins`      |

## Plugins

| Variable                           | Description                   |
| ---------------------------------- | ----------------------------- |
| `AWS_ACCESS_KEY_ID`                | AWS Access Key for AWS plugin |
| `AWS_SECRET_ACCESS_KEY`            | AWS Secret Key for AWS plugin |
| `AZURE_CLIENT_ID`                  | Azure Client ID               |
| `GOOGLE_APPLICATION_CREDENTIALS`   | Path to GCP credentials       |

## E2E Testing

| Variable                      | Description                        |
| ----------------------------- | ---------------------------------- |
| `PULUMICOST_E2E_AWS_REGION`   | Region for AWS E2E tests           |
| `PULUMICOST_E2E_TOLERANCE`    | Cost tolerance for E2E validation  |
