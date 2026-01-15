---
title: Security
layout: default
---

FinFocus is designed with security in mind, especially when handling sensitive cloud credentials
and cost data.

## Credential Handling

FinFocus does **not** store cloud provider credentials. It relies on the environment's configuration
(e.g., `AWS_PROFILE`, `AZURE_CREDENTIALS`) or the credentials provided to the Pulumi engine.

- **Recommendation**: Use short-lived credentials (OIDC) in CI/CD environments.
- **Plugins**: Plugins run as separate processes and inherit the environment variables of the host
  process. Ensure strict control over the environment where FinFocus runs.

## Data Privacy

FinFocus processes cost data locally or within your CI runner.

- **No SaaS Dependency**: The core engine does not send data to any external SaaS platform unless
  you configure a specific plugin (like Vantage) to do so.
- **Local Specs**: By default, pricing data is fetched from public APIs or local spec files.

## Plugin Security

Plugins are executable binaries.

- **Source Verification**: Only install plugins from trusted sources (e.g., the official `rshade`
  organization).
- **Checksums**: Future versions will enforce checksum verification for plugins.

## Container Security

Our Docker images are built using minimal base images (Alpine) and run as non-root users.

- See [Docker Guide](docker.md) for more details.

## Reporting Vulnerabilities

If you discover a security vulnerability, please do not report it via GitHub Issues. Instead, use
the [GitHub Security Advisory](https://github.com/rshade/finfocus/security/advisories)
workflow or follow the instructions in the repository's security policy.
