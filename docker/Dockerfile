# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application directly (avoid make dependency on external tools)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p bin && \
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "docker") && \
    COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown") && \
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    CGO_ENABLED=0 go build \
        -ldflags "-X github.com/rshade/pulumicost-core/pkg/version.version=${VERSION} -X github.com/rshade/pulumicost-core/pkg/version.gitCommit=${COMMIT} -X github.com/rshade/pulumicost-core/pkg/version.buildDate=${BUILD_DATE}" \
        -o bin/pulumicost ./cmd/pulumicost

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S pulumicost && \
    adduser -u 1001 -S pulumicost -G pulumicost

# Create directories for pulumicost data
RUN mkdir -p /home/pulumicost/.pulumicost/plugins && \
    mkdir -p /home/pulumicost/.pulumicost/specs && \
    chown -R pulumicost:pulumicost /home/pulumicost/.pulumicost

# Copy binary from builder stage
COPY --from=builder /app/bin/pulumicost /usr/local/bin/pulumicost

# Make binary executable
RUN chmod +x /usr/local/bin/pulumicost

# Switch to non-root user
USER pulumicost

# Set working directory
WORKDIR /home/pulumicost

# Set environment variables
ENV HOME=/home/pulumicost
ENV PATH=/usr/local/bin:$PATH

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pulumicost --help || exit 1

# Default command
ENTRYPOINT ["pulumicost"]
CMD ["--help"]