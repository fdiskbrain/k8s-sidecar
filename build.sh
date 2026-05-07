#!/bin/bash

set -e

echo "========================================="
echo "Building k8s-configmap-sidecar"
echo "========================================="

# 设置版本号（如果未提供则使用 latest）
VERSION=${VERSION:-latest}

# 下载依赖
echo ""
echo "Downloading dependencies..."
go mod download

# 编译
echo ""
echo "Building binary (version: $VERSION)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags "-X main.Version=${VERSION}" \
  -o ./bin/k8s-configmap-sidecar \
  ./cmd/sidecar/main.go

echo ""
echo "Build completed successfully!"
echo "Binary location: ./bin/k8s-configmap-sidecar"
echo ""
echo "To run locally:"
echo "  ./bin/k8s-configmap-sidecar --help"
