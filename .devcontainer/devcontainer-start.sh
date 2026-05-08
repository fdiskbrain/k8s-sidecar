#!/bin/bash

# DevContainer 快速启动脚本
# 用于在命令行中快速进入开发容器

set -e

echo "🚀 Starting k8s-sidecar DevContainer..."
echo ""

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

# 检查 VS Code 是否安装
if ! command -v code &> /dev/null; then
    echo "⚠️  Warning: VS Code (code command) not found in PATH."
    echo "   Please install VS Code or open the project manually in VS Code."
    echo "   Then use: Dev Containers: Reopen in Container"
    exit 1
fi

# 打开项目
echo "📂 Opening project in DevContainer..."
code --folder-uri vscode-remote://dev-container+$(pwd | sed 's/\//%2F/g')/workspace

echo ""
echo "✅ DevContainer is starting..."
echo "   Please wait for the container to build and initialize."
echo "   Check VS Code status bar for progress."
