# DevContainer 实施总结

## 📦 已创建的文件

### 核心配置文件

1. **`.devcontainer/devcontainer.json`** - DevContainer 主配置
   - 定义容器构建方式
   - VS Code 扩展和设置
   - 端口转发和环境变量
   - Kubeconfig 挂载配置

2. **`.devcontainer/Dockerfile`** - 开发容器镜像定义
   - 基于 golang:1.25-alpine
   - 安装 gcc、musl-dev（支持 race 检测）
   - 配置中国镜像加速
   - 创建非 root 用户 (vscode)

3. **`.devcontainer/README.md`** - 使用文档
   - 快速开始指南
   - 环境特性说明
   - 常用命令参考
   - 故障排查指南

### 辅助文件

4. **`.devcontainer/devcontainer-start.sh`** - 快速启动脚本
   - 一键启动 DevContainer
   - Docker 和 VS Code 检查

5. **`.devcontainer/config.example.yaml`** - 配置示例
   - 展示在 DevContainer 中如何使用

6. **`.devcontainer/devcontainer-feature.json`** - 特性配置（可选）

## ✨ 主要特性

### 1. 完整的 Go 开发环境
- ✅ Go 1.25（项目要求版本）
- ✅ CGO 支持（gcc + musl-dev）
- ✅ Race 检测能力
- ✅ 中国镜像加速（GOPROXY）

### 2. VS Code 集成
- ✅ 自动安装 Go 语言扩展
- ✅ 代码格式化（goimports）
- ✅ Lint 工具（golangci-lint）
- ✅ Kubernetes 和 Docker 扩展

### 3. Kubernetes 支持
- ✅ 自动挂载 ~/.kube 目录
- ✅ kubectl 可用
- ✅ 可直接操作集群

### 4. Docker-in-Docker
- ✅ 可在容器内构建 Docker 镜像
- ✅ 支持多架构构建
- ✅ 测试 CI/CD 流程

### 5. 性能优化
- ✅ Go modules 缓存
- ✅ 中国镜像源（goproxy.cn + 阿里云）
- ✅ Alpine APK 镜像加速

## 🚀 使用方法

### 方法 1：VS Code 图形界面（推荐）

1. 安装 [Dev Containers 扩展](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. 在 VS Code 中打开项目
3. 按 `F1` → 输入 "Dev Containers: Reopen in Container"
4. 等待容器构建完成

### 方法 2：命令行

```bash
# 使用提供的启动脚本
.devcontainer/devcontainer-start.sh

# 或使用 VS Code CLI
code .
# 然后在 VS Code 中选择 Reopen in Container
```

## 📝 开发工作流

### 进入容器后

```bash
# 1. 验证环境
go version
make deps

# 2. 运行测试
make test RACE=""     # 快速测试
make test             # 带 race 检测

# 3. 构建项目
make build

# 4. 本地运行
./bin/k8s-sidecar --help

# 5. 构建 Docker 镜像
make docker-build VERSION=latest
```

### 常用 Make 命令

```bash
make deps      # 下载依赖
make fmt       # 格式化代码
make vet       # 静态检查
make test      # 运行测试
make build     # 编译二进制
make run       # 本地运行
```

## 🔧 自定义配置

### 添加额外工具

编辑 `.devcontainer/Dockerfile`：

```dockerfile
RUN apk add --no-cache <package-name>
```

然后重建容器：**Dev Containers: Rebuild Container**

### 修改 VS Code 设置

编辑 `.devcontainer/devcontainer.json` 中的 `customizations.vscode.settings` 部分。

### 添加更多扩展

在 `customizations.vscode.extensions` 数组中添加扩展 ID。

## 🐛 常见问题

### Q: 容器构建很慢？
A: 首次构建需要下载依赖，后续会使用缓存。确保网络通畅，已配置中国镜像加速。

### Q: Kubeconfig 无法访问？
A: 确保 `~/.kube/config` 文件存在且权限正确（600）。

### Q: Go 工具未安装？
A: 容器启动时会自动安装，或手动运行：
```bash
make deps
go install golang.org/x/tools/cmd/goimports@latest
```

### Q: 如何退出容器？
A: 在 VS Code 中：**Dev Containers: Reopen Folder Locally**

## 📊 与 GitLab CI 的一致性

DevContainer 环境与 GitLab CI 保持一致：
- ✅ 相同的 Go 版本（1.25）
- ✅ 相同的镜像源配置
- ✅ 相同的构建工具（make, gcc）
- ✅ 支持 race 检测
- ✅ 非 root 用户运行

这确保了"在我机器上能跑"的问题不再出现！

## 🎯 下一步

1. ✅ DevContainer 配置已完成
2. 📖 阅读 [.devcontainer/README.md](.devcontainer/README.md) 了解详细用法
3. 🚀 尝试在 DevContainer 中开发
4. 💡 根据团队需求自定义配置

## 📚 相关文档

- [DevContainers 官方文档](https://code.visualstudio.com/docs/devcontainers/containers)
- [DevContainer 规范](https://containers.dev/)
- [项目主 README](../README.md)
- [GitLab CI 指南](../GITLAB_CI_GUIDE.md)
