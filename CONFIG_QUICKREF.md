# 配置管理快速参考

## 快速开始

### 1. 最小化启动（仅使用默认值）

```bash
# 需要至少提供 labelSelector
export SIDECAR_LABEL_SELECTOR='{"app":"myapp"}'
./k8s-sidecar
```

### 2. 使用配置文件

```bash
# 创建 config.yaml
cat > config.yaml <<EOF
labelSelector:
  app: myapp
namespaces:
  - default
outputDir: /etc/config
EOF

# 启动
./k8s-sidecar --config=config.yaml
```

### 3. 使用环境变量

```bash
export SIDECAR_NAMESPACES="default,production"
export SIDECAR_LABEL_SELECTOR='{"app":"grafana","type":"dashboard"}'
export SIDECAR_OUTPUT_DIR="/var/lib/grafana/dashboards"
export SIDECAR_RESYNC_PERIOD="5m"
export SIDECAR_LOG_LEVEL="info"

./k8s-sidecar
```

### 4. 使用命令行参数

```bash
./k8s-sidecar \
  --namespaces="default,production" \
  --label-selector="app=grafana,type=dashboard" \
  --output-dir="/var/lib/grafana/dashboards" \
  --resync-period="5m" \
  --log-level="info"
```

## 环境变量速查表

| 配置项 | 环境变量（推荐） | 环境变量（兼容） | 默认值 |
|--------|-----------------|-----------------|--------|
| Kubeconfig | `SIDECAR_KUBECONFIG` | `KUBECONFIG` | `""` |
| Namespaces | `SIDECAR_NAMESPACES` | `NAMESPACES` | `["default"]` |
| Label Selector | `SIDECAR_LABEL_SELECTOR` | `LABEL_SELECTOR` | `{}` (必需) |
| Output Dir | `SIDECAR_OUTPUT_DIR` | `OUTPUT_DIR` | `/etc/config` |
| Resync Period | `SIDECAR_RESYNC_PERIOD` | `RESYNC_PERIOD` | `10m` |
| Log Level | `SIDECAR_LOG_LEVEL` | `LOG_LEVEL` | `info` |

## 命令行参数速查表

| 参数 | 说明 | 示例 |
|------|------|------|
| `--config` | 配置文件路径 | `--config=/etc/sidecar/config.yaml` |
| `--kubeconfig` | kubeconfig 文件路径 | `--kubeconfig=~/.kube/config` |
| `--namespaces` | 命名空间列表（逗号分隔） | `--namespaces="default,prod"` |
| `--label-selector` | Label 选择器 | `--label-selector="app=myapp"` |
| `--output-dir` | 输出目录 | `--output-dir=/etc/config` |
| `--resync-period` | 同步周期 | `--resync-period=10m` |
| `--log-level` | 日志级别 | `--log-level=debug` |
| `--version` | 显示版本 | `--version` |

## 常用场景

### Grafana Dashboard 同步

```bash
export SIDECAR_NAMESPACES="monitoring"
export SIDECAR_LABEL_SELECTOR='{"app":"grafana","type":"dashboard"}'
export SIDECAR_OUTPUT_DIR="/var/lib/grafana/dashboards"
export SIDECAR_RESYNC_PERIOD="5m"
./k8s-sidecar
```

### 多环境应用配置

```bash
export SIDECAR_NAMESPACES="prod,staging,dev"
export SIDECAR_LABEL_SELECTOR='{"config-type":"application"}'
export SIDECAR_OUTPUT_DIR="/etc/app-config"
export SIDECAR_LOG_LEVEL="warn"
./k8s-sidecar
```

### 开发调试模式

```bash
export SIDECAR_LABEL_SELECTOR='{"app":"test"}'
export SIDECAR_LOG_LEVEL="debug"
export SIDECAR_RESYNC_PERIOD="30s"
./k8s-sidecar
```

### 监控所有命名空间

```bash
export SIDECAR_NAMESPACES="*"
export SIDECAR_LABEL_SELECTOR='{"sync":"true"}'
./k8s-sidecar
```

**注意**: 使用 `"*"` 需要 ClusterRole 权限。

## 配置文件模板

完整的 `config.yaml` 模板：

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

# 同步周期
resyncPeriod: "10m"

# 日志级别
logLevel: "info"
```

## 配置优先级

```
命令行参数 > 环境变量 > 配置文件 > 默认值
```

**示例**: 如果同时在配置文件和环境中设置了 `outputDir`，环境变量的值会生效。

## 时间格式

`resyncPeriod` 支持的时间单位：

- `ns` - 纳秒
- `us` / `µs` - 微秒
- `ms` - 毫秒
- `s` - 秒
- `m` - 分钟
- `h` - 小时

**示例**:
- `30s` - 30 秒
- `5m` - 5 分钟
- `1h` - 1 小时

## 日志级别

- `debug` - 最详细，适合开发调试
- `info` - 常规操作（生产推荐）
- `warn` - 只警告和错误
- `error` - 只错误
- `fatal` - 只致命错误

## 故障排查

### 配置未生效？

```bash
# 1. 检查环境变量是否正确设置
echo $SIDECAR_OUTPUT_DIR

# 2. 启用 debug 日志查看实际配置
export SIDECAR_LOG_LEVEL="debug"
./k8s-sidecar --config=config.yaml

# 3. 确认配置文件路径正确
ls -l /path/to/config.yaml
```

### Label Selector 解析失败？

```bash
# 确保 JSON 格式正确（使用单引号包裹）
export SIDECAR_LABEL_SELECTOR='{"app":"myapp","type":"config"}'

# 或使用命令行参数（更简单）
./k8s-sidecar --label-selector="app=myapp,type=config"
```

### 命名空间未正确解析？

```bash
# 确保使用逗号分隔，无空格
export SIDECAR_NAMESPACES="default,production"

# 不要这样（有空格）
export SIDECAR_NAMESPACES="default, production"  # ❌
```

## 更多信息

- 📖 [完整配置指南](./CONFIG_GUIDE.md)
- 📝 [优化总结](./CONFIG_OPTIMIZATION_SUMMARY.md)
- 📋 [配置示例](./examples/config.yaml.example)
