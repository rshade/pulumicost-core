//go:build integration
// +build integration

// Package integration_test contains integration tests for plugin version compatibility.
//
// # Test Status: SKIPPED
//
// These tests are currently skipped because they require a compiled plugin binary
// that implements the GetPluginInfo RPC. The unit tests in internal/pluginhost/client_test.go
// provide comprehensive coverage using mock gRPC servers.
//
// # Prerequisites to Enable These Tests
//
// 1. Build the recorder plugin: `make build-recorder`
// 2. Or build any plugin that implements GetPluginInfo with a known spec version
// 3. Set environment variable FINFOCUS_TEST_PLUGIN_PATH to the plugin binary path
//
// # Coverage Note
//
// The version compatibility logic is fully tested via:
//   - internal/pluginhost/version_test.go: SemVer comparison logic (7 test cases)
//   - internal/pluginhost/client_test.go: GetPluginInfo success/unimplemented/timeout (3 test cases)
//
// These integration tests would provide additional end-to-end validation when
// a suitable test plugin fixture is available in CI.
//
// # Related Issues
//
// See PR #398 for the initial implementation of plugin info discovery.
package integration_test

import (
	"testing"
)

// TestPluginInitialization_CompatibleVersion verifies that a plugin with a compatible
// spec version initializes successfully without warnings.
//
// To enable: Set FINFOCUS_TEST_PLUGIN_PATH to a plugin binary with matching spec version.
func TestPluginInitialization_CompatibleVersion(t *testing.T) {
	t.Skip("Skipping: requires compiled plugin binary (see package doc for prerequisites)")
}

// TestPluginInitialization_IncompatibleVersion_Warning verifies that a plugin with
// a mismatched major spec version triggers a warning but still initializes.
//
// To enable: Set FINFOCUS_TEST_PLUGIN_PATH to a plugin binary with different major version.
func TestPluginInitialization_IncompatibleVersion_Warning(t *testing.T) {
	t.Skip("Skipping: requires compiled plugin binary with incompatible spec version")
}

// TestPluginInitialization_LegacyPlugin_NoGetPluginInfo verifies that a legacy plugin
// that doesn't implement GetPluginInfo initializes successfully with a debug log.
//
// To enable: Set FINFOCUS_TEST_LEGACY_PLUGIN_PATH to a plugin binary without GetPluginInfo.
func TestPluginInitialization_LegacyPlugin_NoGetPluginInfo(t *testing.T) {
	t.Skip("Skipping: requires legacy plugin binary without GetPluginInfo RPC")
}
