// Package logging provides structured logging with distributed tracing support.
//
// PulumiCost uses zerolog for high-performance structured logging with
// automatic trace ID propagation through contexts.
//
// # Log Levels
//
//   - TRACE: Property extraction, detailed calculations
//   - DEBUG: Function entry/exit, retries, intermediate values
//   - INFO: High-level operations (command start/end)
//   - WARN: Recoverable issues (fallbacks, deprecations)
//   - ERROR: Failures needing attention
//
// # Trace ID Management
//
// Trace IDs are automatically generated or extracted from context:
//
//	traceID := logging.GetOrGenerateTraceID(ctx)
//	ctx = logging.ContextWithTraceID(ctx, traceID)
//
// # Component Loggers
//
// Create sub-loggers for components:
//
//	logger = logging.ComponentLogger(logger, "registry")
//
// # Configuration
//
// Logging can be configured via:
//   - CLI flags (--debug)
//   - Environment variables (PULUMICOST_LOG_LEVEL, PULUMICOST_LOG_FORMAT)
//   - Config file (~/.pulumicost/config.yaml)
package logging
