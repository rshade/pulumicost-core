package recorder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Recorder handles serialization of gRPC requests to JSON files.
// It is thread-safe and handles I/O errors gracefully.
type Recorder struct {
	outputDir string
	logger    zerolog.Logger
	mu        sync.Mutex
	disabled  bool // Set to true if output directory is not writable
}

// RecordedRequest represents a captured gRPC request.
type RecordedRequest struct {
	Timestamp string          `json:"timestamp"`
	Method    string          `json:"method"`
	RequestID string          `json:"requestId"`
	Request   json.RawMessage `json:"request"`
	Metadata  RequestMetadata `json:"metadata"`
}

// RequestMetadata contains optional metadata about the request.
type RequestMetadata struct {
	ReceivedAt       string `json:"receivedAt"`
	ProcessingTimeMs int64  `json:"processingTimeMs"`
}

// NewRecorder creates a new Recorder instance.
// It creates the output directory if it doesn't exist.
func NewRecorder(outputDir string, logger zerolog.Logger) *Recorder {
	r := &Recorder{
		outputDir: outputDir,
		logger:    logger.With().Str("component", "recorder").Logger(),
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		r.logger.Error().Err(err).Str("dir", outputDir).Msg("failed to create output directory")
		r.disabled = true
		return r
	}

	// Verify directory is writable
	testFile := filepath.Join(outputDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		r.logger.Error().Err(err).Str("dir", outputDir).Msg("output directory is not writable")
		r.disabled = true
		return r
	}
	_ = os.Remove(testFile) // Best effort cleanup

	r.logger.Info().Str("dir", outputDir).Msg("recorder initialized")
	return r
}

// RecordRequest serializes a gRPC request to a JSON file.
// It returns an error if the write fails, but the caller should
// typically log and continue rather than failing the request.
func (r *Recorder) RecordRequest(method string, req proto.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.disabled {
		return errors.New("recorder disabled: output directory not writable")
	}

	startTime := time.Now().UTC()

	// Generate unique filename
	filename := r.generateFilename(method)
	filePath := filepath.Join(r.outputDir, filename)

	// Serialize the protobuf request to JSON
	requestJSON, err := r.serializeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	// Create the recorded request structure
	recorded := RecordedRequest{
		Timestamp: startTime.Format(time.RFC3339),
		Method:    method,
		RequestID: ulid.Make().String(),
		Request:   requestJSON,
		Metadata: RequestMetadata{
			ReceivedAt:       startTime.Format(time.RFC3339Nano),
			ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(recorded, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal recorded request: %w", err)
	}

	// Write to file
	if writeErr := os.WriteFile(filePath, data, 0600); writeErr != nil {
		// Check for disk full
		if isDiskFullError(writeErr) {
			r.logger.Warn().Err(writeErr).Msg("disk full - disabling recording")
			r.disabled = true
		}
		return fmt.Errorf("failed to write file: %w", writeErr)
	}

	r.logger.Debug().
		Str("file", filename).
		Str("method", method).
		Msg("recorded request")

	return nil
}

// generateFilename creates a unique filename for a recorded request.
// The format is: <timestamp>_<method>_<ulid>.json.
func (r *Recorder) generateFilename(method string) string {
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	id := ulid.Make()
	return fmt.Sprintf("%s_%s_%s.json", timestamp, method, id.String())
}

// serializeRequest converts a protobuf message to JSON.
func (r *Recorder) serializeRequest(req proto.Message) (json.RawMessage, error) {
	opts := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	data, err := opts.Marshal(req)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Close performs cleanup when the recorder is shutting down.
func (r *Recorder) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info().Msg("recorder closed")
}

// isDiskFullError checks if an error indicates disk full condition.
func isDiskFullError(err error) bool {
	if err == nil {
		return false
	}
	// Check for permission errors first - these are not disk full
	if os.IsPermission(err) {
		return false
	}
	// Check error message for common disk full indicators
	errMsg := err.Error()
	return strings.Contains(errMsg, "no space left on device") ||
		strings.Contains(errMsg, "disk quota exceeded") ||
		strings.Contains(errMsg, "ENOSPC")
}
