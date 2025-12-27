# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3](https://github.com/rshade/pulumicost-core/compare/v0.1.2...v0.1.3) (2025-12-27)


### Added

* add integration tests for --filter flag across cost commands ([#300](https://github.com/rshade/pulumicost-core/issues/300)) ([efcebf6](https://github.com/rshade/pulumicost-core/commit/efcebf60efb48f1f57704a24b738478fa8393518)), closes [#249](https://github.com/rshade/pulumicost-core/issues/249)
* **analyzer:** add ResourceID passthrough for recommendation correlation ([#347](https://github.com/rshade/pulumicost-core/issues/347)) ([680b80a](https://github.com/rshade/pulumicost-core/commit/680b80af73acc657dac79d6bf012a7bf0b3af35b)), closes [#106](https://github.com/rshade/pulumicost-core/issues/106)
* **analyzer:** implement Pulumi Analyzer plugin for zero-click cost estimation ([#229](https://github.com/rshade/pulumicost-core/issues/229)) ([2070b05](https://github.com/rshade/pulumicost-core/commit/2070b05513f6e9ae2580930c02abed8fec3fe790))
* **ci:** add automated nightly failure analysis workflow ([#297](https://github.com/rshade/pulumicost-core/issues/297)) ([ab7c516](https://github.com/rshade/pulumicost-core/commit/ab7c516a8b269f578ba309c68d1dd291ef2d00ef)), closes [#271](https://github.com/rshade/pulumicost-core/issues/271)
* **conformance:** add plugin conformance testing framework ([#215](https://github.com/rshade/pulumicost-core/issues/215)) ([c37cc22](https://github.com/rshade/pulumicost-core/commit/c37cc2283919b4ba4ff736f15f42db7c18297fc5)), closes [#201](https://github.com/rshade/pulumicost-core/issues/201)
* **e2e:** implement E2E testing framework with Pulumi Automation API ([#238](https://github.com/rshade/pulumicost-core/issues/238)) ([ee23ff2](https://github.com/rshade/pulumicost-core/commit/ee23ff2b19b348086e83969457c6927a787b96ac)), closes [#177](https://github.com/rshade/pulumicost-core/issues/177)
* implement CLI filter flag with validation and integration tests ([#332](https://github.com/rshade/pulumicost-core/issues/332)) ([b358566](https://github.com/rshade/pulumicost-core/commit/b3585665e7192b74d6bebfaf3fe5be13c8e8d5e6))
* implement sustainability metrics and finalize plugin sdk mapping ([#315](https://github.com/rshade/pulumicost-core/issues/315)) ([f207c53](https://github.com/rshade/pulumicost-core/commit/f207c534fcdd4c64b5498a459529da6a19eec1fa))
* **plugin:** add reference recorder plugin for request capture and mock responses ([#293](https://github.com/rshade/pulumicost-core/issues/293)) ([733c2f9](https://github.com/rshade/pulumicost-core/commit/733c2f969952718ecde99ea9a8b5a64c74b6ac58))
* **tui:** add shared TUI package with Bubble Tea/Lip Gloss components ([#258](https://github.com/rshade/pulumicost-core/issues/258)) ([e049460](https://github.com/rshade/pulumicost-core/commit/e049460e4ccd5545f456ecf9d2051a6f0bac94f9))
* **tui:** add Spinner and Table components from bubbles library ([#341](https://github.com/rshade/pulumicost-core/issues/341)) ([992db5a](https://github.com/rshade/pulumicost-core/commit/992db5ab4ef20cdce6e1f5d6c1def7382ff03628))


### Fixed

* **deps:** update go dependencies ([#281](https://github.com/rshade/pulumicost-core/issues/281)) ([73364d6](https://github.com/rshade/pulumicost-core/commit/73364d66cf1d53512867cf203689998dcc9b3af6))
* **deps:** update go dependencies ([#314](https://github.com/rshade/pulumicost-core/issues/314)) ([c09f298](https://github.com/rshade/pulumicost-core/commit/c09f298281c8b7e18d47fe086dd6fb5d921fd571))
* **deps:** update module github.com/rshade/pulumicost-spec to v0.4.3 ([#211](https://github.com/rshade/pulumicost-core/issues/211)) ([4cb56d9](https://github.com/rshade/pulumicost-core/commit/4cb56d928ab0b5887fd2fc56c182383d9eedfffe))
* **deps:** update module github.com/spf13/cobra to v1.10.2 ([#240](https://github.com/rshade/pulumicost-core/issues/240)) ([ad3bfd7](https://github.com/rshade/pulumicost-core/commit/ad3bfd7b92d189a912dbae3ae10bbda2067e6bf2))
* update Go version to 1.25.5 and improve plugin integration tests ([#244](https://github.com/rshade/pulumicost-core/issues/244)) ([4f383df](https://github.com/rshade/pulumicost-core/commit/4f383df0df1e1d4d3d23259adef8eb29d6ea41e9))


### Changed

* **pluginhost:** remove PORT env var, use --port flag only ([#295](https://github.com/rshade/pulumicost-core/issues/295)) ([46bcdf2](https://github.com/rshade/pulumicost-core/commit/46bcdf24b718e6f43f0d8f5cf3092d79ac35f8ec))
* **pluginsdk:** adopt pluginsdk environment variable constants ([#272](https://github.com/rshade/pulumicost-core/issues/272)) ([8c6e616](https://github.com/rshade/pulumicost-core/commit/8c6e616bcc33bcd79a599d9a31b218e4aa67c34c)), closes [#230](https://github.com/rshade/pulumicost-core/issues/230)


### Documentation

* **all:** synchronize documentation with codebase features ([#257](https://github.com/rshade/pulumicost-core/issues/257)) ([5881cdc](https://github.com/rshade/pulumicost-core/commit/5881cdcbbd27705d35de3de285411ebcabe4b602)), closes [#256](https://github.com/rshade/pulumicost-core/issues/256)

## [0.1.2](https://github.com/rshade/pulumicost-core/compare/v0.1.1...v0.1.2) (2025-12-03)


### Added

* **logging:** integrate zerolog logging across all components ([#206](https://github.com/rshade/pulumicost-core/issues/206)) ([c152d05](https://github.com/rshade/pulumicost-core/commit/c152d0537c394ffd4a0f07554ec12116cb5dc4a2))


### Fixed

* comprehensive input validation and error handling improvements ([#196](https://github.com/rshade/pulumicost-core/issues/196)) ([47b0e36](https://github.com/rshade/pulumicost-core/commit/47b0e369db86f6268a5e9d0aba87ae5f77773379))
* **deps:** update module github.com/masterminds/semver/v3 to v3.4.0 ([#199](https://github.com/rshade/pulumicost-core/issues/199)) ([be86a7e](https://github.com/rshade/pulumicost-core/commit/be86a7ef047d938b4a2c87ad7fff8f727be693ee))
* **pluginhost:** prevent race condition in plugin port allocation ([#192](https://github.com/rshade/pulumicost-core/issues/192)) ([42c4a0a](https://github.com/rshade/pulumicost-core/commit/42c4a0a488a0aa3f528579640e49ba77c3198d71))

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
