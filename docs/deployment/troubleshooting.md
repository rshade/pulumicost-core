---
title: Deployment Troubleshooting
layout: default
---

If you encounter issues deploying FinFocus, please refer to our comprehensive troubleshooting guide.

- [Troubleshooting Guide](../support/troubleshooting.md)

## Common Deployment Issues

### Docker Permission Denied

If you see permission errors when mounting volumes in Docker, ensure the host directory is owned
by the user running the container or use `chmod` to grant access.

See [Docker Guide](docker.md#troubleshooting) for specific commands.

### CI/CD Pipeline Failures

If FinFocus fails in CI/CD:

1. Enable debug logging: `FINFOCUS_LOG_LEVEL=debug`
2. Check that the `finfocus` binary is in the `PATH`
3. Verify that plugins are correctly installed or cached

For more help, see the [Support](../support/support-channels.md) options.
