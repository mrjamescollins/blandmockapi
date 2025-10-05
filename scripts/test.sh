#!/bin/bash

# Test script for blandmockapi
# Runs unit tests and integration tests with coverage

set -e

echo "==> Running unit tests..."
go test ./internal/... -v -cover -coverprofile=coverage-unit.out

echo ""
echo "==> Generating unit test coverage report..."
go tool cover -func=coverage-unit.out

echo ""
echo "==> Running integration tests..."
go test -tags=integration ./test/integration/... -v -timeout 30s

echo ""
echo "==> All tests passed!"

# Generate HTML coverage report
if [ "$1" == "--html" ]; then
    echo ""
    echo "==> Generating HTML coverage report..."
    go tool cover -html=coverage-unit.out -o coverage.html
    echo "Coverage report: coverage.html"
fi
