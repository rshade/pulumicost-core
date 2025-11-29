# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1](https://github.com/rshade/pulumicost-core/compare/v0.1.0...v0.1.1) (2025-11-29)


### Added

* **pluginsdk:** add UnaryInterceptors support to ServeConfig ([#191](https://github.com/rshade/pulumicost-core/issues/191)) ([e05757a](https://github.com/rshade/pulumicost-core/commit/e05757ad914d0299387cb6a1377ad5d99c843653))


### Changed

* **core:** use pluginsdk from pulumicost-spec ([#189](https://github.com/rshade/pulumicost-core/issues/189)) ([23ae52e](https://github.com/rshade/pulumicost-core/commit/23ae52e4669ba900f6e829d45c63dfb3000cdee7))

## [0.1.0](https://github.com/rshade/pulumicost-core/compare/v0.0.1...v0.1.0) (2025-11-26)


### ⚠ BREAKING CHANGES

* remove encryption from config, use environment variables for secrets ([#149](https://github.com/rshade/pulumicost-core/issues/149))

### Added

* adding in testing ([#155](https://github.com/rshade/pulumicost-core/issues/155)) ([4680d9c](https://github.com/rshade/pulumicost-core/commit/4680d9c9aab57cd8df749dd6f1518805533420a6))
* **cli:** implement plugin install/update/remove commands ([#171](https://github.com/rshade/pulumicost-core/issues/171)) ([c93f761](https://github.com/rshade/pulumicost-core/commit/c93f761e5181830f5b58a6790e7241358999b43e))
* complete actual cost pipeline with cross-provider aggregation t… ([#52](https://github.com/rshade/pulumicost-core/issues/52)) ([c0b032f](https://github.com/rshade/pulumicost-core/commit/c0b032f78531a267b4db155c2f38c35f46c4c3b2))
* complete CLI skeleton implementation with missing flags and tests ([#15](https://github.com/rshade/pulumicost-core/issues/15)) ([994a859](https://github.com/rshade/pulumicost-core/commit/994a859283c1736ee204c3cce745f421ef405927)), closes [#3](https://github.com/rshade/pulumicost-core/issues/3)
* complete plugin development SDK and template system ([#54](https://github.com/rshade/pulumicost-core/issues/54)) ([bee3dec](https://github.com/rshade/pulumicost-core/commit/bee3dec866b9b7f37f686cfa2da10e2bbfa2699b))
* **engine,cli:** implement comprehensive error aggregation system ([#174](https://github.com/rshade/pulumicost-core/issues/174)) ([cc31cb5](https://github.com/rshade/pulumicost-core/commit/cc31cb54fd07d71d6df2117114a07bba200ab962))
* **engine:** implement projected cost pipeline with enhanced spec fa… ([#31](https://github.com/rshade/pulumicost-core/issues/31)) ([2408b47](https://github.com/rshade/pulumicost-core/commit/2408b472154b7b9d92ee09dcbe0fe128557da1a9))
* implement comprehensive actual cost pipeline with aggregation and filtering ([#36](https://github.com/rshade/pulumicost-core/issues/36)) ([db18307](https://github.com/rshade/pulumicost-core/commit/db18307c1ed992ee6a09417341b78bfd43b6e333))
* implement comprehensive CI/CD pipeline setup ([#20](https://github.com/rshade/pulumicost-core/issues/20)) ([71d4a70](https://github.com/rshade/pulumicost-core/commit/71d4a70a083a043529f8ee01ace28284e7a48d0b)), closes [#11](https://github.com/rshade/pulumicost-core/issues/11)
* implement comprehensive configuration management system ([#37](https://github.com/rshade/pulumicost-core/issues/37)) ([4a21a0c](https://github.com/rshade/pulumicost-core/commit/4a21a0cf1a9c815768e90eebb831d61107554fa0))
* implement comprehensive configuration management system ([#38](https://github.com/rshade/pulumicost-core/issues/38)) ([a06d03b](https://github.com/rshade/pulumicost-core/commit/a06d03b4ad0f122a9d9e4967e9562add0a59c03f))
* implement comprehensive logging and error handling infrastructure ([#59](https://github.com/rshade/pulumicost-core/issues/59)) ([615daaf](https://github.com/rshade/pulumicost-core/commit/615daaf7bf3f1ec45b7b83603c2a70cc3d7f7ac1)), closes [#10](https://github.com/rshade/pulumicost-core/issues/10)
* implement comprehensive testing framework and strategy ([#58](https://github.com/rshade/pulumicost-core/issues/58)) ([c8451af](https://github.com/rshade/pulumicost-core/commit/c8451af5f8a57b901aa15bf2287d8cf6e695a4f4))
* integrate real proto definitions from pulumicost-spec ([247fd5b](https://github.com/rshade/pulumicost-core/commit/247fd5b96e850669e4277519b367048dcb23d3e2))
* **logging:** implement zerolog distributed tracing with debug mode ([#184](https://github.com/rshade/pulumicost-core/issues/184)) ([4be8b26](https://github.com/rshade/pulumicost-core/commit/4be8b26290e2b9eb182082770f78f7db7f31adb9))
* **pluginsdk:** implement Supports() gRPC handler ([#165](https://github.com/rshade/pulumicost-core/issues/165)) ([2034a52](https://github.com/rshade/pulumicost-core/commit/2034a52f6cd8d160bfdfcbe0d94b4a9cca5020ba))


### Fixed

* add index.md for GitHub Pages landing page and fix workflow validation ([#96](https://github.com/rshade/pulumicost-core/issues/96)) ([609e4e2](https://github.com/rshade/pulumicost-core/commit/609e4e2df7c7b51639b21abd2f5f10081658773c))
* add proper CSS styling and layout improvements for GitHub Pages ([#107](https://github.com/rshade/pulumicost-core/issues/107)) ([242b3d0](https://github.com/rshade/pulumicost-core/commit/242b3d06d0138c86a827b2dc8a3edc687b5d72bb))
* add proper CSS styling and layout improvements for GitHub Pages ([#143](https://github.com/rshade/pulumicost-core/issues/143)) ([de35bac](https://github.com/rshade/pulumicost-core/commit/de35bacf1537c5029e8dfd0a18ca2fa6e79a887f))
* **deps:** update github.com/rshade/pulumicost-spec digest to 1130a00 ([#39](https://github.com/rshade/pulumicost-core/issues/39)) ([16112bc](https://github.com/rshade/pulumicost-core/commit/16112bca7bb78716bd1ac4da9c323fabf10c9774))
* **deps:** update github.com/rshade/pulumicost-spec digest to 241cb09 ([#32](https://github.com/rshade/pulumicost-core/issues/32)) ([39a83d8](https://github.com/rshade/pulumicost-core/commit/39a83d8b877be68e2cccacd51e7cc564a8abe69f))
* **deps:** update github.com/rshade/pulumicost-spec digest to 35b5694 ([#79](https://github.com/rshade/pulumicost-core/issues/79)) ([8d03c3e](https://github.com/rshade/pulumicost-core/commit/8d03c3e2b4d7ffe26428ce1ee5012d3e2c508cb9))
* **deps:** update github.com/rshade/pulumicost-spec digest to 5825eaa ([#60](https://github.com/rshade/pulumicost-core/issues/60)) ([3bdc514](https://github.com/rshade/pulumicost-core/commit/3bdc5144141bb05430979fd69614bbcde998cde4))
* **deps:** update github.com/rshade/pulumicost-spec digest to 79d1a15 ([#53](https://github.com/rshade/pulumicost-core/issues/53)) ([e9f4add](https://github.com/rshade/pulumicost-core/commit/e9f4add667a4ef4ca26abb724fbfb5dc831530bc))
* **deps:** update github.com/rshade/pulumicost-spec digest to a085bd2 ([#25](https://github.com/rshade/pulumicost-core/issues/25)) ([bbf4974](https://github.com/rshade/pulumicost-core/commit/bbf4974e6a18dc956c8e8b25a9ed95cc3203bea2))
* **deps:** update github.com/rshade/pulumicost-spec digest to d9f31a6 ([#16](https://github.com/rshade/pulumicost-core/issues/16)) ([644ba4e](https://github.com/rshade/pulumicost-core/commit/644ba4ec5dec924a386a0a0e8613335860ed4e80))
* **deps:** update github.com/rshade/pulumicost-spec digest to e3ffb28 ([#67](https://github.com/rshade/pulumicost-core/issues/67)) ([0135b43](https://github.com/rshade/pulumicost-core/commit/0135b4395c4e8fa98e2ed69d3c48ecb8080805a6))
* **deps:** update go dependencies ([#159](https://github.com/rshade/pulumicost-core/issues/159)) ([b2ad29f](https://github.com/rshade/pulumicost-core/commit/b2ad29fff1ef33a2428a851b02e043f235ea0dad))
* **deps:** update go dependencies ([#33](https://github.com/rshade/pulumicost-core/issues/33)) ([e54dcb3](https://github.com/rshade/pulumicost-core/commit/e54dcb39d08beeb16cbd484d547abd88037c7443))
* **deps:** update go dependencies ([#40](https://github.com/rshade/pulumicost-core/issues/40)) ([e59e319](https://github.com/rshade/pulumicost-core/commit/e59e319cb6b620daecbd786174b98c5004613dc3))
* **deps:** update go dependencies ([#49](https://github.com/rshade/pulumicost-core/issues/49)) ([8b99267](https://github.com/rshade/pulumicost-core/commit/8b99267eb48d6a6f0cbf79d6d84e82b34b1025ff))
* **deps:** update module github.com/rshade/pulumicost-spec to v0.2.0 ([#167](https://github.com/rshade/pulumicost-core/issues/167)) ([b6c9271](https://github.com/rshade/pulumicost-core/commit/b6c92712fc62c90a476e937d4c1dc90882229eaf))
* **deps:** update module github.com/spf13/cobra to v1.9.1 ([#17](https://github.com/rshade/pulumicost-core/issues/17)) ([2e0e8aa](https://github.com/rshade/pulumicost-core/commit/2e0e8aaf7633dfb32e44ab999845bce595be7827))
* **deps:** update module google.golang.org/protobuf to v1.36.10 ([#61](https://github.com/rshade/pulumicost-core/issues/61)) ([5dd8cae](https://github.com/rshade/pulumicost-core/commit/5dd8cae604c72d646afe2adc61d3589b3ace763e))


### Changed

* remove encryption from config, use environment variables for secrets ([#149](https://github.com/rshade/pulumicost-core/issues/149)) ([2e3a07b](https://github.com/rshade/pulumicost-core/commit/2e3a07b6d122ef37e0cff9b9a3d025855b92881b)), closes [#99](https://github.com/rshade/pulumicost-core/issues/99)


### Documentation

* complete Vantage plugin documentation ([#145](https://github.com/rshade/pulumicost-core/issues/145)) ([06e6cd7](https://github.com/rshade/pulumicost-core/commit/06e6cd70a9328bde6d6d736146fe16b088aa1f6d)), closes [#103](https://github.com/rshade/pulumicost-core/issues/103)
* first pass at github pages ([#88](https://github.com/rshade/pulumicost-core/issues/88)) ([ceee2f3](https://github.com/rshade/pulumicost-core/commit/ceee2f3fb632f0d1c8960bb36fce1e111988efd3))
* ratify constitution v1.0.0 (establish governance principles) ([#152](https://github.com/rshade/pulumicost-core/issues/152)) ([d40ac0f](https://github.com/rshade/pulumicost-core/commit/d40ac0fab2707b1acf7a0e2ba0db87e424f4afbe))
* update constitution for docstrings ([#176](https://github.com/rshade/pulumicost-core/issues/176)) ([5053db5](https://github.com/rshade/pulumicost-core/commit/5053db5865b6ecf6e2ec430181a7c9445b47cdab))

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
