#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "${PROJECT_ROOT}"

echo "Running linters..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint could not be found. Installing..."
    # This might require sudo or different installation steps depending on the system.
    # Consider providing alternative instructions or requiring it as a prerequisite.
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    # Ensure GOBIN is in PATH or use $(go env GOPATH)/bin/golangci-lint
    if ! command -v golangci-lint &> /dev/null
    then
        echo "Failed to install golangci-lint or it's not in PATH."
        echo "Please install it manually: https://golangci-lint.run/usage/install/"
        exit 1
    fi
fi

# Run golangci-lint
# Adjust flags as needed.
# --fix: to automatically fix issues (use with caution, review changes)
# -E <linter1>,<linter2>: enable specific linters
# -D <linter1>,<linter2>: disable specific linters
# --timeout: set a timeout for the linting process
echo "Executing golangci-lint run ./..."
golangci-lint run ./... --timeout 5m

# You can also run go vet for additional checks
# echo "Running go vet..."
# go vet ./...

# Check formatting with gofmt or gofumpt
# echo "Checking formatting with gofmt..."
# FMT_OUTPUT=$(gofmt -l .) # -l lists files that differ
# if [ -n "$FMT_OUTPUT" ]; then
#    echo "The following files are not correctly formatted (run 'gofmt -w .'):"
#    echo "$FMT_OUTPUT"
#    # exit 1 # Optionally fail the build if not formatted
# else
#    echo "All files are correctly formatted."
# fi


echo "Linting complete."
exit 0