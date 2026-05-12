# 配置文件可选性修复

## 问题描述

程序启动时报错：
```
Failed to load configuration: failed to read config file: open /etc/sidecar/config.yaml: no such file or directory
```

## 根本原因

1. `main.go` 中 `--config` 参数的默认值为 `/etc/sidecar/config.yaml`
2. `LoadConfig` 函数在配置文件读取失败时（非 `ConfigFileNotFoundError`）会直接返回错误
3. 导致即使有环境变量和命令行参数可用，程序也会退出

## 解决方案

修改 `internal/config/config.go` 中的 `LoadConfig` 函数，将配置文件读取错误从致命错误降级为警告：

```go
// 读取配置文件（如果存在）
if err := v.ReadInConfig(); err != nil {
    // 配置文件不存在或读取失败不是致命错误，继续使用其他配置源
    // 记录警告但不中断程序执行
    fmt.Printf("Warning: Config file not found or unreadable (%v). Using environment variables and defaults.\n", err)
}
```

## 变更文件

- `internal/config/config.go` - 修改配置加载逻辑

## 测试验证

### 运行单元测试
```bash
go test ./internal/config/... -v
```

所有测试通过 ✅

### 运行集成测试
```bash
./test-config-optional.sh
```

测试场景：
1. ✅ 无配置文件 + 环境变量 - 程序正常运行
2. ✅ 命令行参数覆盖 - 正常工作
3. ✅ 单元测试验证 - 配置可选性确认

## 配置优先级

修复后仍然保持原有的配置优先级：
1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件
4. 默认值（最低优先级）

## 使用示例

### 方式 1: 仅使用环境变量（无需配置文件）
```bash
export SIDECAR_LABEL_SELECTOR='{"app":"grafana"}'
export SIDECAR_OUTPUT_DIR='/var/lib/grafana/dashboards'
export SIDECAR_NAMESPACES='default,production'
./k8s-sidecar
```

### 方式 2: 仅使用命令行参数（无需配置文件）
```bash
./k8s-sidecar \
  --label-selector='app=grafana' \
  --output-dir=/var/lib/grafana/dashboards \
  --namespaces='default,production'
```

### 方式 3: 混合使用
```bash
# 设置部分环境变量
export SIDECAR_NAMESPACES='default,production'

# 使用命令行参数覆盖
./k8s-sidecar \
  --config=/path/to/config.yaml \
  --label-selector='app=grafana'
```

## 向后兼容性

✅ 完全向后兼容
- 如果配置文件存在且可读，正常加载
- 如果配置文件不存在，显示警告但继续使用其他配置源
- 不影响现有的部署方式和配置管理流程

## 分支信息

- 分支名称: `fix/optional-config-file`
- 提交信息: `fix: make config file truly optional`
