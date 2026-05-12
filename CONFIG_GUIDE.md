# 配置管理指南

本文档详细说明 k8s-sidecar 的配置管理方式。

## 目录

- [概述](#概述)
- [配置优先级](#配置优先级)
- [配置方式](#配置方式)
  - [配置文件](#配置文件)
  - [环境变量](#环境变量)
  - [命令行参数](#命令行参数)
- [配置项说明](#配置项说明)
- [示例](#示例)
- [最佳实践](#最佳实践)

## 概述

k8s-sidecar 使用 [Viper](https://github.com/spf13/viper) 库进行配置管理，支持多种配置源和灵活的配置加载方式。

### 主要特性

- ✅ **多格式支持**: YAML、JSON、TOML
- ✅ **多配置源**: 配置文件、环境变量、命令行参数
- ✅ **自动优先级**: 智能合并多个配置源
- ✅ **默认值**: 合理的默认配置，减少配置工作量
- ✅ **类型安全**: 自动类型转换和验证
- ✅ **热重载准备**: 支持运行时配置更新（未来版本）

## 配置优先级

配置值的优先级从高到低：

1. **命令行参数** - 最高优先级，适合临时覆盖
2. **环境变量** - 中等优先级，适合容器化部署
3. **配置文件** - 基础配置，适合持久化存储
4. **默认值** - 最低优先级，确保程序可运行

```
命令行参数 > 环境变量 > 配置文件 > 默认值
```

## 配置方式

### 配置文件

创建 `config.yaml` 文件：

```yaml
kubeconfig: ""
namespaces:
  - default
  - production
labelSelector:
  app: grafana
  type: dashboard
outputDir: "/etc/config"
resyncPeriod: "10m"
logLevel: "info"
```

使用配置文件启动：

```bash
./k8s-sidecar --config=/path/to/config.yaml
```

**配置文件搜索路径**（按顺序）：

1. `/etc/k8s-sidecar/config.yaml`
2. `./config.yaml` (当前目录)
3. `$HOME/.k8s-sidecar/config.yaml`

如果不指定 `--config` 参数，程序会自动在上述位置查找配置文件。

### 环境变量

所有配置项都可以通过环境变量设置。支持两种格式：

#### 无前缀格式

```bash
export KUBECONFIG="/path/to/kubeconfig"
export NAMESPACES="default,production"
export LABEL_SELECTOR='{"app":"grafana","type":"dashboard"}'
export OUTPUT_DIR="/etc/config"
export RESYNC_PERIOD="10m"
export LOG_LEVEL="info"
```

#### SIDECAR_ 前缀格式（推荐）

```bash
export SIDECAR_KUBECONFIG="/path/to/kubeconfig"
export SIDECAR_NAMESPACES="default,production"
export SIDECAR_LABEL_SELECTOR='{"app":"grafana","type":"dashboard"}'
export SIDECAR_OUTPUT_DIR="/etc/config"
export SIDECAR_RESYNC_PERIOD="10m"
export SIDECAR_LOG_LEVEL="info"
```

### 命令行参数

```bash
./k8s-sidecar \
  --kubeconfig="/path/to/kubeconfig" \
  --namespaces="default,production" \
  --label-selector="app=grafana,type=dashboard" \
  --output-dir="/etc/config" \
  --resync-period="10m" \
  --log-level="info"
```

**注意**: 命令行参数会被转换为环境变量，因此具有最高优先级。

## 配置项说明

### kubeconfig

- **类型**: `string`
- **默认值**: `""` (空字符串)
- **环境变量**: `KUBECONFIG`, `SIDECAR_KUBECONFIG`
- **命令行参数**: `--kubeconfig`
- **说明**: Kubernetes kubeconfig 文件路径。如果为空，程序将使用 In-Cluster 模式（适用于 Pod 内部运行）。

### namespaces

- **类型**: `[]string` (字符串数组)
- **默认值**: `["default"]`
- **环境变量**: `NAMESPACES`, `SIDECAR_NAMESPACES` (逗号分隔)
- **命令行参数**: `--namespaces` (逗号分隔)
- **说明**: 要监控的命名空间列表。使用 `"*"` 表示所有命名空间。

**示例**:

```yaml
# YAML 格式
namespaces:
  - default
  - production
  - staging
```

```bash
# 环境变量格式
export SIDECAR_NAMESPACES="default,production,staging"

# 监控所有命名空间
export SIDECAR_NAMESPACES="*"
```

### labelSelector

- **类型**: `map[string]string` (键值对)
- **默认值**: `{}` (必须提供)
- **环境变量**: `LABEL_SELECTOR`, `SIDECAR_LABEL_SELECTOR` (JSON 格式)
- **命令行参数**: `--label-selector` (key=value 格式)
- **说明**: Label 选择器，用于过滤要同步的 ConfigMap。只有带有匹配标签的 ConfigMap 才会被同步。

**示例**:

```yaml
# YAML 格式
labelSelector:
  app: grafana
  type: dashboard
```

```bash
# 环境变量格式 (JSON)
export SIDECAR_LABEL_SELECTOR='{"app":"grafana","type":"dashboard"}'

# 命令行参数格式
./k8s-sidecar --label-selector="app=grafana,type=dashboard"
```

### outputDir

- **类型**: `string`
- **默认值**: `"/etc/config"`
- **环境变量**: `OUTPUT_DIR`, `SIDECAR_OUTPUT_DIR`
- **命令行参数**: `--output-dir`
- **说明**: 配置文件输出目录。Sidecar 会将 ConfigMap 的内容写入此目录。

**示例**:

```yaml
outputDir: "/var/lib/grafana/dashboards"
```

```bash
export SIDECAR_OUTPUT_DIR="/var/lib/grafana/dashboards"
```

### resyncPeriod

- **类型**: `duration` (时间间隔)
- **默认值**: `"10m"` (10 分钟)
- **环境变量**: `RESYNC_PERIOD`, `SIDECAR_RESYNC_PERIOD`
- **命令行参数**: `--resync-period`
- **说明**: Informer 重新同步周期。即使没有事件触发，Informer 也会定期重新同步以确保数据一致性。

**支持的时间单位**:

- `ns` - 纳秒
- `us` 或 `µs` - 微秒
- `ms` - 毫秒
- `s` - 秒
- `m` - 分钟
- `h` - 小时

**示例**:

```yaml
resyncPeriod: "5m"    # 5 分钟
resyncPeriod: "30s"   # 30 秒
resyncPeriod: "1h"    # 1 小时
```

```bash
export SIDECAR_RESYNC_PERIOD="5m"
```

### logLevel

- **类型**: `string`
- **默认值**: `"info"`
- **环境变量**: `LOG_LEVEL`, `SIDECAR_LOG_LEVEL`
- **命令行参数**: `--log-level`
- **说明**: 日志级别。控制日志输出的详细程度。

**可选值**:

- `debug` - 调试级别，输出最详细的日志
- `info` - 信息级别，输出常规操作日志（推荐）
- `warn` - 警告级别，只输出警告和错误
- `error` - 错误级别，只输出错误日志
- `fatal` - 致命级别，只输出致命错误

**示例**:

```yaml
logLevel: "debug"
```

```bash
export SIDECAR_LOG_LEVEL="debug"
```

## 示例

### 示例 1: 最小化配置

仅指定必需的 `labelSelector`，其他使用默认值：

```yaml
labelSelector:
  app: myapp
```

### 示例 2: Grafana Dashboard 同步

```yaml
namespaces:
  - monitoring
labelSelector:
  app: grafana
  type: dashboard
outputDir: "/var/lib/grafana/dashboards"
resyncPeriod: "5m"
logLevel: "info"
```

### 示例 3: 多环境配置

```yaml
namespaces:
  - production
  - staging
  - development
labelSelector:
  config-type: application
outputDir: "/etc/app-config"
resyncPeriod: "10m"
logLevel: "warn"
```

### 示例 4: 使用环境变量覆盖

基础配置文件 `config.yaml`:

```yaml
labelSelector:
  app: myapp
outputDir: "/etc/config"
```

通过环境变量覆盖：

```bash
export SIDECAR_NAMESPACES="production"
export SIDECAR_LOG_LEVEL="debug"
./k8s-sidecar --config=config.yaml
```

最终配置：
- `namespaces`: `["production"]` (来自环境变量)
- `labelSelector`: `{"app": "myapp"}` (来自配置文件)
- `outputDir`: `"/etc/config"` (来自配置文件)
- `logLevel`: `"debug"` (来自环境变量)
- 其他字段使用默认值

### 示例 5: Kubernetes Deployment 中的配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
        - name: sidecar
          image: k8s-sidecar:latest
          env:
            - name: SIDECAR_NAMESPACES
              value: "default"
            - name: SIDECAR_LABEL_SELECTOR
              value: '{"app":"myapp","type":"config"}'
            - name: SIDECAR_OUTPUT_DIR
              value: "/etc/config"
            - name: SIDECAR_LOG_LEVEL
              value: "info"
          volumeMounts:
            - name: config-volume
              mountPath: /etc/config
      volumes:
        - name: config-volume
          emptyDir: {}
```

## 最佳实践

### 1. 配置文件管理

- 将基础配置保存在版本控制系统中
- 使用 `.gitignore` 排除包含敏感信息的配置文件
- 为不同环境维护不同的配置文件（`config-dev.yaml`, `config-prod.yaml`）

### 2. 敏感信息处理

- **不要**在配置文件中存储敏感信息（如 kubeconfig 路径可能包含凭证）
- 使用 Kubernetes Secrets 或外部密钥管理系统
- 通过环境变量注入敏感配置

### 3. 环境变量命名

- 推荐使用 `SIDECAR_` 前缀，避免与其他应用冲突
- 在 CI/CD 流水线中使用环境变量配置不同环境

### 4. 配置验证

- 程序启动时会自动验证配置
- 必需字段缺失或无效会导致启动失败
- 查看错误日志快速定位配置问题

### 5. 日志级别选择

- **开发环境**: 使用 `debug` 级别，便于排查问题
- **生产环境**: 使用 `info` 或 `warn` 级别，减少日志量
- **故障排查**: 临时切换到 `debug` 级别

### 6. Resync Period 调优

- **默认值** (`10m`): 适合大多数场景
- **高频变更**: 缩短到 `5m` 或更短
- **低频变更**: 延长到 `30m` 或更长，减少 API Server 压力
- **注意**: 过短的周期会增加 API Server 负载

### 7. Namespace 选择

- **单命名空间**: 明确指定命名空间，遵循最小权限原则
- **多命名空间**: 使用列表指定多个命名空间
- **所有命名空间**: 使用 `"*"`，但需要 ClusterRole 权限

### 8. Label Selector 设计

- 使用有意义的标签组合
- 避免过于宽泛的选择器（如同步所有 ConfigMap）
- 建议使用 `app` + `type` 的组合标签

## 故障排查

### 问题 1: 配置未生效

**检查清单**:

1. 确认配置文件路径正确
2. 检查 YAML 语法是否正确
3. 验证环境变量是否设置
4. 查看启动日志中的配置摘要

**解决方案**:

```bash
# 启用 debug 日志查看详细配置
export SIDECAR_LOG_LEVEL="debug"
./k8s-sidecar --config=config.yaml
```

### 问题 2: Label Selector 解析失败

**常见原因**:

- JSON 格式错误
- 缺少引号或括号

**解决方案**:

```bash
# 正确的 JSON 格式
export SIDECAR_LABEL_SELECTOR='{"app":"myapp","type":"config"}'

# 或使用命令行参数（更简单）
./k8s-sidecar --label-selector="app=myapp,type=config"
```

### 问题 3: 命名空间未正确解析

**常见原因**:

- 逗号后有空格
- 使用了错误的分隔符

**解决方案**:

```bash
# 正确：无空格
export SIDECAR_NAMESPACES="default,production"

# 错误：有空格
export SIDECAR_NAMESPACES="default, production"
```

## 总结

k8s-sidecar 的配置管理系统提供了极大的灵活性：

- 🎯 **简单场景**: 使用默认值或最小化配置
- 🔧 **复杂场景**: 混合使用配置文件和环境变量
- 🚀 **生产环境**: 通过环境变量和 Kubernetes ConfigMap/Secrets 管理配置
- 🧪 **开发调试**: 使用命令行参数快速测试不同配置

充分利用 Viper 的强大功能，可以让配置管理变得简单而高效。
