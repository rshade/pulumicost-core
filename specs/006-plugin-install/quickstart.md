# Quickstart: Plugin Install/Update/Remove

## Install a Plugin

```bash
# Install from registry (latest version)
finfocus plugin install kubecost

# Install specific version
finfocus plugin install kubecost@v1.0.0

# Install from GitHub URL
finfocus plugin install github.com/rshade/finfocus-plugin-custom

# Install without saving to config
finfocus plugin install kubecost --no-save
```

## Update Plugins

```bash
# Update single plugin to latest
finfocus plugin update kubecost

# Update to specific version
finfocus plugin update kubecost@v2.0.0

# Update all plugins
finfocus plugin update --all

# Preview updates without installing
finfocus plugin update --all --dry-run
```

## Remove Plugins

```bash
# Remove plugin
finfocus plugin remove kubecost

# Remove all versions
finfocus plugin remove kubecost --all-versions

# Remove but keep config entry
finfocus plugin remove kubecost --keep-config
```

## List Installed Plugins

```bash
finfocus plugin list
```

## Configuration

Installed plugins are saved to `~/.finfocus/config.yaml`:

```yaml
plugins:
  - name: kubecost
    url: github.com/rshade/finfocus-plugin-kubecost
    version: v0.0.1
```

Missing configured plugins are auto-installed on startup.

## Authentication

For higher API rate limits (5000/hr vs 60/hr), set a GitHub token:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

Or use the GitHub CLI:

```bash
gh auth login
```
