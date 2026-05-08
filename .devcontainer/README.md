# DevContainer for k8s-sidecar

本目录包含 k8s-sidecar 项目的 DevContainer 配置，提供一致且可复现的开发环境。

## 📋 前置要求

- [Visual Studio Code](https://code.visualstudio.com/)
- [Dev Containers 扩展](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) 或 [Podman](https://podman.io/)

## 🚀 快速开始

### 1. 在容器中打开项目

1. 在 VS Code 中打开项目文件夹
2. 按 `F1` 或 `Ctrl+Shift+P` 打开命令面板
3. 输入并选择：**Dev Containers: Reopen in Container**
4. 等待容器构建和初始化完成（首次可能需要几分钟）

### 2. 验证环境

容器启动后，在终端运行以下命令验证环境：

```bash
# 检查 Go 版本
go version

# 检查依赖
make deps

# 运行测试
make test RACE=""

# 构建项目
make build
```

## 🛠️ 开发环境特性

### 预装工具

- **Go 1.25** - 项目要求的 Go 版本
- **gcc & musl-dev** - 支持 CGO 和 race 检测
- **Docker-in-Docker** - 可在容器内构建 Docker 镜像
- **git, make, curl** - 基础开发工具
- **vim, bash** - 编辑器支持

### VS Code 扩展

自动安装以下扩展：

- **[Go](https://marketplace.visualstudio.com/items?itemName=golang.go)** - Go 语言支持
- **[Docker](https://marketplace.visualstudio.com/items?itemName=ms-azuretools.vscode-docker)** - Docker 管理
- **[YAML](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)** - YAML 文件支持
- **[Kubernetes](https://marketplace.visualstudio.com/items?itemName=ms-kubernetes-tools.vscode-kubernetes-tools)** - K8s 资源管理

### 环境变量

已配置中国镜像加速：

```bash
GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
GONOSUMDB=*
GONOPROXY=*
CGO_ENABLED=1
```

### Kubeconfig 支持

自动挂载本地 `~/.kube` 目录到容器中，方便进行 Kubernetes 集群操作。

## 📝 常用命令

### 构建和测试

```bash
# 下载依赖
make deps

# 格式化代码
make fmt

# 运行静态检查
make vet

# 运行测试（带 race 检测）
make test

# 运行测试（不带 race 检测，更快）
make test RACE=""

# 构建二进制
make build

# 本地运行
make run
```

### Docker 操作

```bash
# 构建单平台镜像
make docker-build VERSION=latest

# 构建多架构镜像
make docker-buildx PLATFORMS=linux/amd64,linux/arm64

# 构建并加载到本地 Docker
make docker-buildx-local VERSION=latest
```

## 🔧 自定义配置

### 修改 DevContainer 配置

编辑以下文件来自定义开发环境：

- **devcontainer.json** - 主配置文件（扩展、设置、端口等）
- **Dockerfile** - 容器镜像定义（工具、环境变量等）

### 添加额外工具

在 `.devcontainer/Dockerfile` 中添加：

```dockerfile
RUN apk add --no-cache <package-name>
```

然后重新构建容器：**Dev Containers: Rebuild Container**

## 🐛 故障排查

### 容器构建失败

1. 检查 Docker 是否正常运行
2. 清理缓存并重建：
   ```bash
   docker system prune -f
   ```
3. 重新打开容器：**Dev Containers: Rebuild and Reopen in Container**

### Go 工具未安装

手动安装：
```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Kubeconfig 权限问题

确保本地 kubeconfig 文件可读：
```bash
chmod 600 ~/.kube/config
```

## 📚 更多信息

- [Dev Containers 官方文档](https://code.visualstudio.com/docs/devcontainers/containers)
- [Dev Container 规范](https://containers.dev/)
- [项目 README](../README.md)
