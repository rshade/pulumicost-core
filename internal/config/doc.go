// Package config handles configuration loading and management for PulumiCost.
//
// Configuration is loaded from ~/.pulumicost/config.yaml with support for:
//   - Plugin directories and settings
//   - Logging configuration (level, format)
//   - Default output format preferences
//
// # Configuration Precedence
//
//  1. CLI flags (highest priority)
//  2. Environment variables (PULUMICOST_*)
//  3. Config file (~/.pulumicost/config.yaml)
//  4. Built-in defaults (lowest priority)
//
// # Strict Mode
//
// Enable strict mode with PULUMICOST_CONFIG_STRICT=true to return errors
// instead of falling back to defaults when configuration loading fails.
package config
