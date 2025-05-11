#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Variables
APP_NAME="dataseap-server"
CMD_PATH="./cmd/dataseap-server/main.go"
OUTPUT_DIR="./bin"
OUTPUT_BINARY="${OUTPUT_DIR}/${APP_NAME}"

# Version information (can be passed via environment variables or arguments)
VERSION=${APP_VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")}
GIT_COMMIT=${APP_GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
BUILD_DATE=${APP_BUILD_DATE:-$(date -u +'%Y-%m-%dT%H:%M:%SZ')}

# Go build flags
# CGO_ENABLED=0 for static builds, helpful for scratch Docker images or alpine
# -s -w to strip debug symbols and DWARF information, reducing binary size
LDFLAGS="-s -w \
-X github.com/turtacn/dataseap/pkg/common/constants.ServiceVersion=${VERSION} \
-X main.GitCommit=${GIT_COMMIT} \
-X main.BuildDate=${BUILD_DATE}"
# Note: For main.GitCommit and main.BuildDate to work, you need to declare these variables in your main package:
# var (
#	 GitCommit string
#	 BuildDate string
# )

# Ensure output directory exists
mkdir -p ${OUTPUT_DIR}

echo "Building ${APP_NAME}..."
echo "Version: ${VERSION}"
echo "Git Commit: ${GIT_COMMIT}"
echo "Build Date: ${BUILD_DATE}"

# Run linters and tests (optional, but recommended)
# echo "Running linters..."
# golangci-lint run ./...

# echo "Running tests..."
# go test -v ./...

# Build the application
echo "Compiling binary to ${OUTPUT_BINARY}..."
CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${OUTPUT_BINARY} ${CMD_PATH}

echo "Build complete: ${OUTPUT_BINARY}"
ls -lh ${OUTPUT_BINARY}

# Example: Build Docker image after successful Go build (optional)
# BUILD_DOCKER_IMAGE=${BUILD_DOCKER_IMAGE:-false}
# if [ "$BUILD_DOCKER_IMAGE" = true ]; then
#   echo "Building Docker image..."
#   DOCKER_IMAGE_NAME="turtacn/dataseap"
#   DOCKER_IMAGE_TAG="${VERSION}"
#   docker build -t "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" \
#     --build-arg VERSION="${VERSION}" \
#     --build-arg GIT_COMMIT="${GIT_COMMIT}" \
#     --build-arg BUILD_DATE="${BUILD_DATE}" \
#     -f ./deployments/docker/Dockerfile .
#   echo "Docker image built: ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
# fi

exit 0