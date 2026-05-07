#!/bin/bash

# 本地测试脚本 - 使用 fake Kubernetes 客户端进行测试

set -e

echo "========================================="
echo "Testing k8s-configmap-sidecar"
echo "========================================="

# 运行单元测试
echo ""
echo "Running unit tests..."
go test -v -race -coverprofile=coverage.out ./...

# 显示测试覆盖率
echo ""
echo "Test coverage report:"
go tool cover -func=coverage.out | grep total

# 清理
rm -f coverage.out

echo ""
echo "Tests completed!"
