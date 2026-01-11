package pluginhost

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsUnimplementedError checks if the error is a gRPC Unimplemented error.
func IsUnimplementedError(err error) bool {
	return status.Code(err) == codes.Unimplemented
}
