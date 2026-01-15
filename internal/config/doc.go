// Package config handles configuration loading and management for FinFocus.
//
// Configuration is loaded from ~/.finfocus/config.yaml with support for:
//   - Plugin directories and settings
//   - Logging configuration (level, format)
//   - Default output format preferences
//
// # Configuration Precedence
//
//  1. CLI flags (highest priority)
//  2. Environment variables (FINFOCUS_*)
//  3. Config file (~/.finfocus/config.yaml)
//  4. Built-in defaults (lowest priority)
//
// # Strict Mode
//
// Enable strict mode with FINFOCUS_CONFIG_STRICT=true to return errors
// instead of falling back to defaults when configuration loading fails.
package config
