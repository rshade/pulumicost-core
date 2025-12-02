package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// auditLoggerKey is a private type for context keys to avoid collisions.
type auditLoggerKey struct{}

// AuditEntry represents an audit log record for cost operations.
type AuditEntry struct {
	Timestamp   time.Time         // When the operation occurred
	TraceID     string            // Request correlation ID
	Command     string            // CLI command name (e.g., "cost projected")
	Parameters  map[string]string // Relevant parameters (file path, dates, etc.)
	Duration    time.Duration     // How long the operation took
	Success     bool              // Whether operation succeeded
	ResultCount int               // Number of results returned
	TotalCost   float64           // Total cost calculated (if applicable)
	Error       string            // Error message if failed
}

// NewAuditEntry creates a new AuditEntry with the given command and trace ID.
// Use the With* methods to populate additional fields.
func NewAuditEntry(command, traceID string) *AuditEntry {
	return &AuditEntry{
		Timestamp:  time.Now().UTC(),
		TraceID:    traceID,
		Command:    command,
		Parameters: make(map[string]string),
	}
}

// WithParameters adds parameters to the audit entry.
func (e *AuditEntry) WithParameters(params map[string]string) *AuditEntry {
	e.Parameters = params
	return e
}

// WithSuccess marks the entry as successful with result count and total cost.
func (e *AuditEntry) WithSuccess(resultCount int, totalCost float64) *AuditEntry {
	e.Success = true
	e.ResultCount = resultCount
	e.TotalCost = totalCost
	return e
}

// WithError marks the entry as failed with the given error message.
func (e *AuditEntry) WithError(errMsg string) *AuditEntry {
	e.Success = false
	e.Error = errMsg
	return e
}

// WithDuration calculates and sets the duration from the given start time.
func (e *AuditEntry) WithDuration(start time.Time) *AuditEntry {
	e.Duration = time.Since(start)
	return e
}

// AuditLogger writes audit entries.
type AuditLogger interface {
	// Log writes an audit entry
	Log(ctx context.Context, entry AuditEntry)

	// Enabled returns whether audit logging is active
	Enabled() bool
}

// AuditLoggerConfig holds configuration for creating an AuditLogger.
type AuditLoggerConfig struct {
	Enabled bool      // Enable audit logging
	Writer  io.Writer // Where to write audit logs (nil uses os.Stderr)
	File    string    // Optional: separate audit file path
}

// zerologAuditLogger implements AuditLogger using zerolog.
type zerologAuditLogger struct {
	logger  zerolog.Logger
	enabled bool
}

// NewAuditLogger creates a new AuditLogger with the given configuration.
//
//nolint:nestif // Writer initialization has acceptable complexity for configuration logic
func NewAuditLogger(cfg AuditLoggerConfig) AuditLogger {
	if !cfg.Enabled {
		return &noOpAuditLogger{}
	}

	writer := cfg.Writer
	if writer == nil {
		// Try to open the audit file if specified
		if cfg.File != "" {
			file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				// Fall back to stderr
				writer = os.Stderr
			} else {
				writer = file
			}
		} else {
			writer = os.Stderr
		}
	}

	logger := zerolog.New(writer).
		With().
		Timestamp().
		Str("component", "audit").
		Logger()

	return &zerologAuditLogger{
		logger:  logger,
		enabled: true,
	}
}

// Log writes an audit entry to the log.
func (a *zerologAuditLogger) Log(_ context.Context, entry AuditEntry) {
	if !a.enabled {
		return
	}

	event := a.logger.Info().
		Bool("audit", true).
		Str("command", entry.Command).
		Str("trace_id", entry.TraceID).
		Int64("duration_ms", entry.Duration.Milliseconds()).
		Bool("success", entry.Success)

	// Add result fields if successful
	if entry.Success {
		event = event.
			Int("result_count", entry.ResultCount).
			Float64("total_cost", entry.TotalCost)
	}

	// Add error if present
	if entry.Error != "" {
		event = event.Str("error", entry.Error)
	}

	// Add parameters with redaction
	if len(entry.Parameters) > 0 {
		safeParams := SafeParams(entry.Parameters)
		event = event.Interface("parameters", safeParams)
	}

	event.Msg("cost query completed")
}

// Enabled returns whether audit logging is active.
func (a *zerologAuditLogger) Enabled() bool {
	return a.enabled
}

// noOpAuditLogger is a no-operation implementation of AuditLogger.
type noOpAuditLogger struct{}

// Log does nothing.
func (n *noOpAuditLogger) Log(_ context.Context, _ AuditEntry) {}

// Enabled returns false.
func (n *noOpAuditLogger) Enabled() bool {
	return false
}

// NoOpAuditLogger returns a no-op audit logger for when auditing is disabled.
func NoOpAuditLogger() AuditLogger {
	return &noOpAuditLogger{}
}

// SafeParams creates a copy of parameters with sensitive values redacted.
func SafeParams(params map[string]string) map[string]string {
	safe := make(map[string]string, len(params))
	for k, v := range params {
		if IsSensitiveKey(k) {
			safe[k] = "[REDACTED]"
		} else {
			safe[k] = v
		}
	}
	return safe
}

// IsSensitiveKey checks if a key name contains sensitive patterns.
// This is exported to allow callers to check before logging.
func IsSensitiveKey(key string) bool {
	return isSensitiveKey(key)
}

// ContextWithAuditLogger stores an AuditLogger in the context.
func ContextWithAuditLogger(ctx context.Context, logger AuditLogger) context.Context {
	return context.WithValue(ctx, auditLoggerKey{}, logger)
}

// AuditLoggerFromContext extracts the AuditLogger from context.
// Returns a no-op logger if none is stored.
func AuditLoggerFromContext(ctx context.Context) AuditLogger {
	if logger, ok := ctx.Value(auditLoggerKey{}).(AuditLogger); ok {
		return logger
	}
	return &noOpAuditLogger{}
}
