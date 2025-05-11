# Stage 1: Build the application
FROM golang:1.20.2-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools (e.g., git for private modules if any)
RUN apk add --no-cache git build-base

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the entire application source code
COPY . .

# Set build arguments for version and other metadata (optional)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION="0.1.0-dev"
ARG GIT_COMMIT=""
ARG BUILD_DATE=""

# Build the application
# CGO_ENABLED=0 for static linking & smaller images where possible
# -ldflags for embedding version information
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w \
    -X github.com/turtacn/dataseap/pkg/common/constants.ServiceVersion=${VERSION} \
    -X 'main.GitCommit=${GIT_COMMIT}' \
    -X 'main.BuildDate=${BUILD_DATE}'" \
    -o /app/bin/dataseap-server ./cmd/dataseap-server/main.go

# Stage 2: Create the final lightweight image
FROM alpine:3.19

# Set working directory
WORKDIR /app

# Create a non-root user and group for security
RUN addgroup -S dataseap && adduser -S -G dataseap dataseap

# Copy the built binary from the builder stage
COPY --from=builder /app/bin/dataseap-server /app/dataseap-server

# Copy configuration files (optional, can also be mounted via volume)
# Create a default config directory and copy a sample if you have one.
# For this example, we assume config will be mounted or default path inside container is used.
RUN mkdir -p /app/config && chown dataseap:dataseap /app/config
# COPY config/config.yaml.example /app/config/config.yaml

# Ensure the binary is executable
RUN chmod +x /app/dataseap-server

# Change ownership to the non-root user
RUN chown dataseap:dataseap /app/dataseap-server

# Switch to the non-root user
USER dataseap

# Expose the default HTTP and gRPC ports
# These should match the default ports in your application config
EXPOSE 8080 50051

# Set the entrypoint for the container
ENTRYPOINT ["/app/dataseap-server"]

# Default command (can be overridden)
# CMD ["--config", "/app/config/config.yaml"]