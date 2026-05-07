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
./bin/k8s-configmap-sidecar \
  --namespaces=default,production \
  --label-selector=app=myapp,type=config \
  --output-dir=/etc/config \
  --log-level=info

# 或使用配置文件
./bin/k8s-configmap-sidecar --config=/etc/sidecar/config.yaml

# 使用 kubeconfig 文件（本地开发）
./bin/k8s-configmap-sidecar \
  --kubeconfig=$HOME/.kube/config \
  --namespaces=default \
  --label-selector=app=myapp
```

### 3. Docker 构建

```bash
# 构建镜像
make docker-build VERSION=latest

# 或直接使用 Dockerfile
docker build -t k8s-configmap-sidecar:latest .
```

### 4. Kubernetes 部署

```bash
# 应用 RBAC
kubectl apply -f examples/rbac.yaml

# 部署应用
kubectl apply -f examples/deployment.yaml
```

## 配置说明

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

### 环境变量

| 变量名 | 说明 | 默认值 |
|-------|------|--------|
| `KUBECONFIG` | kubeconfig 路径 | `""` |
| `NAMESPACES` | 命名空间列表 | `default` |
| `LABEL_SELECTOR` | Label 选择器（JSON） | 必填 |
| `OUTPUT_DIR` | 输出目录 | `/etc/config` |
| `LOG_LEVEL` | 日志级别 | `info` |

### 配置文件示例

```
kubeconfig: ""
namespaces:
  - default
  - production
labelSelector:
  app: myapp
  type: config
outputDir: "/etc/config"
resyncPeriod: "10m"
logLevel: "info"
```

## 文件同步规则

### 目录结构

```
{OutputDir}/
├── {namespace}/
│   └── {configmap-name}/
│       ├── key1 → 文件内容
│       ├── key2 → 文件内容
│       └── ...
```

### 示例

ConfigMap:
```yaml
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
└── default/
    └── my-config/
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
