#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-1.0.0}"
BINARY="bin/spark-debug-mcp"
IMAGE="spark-debug-mcp:${VERSION}"

echo "=== Spark Debug MCP Build ==="
echo "Version: ${VERSION}"

echo ""
echo "--- Formatting ---"
go fmt ./...

echo ""
echo "--- Linting ---"
if command -v golangci-lint &>/dev/null; then
  golangci-lint run ./...
else
  echo "golangci-lint not installed, skipping (install: https://golangci-lint.run/welcome/install/)"
fi

echo ""
echo "--- Testing ---"
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1

echo ""
echo "--- Building binary ---"
mkdir -p bin
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o "${BINARY}" ./cmd/server/
echo "Binary: ${BINARY}"

echo ""
echo "--- Building Docker image ---"
if command -v docker &>/dev/null; then
  docker build -t "${IMAGE}" .
  echo "Image: ${IMAGE}"
else
  echo "Docker not installed, skipping image build"
fi

echo ""
echo "=== Build Complete ==="
echo "Version: ${VERSION}"
"${BINARY}" version
