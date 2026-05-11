# K8s ConfigMap Sidecar

一个基于 Kubernetes Informer 机制的 ConfigMap 同步工具，以 Sidecar 容器模式运行，自动将 ConfigMap 同步到本地文件系统。

## 功能特性

- ✅ **高效监听**: 使用 Kubernetes Informer 机制，减少 API Server 压力
- ✅ **Label 选择器**: 基于 Label 动态发现和监控 ConfigMap
- ✅ **多命名空间**: 支持监控多个或所有命名空间
- ✅ **原子写入**: 使用临时文件 + rename 确保文件一致性
- ✅ **双模式认证**: 支持 In-Cluster 和 kubeconfig 两种模式
- ✅ **实时同步**: ConfigMap 变化时自动更新文件系统
- ✅ **结构化日志**: JSON 格式日志，便于日志收集和分析

## 前置要求

- **Go 1.25+**
- Kubernetes 集群（用于测试）或本地 kubeconfig
- Docker（可选，用于容器化部署）

### 💻 使用 DevContainer（推荐）

本项目提供完整的 DevContainer 配置，可以快速启动一致的开发环境：

1. 安装 [VS Code](https://code.visualstudio.com/) 和 [Dev Containers 扩展](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. 在 VS Code 中打开项目
3. 按 `F1`，选择 **Dev Containers: Reopen in Container**
4. 等待容器构建完成即可开始开发

详细文档请参考：[.devcontainer/README.md](.devcontainer/README.md)

## 架构设计

```
┌─────────────────────────────────────────┐
│           Pod                           │
│                                         │
│  ┌──────────────┐    ┌──────────────┐  │
│  │   Main App   │    │   Sidecar    │  │
│  │   Container  │    │  Container   │  │
│  │              │◄───│  File Sync   │  │
│  └──────────────┘    └──────┬───────┘  │
└─────────────────────────────┼──────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Kubernetes API   │
                    │      Server       │
                    └───────────────────┘
```

## 快速开始

### 1. 构建

```bash
# 下载依赖
make deps

# 编译
make build
```

### 2. 运行

```bash
# 使用命令行参数
./bin/k8s-sidecar \
  --namespaces=default,production \
  --label-selector=app=myapp,type=config \
  --output-dir=/etc/config \
  --log-level=info

# 或使用配置文件
./bin/k8s-sidecar --config=/etc/sidecar/config.yaml

# 使用 kubeconfig 文件（本地开发）
./bin/k8s-sidecar \
  --kubeconfig=$HOME/.kube/config \
  --namespaces=default \
  --label-selector=app=myapp
```

### 3. Docker 构建

```bash
# 构建镜像
make docker-build VERSION=latest

# 或直接使用 Dockerfile
docker build -t k8s-sidecar:latest .
```

### 4. Kubernetes 部署

#### 通用应用配置同步

```bash
# 应用 RBAC
kubectl apply -f examples/rbac.yaml

# 部署应用
kubectl apply -f examples/deployment.yaml
```

#### 📊 Grafana Dashboard 自动加载（推荐）

使用 sidecar 自动同步 Grafana Dashboard ConfigMap 到文件系统：

```bash
# 一键部署 Grafana + Sidecar
kubectl apply -f examples/rbac-grafana.yaml
kubectl apply -f examples/grafana-dashboard.yaml
kubectl apply -f examples/deployment-grafana.yaml

# 访问 Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

**特性**：
- ✅ 自动检测 Dashboard ConfigMap 变化
- ✅ 实时同步到 Grafana 文件系统
- ✅ 支持多命名空间、多环境
- ✅ 零停机更新 Dashboard

📖 **详细文档**: [Grafana 集成指南](examples/GRAFANA_INTEGRATION.md) | [快速参考](examples/GRAFANA_QUICKREF.md)

### 5. CI/CD 自动化构建

项目已配置 GitLab CI/CD Pipeline，支持自动构建和推送多架构镜像：

```bash
# 推送代码触发流水线
git push origin main

# 发布版本（自动打标签）
git tag v1.0.0
git push origin v1.0.0
```

**中国镜像加速**: 项目已内置完整的中国镜像加速配置，显著提升构建速度：
- Go 模块代理：goproxy.cn + 阿里云
- Alpine APK 镜像：mirrors.aliyun.com
- Docker 镜像加速：阿里云容器镜像服务

详细配置请参考：
- 📖 [GitLab CI/CD 指南](GITLAB_CI_GUIDE.md)
- ⚡ [快速参考](GITLAB_CI_QUICKREF.md)
- ✅ [实施清单](GITLAB_CI_CHECKLIST.md)
- 🚀 [加速配置说明](CI_CD_ACCELERATION.md)

## 配置说明

k8s-sidecar 使用 [Viper](https://github.com/spf13/viper) 进行配置管理，支持多种配置源和灵活的优先级机制。

### 📚 详细文档

- 📖 **[完整配置指南](CONFIG_GUIDE.md)** - 详细的配置说明、示例和最佳实践
- ⚡ **[快速参考](CONFIG_QUICKREF.md)** - 常用配置速查表
- 📋 **[配置示例](examples/config.yaml.example)** - 完整的 YAML 配置模板
- 🔧 **[优化总结](CONFIG_OPTIMIZATION_SUMMARY.md)** - Viper 重构说明

### 配置优先级

```
命令行参数 > 环境变量 > 配置文件 > 默认值
```

### 命令行参数

| 参数 | 简写 | 说明 | 默认值 |
|-----|------|------|--------|
| `--kubeconfig` | `-k` | kubeconfig 文件路径 | `""` (in-cluster) |
| `--namespaces` | `-n` | 命名空间列表（逗号分隔） | `default` |
| `--label-selector` | `-l` | Label 选择器 | 必填 |
| `--output-dir` | `-o` | 输出目录 | `/etc/config` |
| `--resync-period` | `-r` | Resync 周期 | `10m` |
| `--log-level` | `-v` | 日志级别 | `info` |
| `--config` | `-c` | 配置文件路径 | `/etc/sidecar/config.yaml` |

**示例**:
```bash
./bin/k8s-sidecar \
  --namespaces="default,production" \
  --label-selector="app=myapp,type=config" \
  --output-dir="/etc/config" \
  --log-level="info"
```

### 环境变量

支持两种格式的环境变量：

#### 推荐格式（带 SIDECAR_ 前缀）

| 变量名 | 说明 | 默认值 |
|-------|------|--------|
| `SIDECAR_KUBECONFIG` | kubeconfig 路径 | `""` |
| `SIDECAR_NAMESPACES` | 命名空间列表（逗号分隔） | `default` |
| `SIDECAR_LABEL_SELECTOR` | Label 选择器（JSON） | 必填 |
| `SIDECAR_OUTPUT_DIR` | 输出目录 | `/etc/config` |
| `SIDECAR_RESYNC_PERIOD` | Resync 周期 | `10m` |
| `SIDECAR_LOG_LEVEL` | 日志级别 | `info` |

#### 兼容格式（无前缀）

| 变量名 | 说明 |
|-------|------|
| `KUBECONFIG` | kubeconfig 路径 |
| `NAMESPACES` | 命名空间列表 |
| `LABEL_SELECTOR` | Label 选择器（JSON） |
| `OUTPUT_DIR` | 输出目录 |
| `RESYNC_PERIOD` | Resync 周期 |
| `LOG_LEVEL` | 日志级别 |

**示例**:
```bash
export SIDECAR_NAMESPACES="default,production"
export SIDECAR_LABEL_SELECTOR='{"app":"myapp","type":"config"}'
export SIDECAR_OUTPUT_DIR="/etc/config"
export SIDECAR_RESYNC_PERIOD="5m"
export SIDECAR_LOG_LEVEL="debug"

./bin/k8s-sidecar
```

### 配置文件

程序会自动在以下位置查找配置文件（按顺序）：

1. `/etc/k8s-sidecar/config.yaml`
2. `./config.yaml` (当前目录)
3. `$HOME/.k8s-sidecar/config.yaml`

也可以通过 `--config` 参数指定路径。

**支持的格式**: YAML、JSON、TOML

**完整示例** ([查看模板](examples/config.yaml.example)):

```yaml
# Kubernetes kubeconfig 路径（空表示 In-Cluster 模式）
kubeconfig: ""

# 要监控的命名空间
namespaces:
  - default
  - production

# Label 选择器
labelSelector:
  app: myapp
  type: config

# 输出目录
outputDir: "/etc/config"

# 同步周期（支持 ns, us, ms, s, m, h）
resyncPeriod: "10m"

# 日志级别（debug, info, warn, error, fatal）
logLevel: "info"
```

### 混合使用示例

可以在配置文件中设置基础配置，通过环境变量覆盖特定值：

**config.yaml**:
```yaml
labelSelector:
  app: myapp
outputDir: "/etc/config"
```

**启动命令**:
```bash
export SIDECAR_NAMESPACES="production"
export SIDECAR_LOG_LEVEL="debug"
./bin/k8s-sidecar --config=config.yaml
```

最终配置：
- `namespaces`: `["production"]` (来自环境变量)
- `labelSelector`: `{"app": "myapp"}` (来自配置文件)
- `outputDir`: `"/etc/config"` (来自配置文件)
- `logLevel`: `"debug"` (来自环境变量)
- 其他字段使用默认值

## 文件同步规则

### 目录结构

```
{OutputDir}/
├── key1 → 文件内容 (来自 ConfigMap)
├── key2 → 文件内容 (来自 ConfigMap)
└── ...
```

### 示例

ConfigMap:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  app.conf: |
    port = 8080
  config.json: |
    {"debug": true}
```

同步后的文件:
```
/etc/config/
├── app.conf
└── config.json
```

## 开发

### 项目结构

```
k8s-sidecar/
├── cmd/
│   └── sidecar/
│       └── main.go              # 主入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── client/
│   │   └── kubernetes.go        # K8s 客户端
│   ├── informer/
│   │   └── manager.go           # Informer 管理器
│   ├── sync/
│   │   └── file_sync.go         # 文件同步服务
│   └── logger/
│       └── logger.go            # 日志
├── examples/
│   ├── deployment.yaml          # 部署示例
│   └── rbac.yaml               # RBAC 配置
├── go.mod
├── Makefile
└── Dockerfile
```

### 运行测试

```bash
# 单元测试
make test

# 格式化代码
make fmt

# 代码检查
make vet
```

## 工作原理

1. **初始化**: 加载配置，创建 Kubernetes 客户端
2. **Informer 启动**: 为每个命名空间创建带 Label Selector 的 ConfigMap Informer
3. **事件监听**: 
   - Add/Update: 将 ConfigMap 的 data 写入文件系统
   - Delete: 删除对应的文件目录
4. **文件同步**: 使用原子写入（临时文件 → 重命名）确保一致性

## 性能优化

- **Informer 缓存**: 本地缓存 ConfigMap 数据，减少 API 调用
- **Resync 周期**: 默认 10 分钟，避免频繁同步
- **QPS 限制**: QPS=5, Burst=10，保护 API Server
- **增量更新**: 仅在内容变化时写入文件

## 安全

- **最小权限**: 仅需 ConfigMap 的 get/list/watch 权限
- **非 Root 运行**: 容器以非 root 用户运行
- **RBAC**: 使用 Role 而非 ClusterRole（限定命名空间）

## 故障排查

### 查看日志

```bash
kubectl logs <pod-name> -c configmap-sidecar
```

### 常见问题

**Q: ConfigMap 没有同步？**
- 检查 Label 是否正确
- 验证 RBAC 权限
- 查看日志确认 Informer 是否启动

**Q: 文件写入失败？**
- 检查磁盘空间
- 验证目录权限
- 查看错误日志

## License

MIT

## Contributing

欢迎提交 Issue 和 Pull Request！
