// Package conformance provides a comprehensive testing framework for validating
// plugin protocol compliance with the PulumiCost gRPC protocol.
//
// The conformance suite enables plugin developers to verify their implementations
// against the protocol requirements defined in the pulumicost-spec repository.
// It supports both TCP and stdio communication modes and provides detailed
// reporting in multiple formats (table, JSON, JUnit XML).
//
// # Components
//
// The package consists of several key components:
//
//   - Suite: The main orchestrator that runs conformance tests against a plugin
//   - Runner: Executes individual test cases with timeout and error handling
//   - Reporter: Generates test reports in table, JSON, and JUnit XML formats
//   - Types: Shared type definitions for test configuration and results
//
// # Usage
//
// Basic usage via CLI:
//
//	pulumicost plugin conformance ./my-plugin-binary
//
// Programmatic usage:
//
//	suite, err := conformance.NewSuite(conformance.SuiteConfig{
//	    PluginPath: "./my-plugin",
//	    CommMode:   conformance.CommModeTCP,
//	    Verbosity:  conformance.VerbosityNormal,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	report, err := suite.Run(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	report.WriteTable(os.Stdout)
//
// # Test Categories
//
// Tests are organized into categories:
//
//   - protocol: Basic protocol compliance (Name RPC, response formats)
//   - error: Error handling and gRPC status codes
//   - performance: Timeout behavior and batch handling
//   - context: Context cancellation and deadline propagation
//
// # Output Formats
//
// The suite supports three output formats:
//
//   - table: Human-readable console output with pass/fail indicators
//   - json: Machine-readable JSON for programmatic processing
//   - junit: JUnit XML format for CI/CD integration
package conformance
