# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### BREAKING CHANGES

- **Removed encryption functionality from config package**: The built-in encryption system using PBKDF2 has been completely removed due to security concerns about weak key derivation. Users should now store sensitive values (API keys, credentials) as environment variables instead of in configuration files. This is the industry-standard approach for CLI tools and follows best practices for secret management.
  - Removed `EncryptValue()` and `DecryptValue()` methods from Config
  - Removed `--encrypt` flag from `pulumicost config set` command
  - Removed `--decrypt` flag from `pulumicost config get` command
  - Removed all encryption-related infrastructure (deriveKey, master key management)

  **Migration Guide**:
  - Remove any encrypted values from your `~/.pulumicost/config.yaml`
  - Store sensitive values as environment variables using the pattern: `PULUMICOST_PLUGIN_<PLUGIN_NAME>_<KEY_NAME>`
  - Example: `export PULUMICOST_PLUGIN_AWS_SECRET_KEY="your-secret"`
  - Environment variables automatically override config file values

### Changed

- Updated CLI command documentation to recommend environment variables for sensitive data
- Updated README with comprehensive configuration and environment variable documentation
- Simplified config package by removing unused encryption dependencies

### Removed

- PBKDF2-based encryption key derivation (security vulnerability)
- AES-256-GCM encryption for configuration values
- Master key file creation and management
- Encryption-related tests and validation

## [0.1.0] - 2025-01-14

### Added

- Initial release of PulumiCost Core CLI
- Projected cost calculation from Pulumi plans
- Actual cost queries with time ranges and filtering
- Cross-provider cost aggregation
- Plugin-based architecture for extensibility
- Configuration management system
- Multiple output formats (table, JSON, NDJSON)
- Resource grouping and filtering capabilities
- Comprehensive testing framework
