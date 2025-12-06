// Package analyzer implements the Pulumi Analyzer plugin interface for cost estimation.
//
// The analyzer integrates with Pulumi's preview workflow to provide real-time cost
// estimates for cloud resources. It implements the pulumirpc.Analyzer gRPC service
// interface, enabling zero-click cost estimation directly within `pulumi preview` output.
//
// # Architecture
//
// The package consists of three main components:
//
//   - Server: gRPC service implementation (server.go)
//   - Mapper: Resource type conversion (mapper.go)
//   - Diagnostics: Cost result formatting (diagnostics.go)
//
// # Protocol
//
// The analyzer follows Pulumi's plugin handshake protocol:
//
//  1. Plugin starts and listens on a random TCP port
//  2. Port number is printed to stdout (CRITICAL: only output to stdout)
//  3. Pulumi engine connects via gRPC
//  4. Handshake and ConfigureStack RPCs establish context
//  5. AnalyzeStack RPC receives all resources for cost calculation
//  6. Diagnostics are returned with cost estimates
//
// # Usage
//
// The analyzer is invoked via the CLI:
//
//	pulumicost analyzer serve
//
// And configured in Pulumi.yaml:
//
//	analyzers:
//	  - cost
//
// # Configuration
//
// Analyzer settings are read from ~/.pulumicost/config.yaml:
//
//	analyzer:
//	  timeout:
//	    per_resource: 5s
//	    total: 60s
//	    warn_threshold: 30s
//
// # Logging
//
// All logs are written to stderr to preserve the stdout handshake.
// Use the existing zerolog configuration via internal/logging.
package analyzer
