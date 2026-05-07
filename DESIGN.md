# K8s ConfigMap Sidecar 设计文档

## 1. 项目概述

### 1.1 项目背景
K8s ConfigMap Sidecar 是一个运行在 Kubernetes Pod 中的辅助容器，用于动态监控指定命名空间下带有特定 label 的 ConfigMap 资源，并将其数据自动同步到本地文件系统中。这使得主应用能够以文件方式访问 ConfigMap 配置，而无需重启或重新加载。

### 1.2 核心功能
- 基于 Label 选择器动态发现 ConfigMap
- 实时监听 ConfigMap 变化并同步到文件系统
- 支持多命名空间监控
- 支持 In Cluster 和 kubeconfig 两种认证模式
- 使用 Informer 机制减少 API Server 压力

### 1.3 技术栈

- **语言**: Golang 1.25+
- **K8s 库**: client-go v0.34.0, api v0.34.0, apimachinery v0.34.0
- **日志**: zap v1.26.0
- **配置**: yaml.v3 v3.0.1
- **运行环境**: Kubernetes 1.34+

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────┐
│           Pod                           │
│                                         │
│  ┌──────────────┐    ┌──────────────┐  │
│  │   Main App   │    │   Sidecar    │  │
│  │   Container  │    │  Container   │  │
│  │              │    │              │  │
│  │              │◄───│  File Sync   │  │
│  └──────────────┘    └──────┬───────┘  │
│                             │          │
└─────────────────────────────┼──────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Kubernetes API   │
                    │      Server       │
                    │                   │
                    │  ┌─────────────┐  │
                    │  │ ConfigMap   │  │
                    │  │ (labeled)   │  │
                    │  └─────────────┘  │
                    └───────────────────┘
```

### 2.2 组件设计

```
┌─────────────────────────────────────────────────────┐
│                 Sidecar 内部架构                     │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌──────────────┐                                   │
│  │   Config     │ 配置管理                           │
│  └──────┬───────┘                                   │
│         │                                            │
│  ┌──────▼───────┐                                   │
│  │   Client     │ K8s 客户端初始化                  │
│  └──────┬───────┘                                   │
│         │                                            │
│  ┌──────▼────────────────┐                          │
│  │   Informer Manager    │ 多命名空间 Informer 管理  │
│  │  ┌─────────────────┐  │                          │
│  │  │ Namespace A     │  │                          │
│  │  │ - ConfigMap     │  │                          │
│  │  │ - Event Handler │  │                          │
│  │  └─────────────────┘  │                          │
│  │  ┌─────────────────┐  │                          │
│  │  │ Namespace B     │  │                          │
│  │  │ - ConfigMap     │  │                          │
│  │  │ - Event Handler │  │                          │
│  │  └─────────────────┘  │                          │
│  └──────────┬────────────┘                          │
│             │                                        │
│  ┌──────────▼────────────┐                          │
│  │   File Sync Service   │ 文件同步服务              │
│  │  - Write/Update       │                          │
│  │  - Delete             │                          │
│  └──────────┬────────────┘                          │
│             │                                        │
│  ┌──────────▼────────────┐                          │
│  │   Logger & Metrics    │ 日志和监控                │
│  └───────────────────────┘                          │
└─────────────────────────────────────────────────────┘
```

## 3. 详细设计

### 3.1 配置管理

#### 3.1.1 配置结构
```go
type Config struct {
    // KubeConfig kubeconfig 文件路径（可选，为空时使用 In Cluster 模式）
    KubeConfig string `json:"kubeconfig,omitempty"`
    
    // Namespaces 要监控的命名空间列表
    // 支持特殊值 "*" 表示所有命名空间
    Namespaces []string `json:"namespaces"`
    
    // LabelSelector label 选择器
    LabelSelector map[string]string `json:"labelSelector"`
    
    // OutputDir 配置文件输出目录
    OutputDir string `json:"outputDir" default:"/etc/config"`
    
    // ResyncPeriod Informer 重新同步周期（默认 10 分钟）
    ResyncPeriod time.Duration `json:"resyncPeriod" default:"10m"`
}
```

#### 3.1.2 配置来源优先级
1. 环境变量（最高优先级）
2. 配置文件（`/etc/sidecar/config.yaml`）
3. 命令行参数
4. 默认值（最低优先级）

### 3.2 Kubernetes 客户端

#### 3.2.1 客户端初始化流程
```
开始
  │
  ├─ 检查是否提供 kubeconfig
  │   │
  │   ├─ 是 → 使用 kubeconfig 创建 REST Config
  │   │
  │   └─ 否 → 使用 In Cluster Config
  │
  ├─ 创建 Kubernetes ClientSet
  │
  └─ 验证连接
       │
       ├─ 成功 → 返回客户端实例
       │
       └─ 失败 → 返回错误
```

#### 3.2.2 关键实现
- 使用 `rest.InClusterConfig()` 获取集群内配置
- 使用 `clientcmd.BuildConfigFromFlags()` 从 kubeconfig 加载
- 设置合理的 QPS 和 Burst 限制（QPS: 5, Burst: 10）

### 3.3 Informer 管理器

#### 3.3.1 多命名空间 Informer 策略

**方案选择**: 为每个命名空间创建独立的 SharedInformer

**优势**:
- 权限隔离清晰
- 易于控制每个命名空间的同步行为
- 故障隔离，单个命名空间问题不影响其他

**实现**:
```go
type InformerManager struct {
    clientset  kubernetes.Interface
    config     *Config
    informers  map[string]cache.SharedIndexInformer // key: namespace
    stopChans  map[string]chan struct{}
}
```

#### 3.3.2 Informer 配置
- **ResyncPeriod**: 10 分钟（默认）
- **Indexers**: 使用默认索引
- **ListWatch**: 带 Label Selector 的 ListWatch

#### 3.3.3 Label Selector 转换
```go
// 将 map 转换为 Kubernetes Label Selector
func buildLabelSelector(labels map[string]string) string {
    var selectors []string
    for k, v := range labels {
        selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(selectors, ",")
}
```

### 3.4 事件处理

#### 3.4.1 事件类型和处理逻辑

| 事件类型 | 触发条件 | 处理逻辑 |
|---------|---------|---------|
| Add | 新的 ConfigMap 创建 | 将所有 data 写入文件系统 |
| Update | ConfigMap 更新 | 对比差异，覆盖变化的文件 |
| Delete | ConfigMap 删除 | 删除对应的文件目录 |

#### 3.4.2 文件同步策略

**目录结构**:
```
{OutputDir}/
├── {namespace}/
│   └── {configmap-name}/
│       ├── key1 → 文件内容
│       ├── key2 → 文件内容
│       └── ...
```

**文件写入规则**:
1. 每个 ConfigMap 对应一个子目录 `{namespace}/{configmap-name}`
2. ConfigMap 的每个 key 对应目录下的一个文件
3. 文件名 = key，文件内容 = data[key]
4. 覆盖写入，不备份旧文件
5. 原子写入（先写临时文件，再重命名）

#### 3.4.3 原子写入实现
```go
func atomicWriteFile(filePath string, content []byte, perm os.FileMode) error {
    // 1. 创建临时文件
    tmpFile, err := ioutil.TempFile(dir, ".tmp-*")
    if err != nil {
        return err
    }
    
    // 2. 写入内容
    _, err = tmpFile.Write(content)
    if err != nil {
        tmpFile.Close()
        os.Remove(tmpFile.Name())
        return err
    }
    tmpFile.Close()
    
    // 3. 设置权限
    err = os.Chmod(tmpFile.Name(), perm)
    if err != nil {
        os.Remove(tmpFile.Name())
        return err
    }
    
    // 4. 原子重命名
    return os.Rename(tmpFile.Name(), filePath)
}
```

### 3.5 错误处理和重试

#### 3.5.1 错误分类

| 错误类型 | 处理策略 |
|---------|---------|
| K8s API 连接失败 | 指数退避重试，记录日志 |
| 文件写入失败 | 立即重试 3 次，失败则记录错误 |
| 权限不足 | 记录致命错误，退出容器 |
| 磁盘空间不足 | 记录错误，等待人工干预 |

#### 3.5.2 重试机制
- 初始间隔: 1 秒
- 最大间隔: 60 秒
- 最大重试次数: 5 次
- 退避策略: 指数退避（1s → 2s → 4s → 8s → 16s）

### 3.6 日志和监控

#### 3.6.1 日志级别
- **DEBUG**: 详细的调试信息
- **INFO**: 正常操作信息（默认级别）
- **WARN**: 警告信息
- **ERROR**: 错误信息
- **FATAL**: 致命错误

#### 3.6.2 关键日志点
1. 启动时配置信息
2. Informer 启动/停止
3. ConfigMap 事件（Add/Update/Delete）
4. 文件写入成功/失败
5. 错误和异常

#### 3.6.3 日志格式
```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "INFO",
  "message": "ConfigMap synced successfully",
  "namespace": "default",
  "configmap": "my-config",
  "files_updated": 3
}
```

## 4. 接口设计

### 4.1 命令行参数

```bash
./k8s-configmap-sidecar [flags]
```

| 参数 | 简写 | 说明 | 默认值 |
|-----|------|------|--------|
| --kubeconfig | -k | kubeconfig 文件路径 | "" (使用 in-cluster) |
| --namespaces | -n | 命名空间列表（逗号分隔） | ["default"] |
| --label-selector | -l | Label 选择器（逗号分隔的 key=value） | 必填 |
| --output-dir | -o | 输出目录 | "/etc/config" |
| --resync-period | -r | Resync 周期 | "10m" |
| --log-level | -v | 日志级别 | "info" |
| --config | -c | 配置文件路径 | "/etc/sidecar/config.yaml" |

### 4.2 环境变量

| 变量名 | 说明 | 默认值 |
|-------|------|--------|
| KUBECONFIG | kubeconfig 路径 | "" |
| NAMESPACES | 命名空间列表 | "default" |
| LABEL_SELECTOR | Label 选择器（JSON 格式） | 必填 |
| OUTPUT_DIR | 输出目录 | "/etc/config" |
| RESYNC_PERIOD | Resync 周期 | "10m" |
| LOG_LEVEL | 日志级别 | "info" |

### 4.3 配置文件示例

```
# /etc/sidecar/config.yaml
kubeconfig: ""  # 空字符串表示使用 in-cluster 模式
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

## 5. 部署设计

### 5.1 Deployment 示例

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-with-sidecar
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: main-app
        image: myapp:latest
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
      
      - name: configmap-sidecar
        image: k8s-configmap-sidecar:latest
        args:
        - --namespaces=default,production
        - --label-selector=app=myapp,type=config
        - --output-dir=/etc/config
        env:
        - name: LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
      
      volumes:
      - name: config-volume
        emptyDir: {}
```

### 5.2 RBAC 配置

```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: configmap-watcher
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: configmap-watcher-binding
roleRef:
  kind: Role
  name: configmap-watcher
subjects:
- kind: ServiceAccount
  name: default
```

## 6. 性能优化

### 6.1 减少 API Server 压力

1. **使用 Informer 缓存机制**
   - 本地缓存所有 ConfigMap 数据
   - 通过 Watch 机制接收增量更新
   - 避免频繁 List 操作

2. **合理的 Resync 周期**
   - 默认 10 分钟，避免过于频繁的同步
   - 仅在必要时才重新同步

3. **高效的 EventHandler**
   - 异步处理文件写入操作
   - 批量处理多个事件

### 6.2 内存优化

- 只缓存必要的 ConfigMap 字段
- 及时清理已删除 ConfigMap 的缓存
- 限制日志文件大小

## 7. 安全设计

### 7.1 权限最小化
- 仅需要 ConfigMap 的 get/list/watch 权限
- 使用 Role 而非 ClusterRole（限定命名空间）
- 不使用 root 用户运行

### 7.2 容器安全
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: false  # 需要写入文件
  allowPrivilegeEscalation: false
```

## 8. 测试策略

### 8.1 单元测试
- 配置解析测试
- Label Selector 转换测试
- 文件写入功能测试
- 原子写入测试

### 8.2 集成测试
- 使用 fake client 测试 Informer 逻辑
- 模拟 ConfigMap 事件
- 验证文件同步正确性

### 8.3 E2E 测试
- 在真实 K8s 集群中测试
- 创建/更新/删除 ConfigMap
- 验证文件同步结果

## 9. 项目结构

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
├── pkg/
│   └── utils/
│       └── file_utils.go        # 文件工具
├── examples/
│   ├── deployment.yaml          # 部署示例
│   └── rbac.yaml               # RBAC 配置
├── test/
│   ├── unit/                   # 单元测试
│   └── e2e/                    # E2E 测试
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

## 10. 开发计划

### Phase 1: 核心功能（预计 2-3 天）
- [ ] 项目初始化和依赖安装
- [ ] 配置管理模块
- [ ] Kubernetes 客户端初始化
- [ ] 基础 Informer 实现
- [ ] 文件同步服务

### Phase 2: 增强功能（预计 1-2 天）
- [ ] 多命名空间支持
- [ ] 原子写入实现
- [ ] 错误处理和重试
- [ ] 日志系统

### Phase 3: 测试和优化（预计 1-2 天）
- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能优化
- [ ] 文档完善

### Phase 4: 部署和验收（预计 1 天）
- [ ] Docker 镜像构建
- [ ] E2E 测试
- [ ] 部署验证
- [ ] 验收测试

## 11. 关键技术点总结

### 11.1 Informer 机制
- 使用 SharedInformerFactory 创建带 Label Selector 的 Informer
- 注册 ResourceEventHandler 处理事件
- 利用本地缓存减少 API 调用

### 11.2 文件同步
- 原子写入保证数据一致性
- 覆盖写入简化实现
- 目录隔离避免冲突

### 11.3 多命名空间
- 为每个命名空间创建独立 Informer
- 统一管理生命周期
- 错误隔离

## 12. 风险和挑战

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| API Server 不可用 | 无法同步配置 | 使用本地缓存，重试机制 |
| 磁盘空间不足 | 文件写入失败 | 监控磁盘使用，告警 |
| ConfigMap 数量过多 | 内存占用高 | 限制监控范围，合理设置 label |
| 权限配置错误 | 无法访问 ConfigMap | 详细的错误提示，RBAC 文档 |

---

**文档版本**: v1.0  
**创建日期**: 2024  
**最后更新**: 2024
