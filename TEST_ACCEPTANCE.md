# 测试验收计划

## 测试环境准备

### 1. Kubernetes 集群
- [ ] Minikube / Kind / K3s 本地集群
- [ ] 或远程开发/测试集群
- [ ] kubectl 配置正确

### 2. 构建环境
```bash
cd /data/fbdn/k8s-sidecar
export GOPROXY=https://goproxy.io,direct
go mod download
make build
```

## 单元测试阶段 ✅

### 配置模块测试
```bash
go test -v ./internal/config/...
```

**验收标准**:
- [x] TestLoadConfig - 配置文件加载
- [x] TestBuildLabelSelectorString - Label 转换
- [x] TestIsAllNamespaces - 全命名空间检测
- [x] TestValidateConfig - 配置验证

### 文件同步测试
```bash
go test -v ./internal/sync/...
```

**验收标准**:
- [x] TestFileSyncService_SyncConfigMap - ConfigMap 同步
- [x] TestFileSyncService_DeleteConfigMap - 删除操作
- [x] TestFileSyncService_AtomicWrite - 原子写入

## 集成测试阶段

### 测试 1: 基本功能测试

#### 1.1 创建测试环境
```bash
# 创建测试命名空间
kubectl create namespace test-sidecar

# 应用 RBAC
kubectl apply -f examples/rbac.yaml -n test-sidecar
```

#### 1.2 创建测试 ConfigMap
```bash
kubectl apply -n test-sidecar -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config-1
  labels:
    app: test-app
    type: config
data:
  config.yaml: |
    server:
      port: 8080
      host: 0.0.0.0
  settings.json: |
    {
      "debug": true,
      "logLevel": "info"
    }
EOF
```

#### 1.3 本地运行 Sidecar
```
mkdir -p /tmp/test-sidecar-output

./bin/k8s-sidecar \
  --kubeconfig=$HOME/.kube/config \
  --namespaces=test-sidecar \
  --label-selector=app=test-app,type=config \
  --output-dir=/tmp/test-sidecar-output \
  --log-level=debug
```

**验收标准**:
- [ ] Sidecar 成功启动
- [ ] 日志显示 Informer 启动成功
- [ ] 检测到测试 ConfigMap
- [ ] 文件正确写入到 `/tmp/test-sidecar-output/test-sidecar/test-config-1/`

#### 1.4 验证文件内容
```
# 检查目录结构
ls -la /tmp/test-sidecar-output/

# 验证文件内容
cat /tmp/test-sidecar-output/config.yaml
cat /tmp/test-sidecar-output/settings.json
```

**验收标准**:
- [ ] 文件直接在输出根目录: `{keys}`
- [ ] 文件内容与 ConfigMap data 一致
- [ ] 文件格式正确

### 测试 2: 动态更新测试

#### 2.1 更新 ConfigMap
```bash
kubectl edit configmap test-config-1 -n test-sidecar
```

修改 `port: 8080` 为 `port: 9090`

#### 2.2 观察日志
应该看到:
```json
{"level":"info","message":"ConfigMap updated event received",...}
{"level":"info","message":"ConfigMap synced successfully","files_updated":2,...}
```

#### 2.3 验证文件更新
```bash
cat /tmp/test-sidecar-output/test-sidecar/test-config-1/config.yaml
```

**验收标准**:
- [ ] 检测到 ConfigMap 变化
- [ ] 文件内容已更新
- [ ] 无错误日志

### 测试 3: 删除测试

#### 3.1 删除 ConfigMap
```bash
kubectl delete configmap test-config-1 -n test-sidecar
```

#### 3.2 观察日志
应该看到:
```json
{"level":"info","message":"ConfigMap deleted event received",...}
{"level":"info","message":"ConfigMap directory deleted",...}
```

#### 3.3 验证目录删除
```bash
ls -la /tmp/test-sidecar-output/test-sidecar/
```

**验收标准**:
- [ ] 检测到删除事件
- [ ] 对应目录已被删除
- [ ] 无残留文件

### 测试 4: 多命名空间测试

#### 4.1 创建第二个命名空间
```bash
kubectl create namespace test-sidecar-2
kubectl apply -f examples/rbac.yaml -n test-sidecar-2

# 在第二个命名空间创建 ConfigMap
kubectl apply -n test-sidecar-2 -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config-2
  labels:
    app: test-app
    type: config
data:
  app.conf: |
    port = 7070
EOF
```

#### 4.2 多命名空间测试

```bash
./bin/k8s-sidecar \
  --kubeconfig=$HOME/.kube/config \
  --namespaces=test-sidecar,test-sidecar-2 \
  --label-selector=app=test-app,type=config \
  --output-dir=/tmp/test-multi-ns \
  --log-level=debug
```

**验收标准**:
- [ ] 两个命名空间的 Informer 都启动
- [ ] 两个 ConfigMap 都被同步
- [ ] 所有文件在输出根目录（注意：扁平结构下不同ConfigMap的同名key会覆盖）

#### 4.3 验证目录结构
```bash
tree /tmp/test-multi-ns/
```

预期输出:
```
/tmp/test-multi-ns/
├── test-sidecar/
│   └── test-config-1/
│       ├── config.yaml
│       └── settings.json
└── test-sidecar-2/
    └── test-config-2/
        └── app.conf
```

### 测试 5: Label 选择器测试

#### 5.1 创建不同 label 的 ConfigMap
```bash
# 这个应该被监控
kubectl apply -n test-sidecar -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: matching-config
  labels:
    app: test-app
    type: config
data:
  matched.txt: "yes"
EOF

# 这个不应该被监控
kubectl apply -n test-sidecar -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: non-matching-config
  labels:
    app: other-app
    type: config
data:
  matched.txt: "no"
EOF
```

**验收标准**:
- [ ] 只有匹配 label 的 ConfigMap 被同步
- [ ] `matching-config` 出现在输出目录
- [ ] `non-matching-config` 不在输出目录

### 测试 6: In-Cluster 模式测试

#### 6.1 Docker 镜像测试

```bash
# 构建测试镜像
docker build -t k8s-sidecar:test .
```

#### 6.2 部署到集群
```bash
kubectl apply -f examples/deployment.yaml -n test-sidecar
```

#### 6.3 验证 Pod 运行
```bash
kubectl get pods -n test-sidecar -l app=myapp
kubectl logs <pod-name> -c configmap-sidecar -n test-sidecar
```

**验收标准**:
- [ ] Pod 正常启动
- [ ] 使用 In-Cluster 配置连接成功
- [ ] ConfigMap 同步正常工作

### 测试 7: 压力测试

#### 7.1 批量创建 ConfigMap
```bash
for i in {1..50}; do
  kubectl apply -n test-sidecar -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: stress-test-$i
  labels:
    app: stress-test
    type: config
data:
  file$i.txt: "Content of config $i"
  data$i.yaml: |
    key: value$i
    number: $i
EOF
done
```

#### 7.2 观察性能
```bash
# 监控内存和 CPU
kubectl top pod <pod-name> -n test-sidecar

# 观察日志
kubectl logs -f <pod-name> -c configmap-sidecar -n test-sidecar
```

**验收标准**:
- [ ] 所有 ConfigMap 都被同步
- [ ] 内存增长在合理范围
- [ ] API Server QPS 在限制内
- [ ] 无明显延迟

### 测试 8: 故障恢复测试

#### 8.1 模拟网络中断
（可选，需要测试环境支持）

#### 8.2 重启 Sidecar
```bash
# 找到 Pod
POD=$(kubectl get pods -n test-sidecar -l app=myapp -o jsonpath='{.items[0].metadata.name}')

# 重启 sidecar 容器
kubectl delete pod $POD -n test-sidecar
```

#### 8.3 验证自动恢复
```bash
# 新 Pod 启动后检查文件
kubectl exec <new-pod> -c main-app -- ls -la /etc/config/
```

**验收标准**:
- [ ] 重启后自动重新同步
- [ ] 数据完整性保持
- [ ] 无数据丢失

## 性能测试

```bash
# 测量启动时间
time ./bin/k8s-sidecar --help

# 检查内存使用
ps aux | grep k8s-sidecar
```

## 验收检查清单

### 功能验收
- [ ] ✅ Label 选择器动态配置
- [ ] ✅ Key 作为文件名
- [ ] ✅ 覆盖写入策略
- [ ] ✅ 默认 Resync 周期（10分钟）
- [ ] ✅ 多命名空间支持
- [ ] ✅ In-Cluster 模式
- [ ] ✅ kubeconfig 模式
- [ ] ✅ 原子文件写入
- [ ] ✅ Informer 缓存机制

### 质量验收
- [ ] 代码通过 `go vet` 检查
- [ ] 代码通过 `go fmt` 格式化
- [ ] 单元测试通过率 100%
- [ ] 无明显的 goroutine 泄漏
- [ ] 错误处理完善

### 文档验收
- [x] DESIGN.md - 设计文档
- [x] README.md - 项目文档
- [x] QUICKSTART.md - 快速开始
- [x] PROJECT_SUMMARY.md - 项目总结
- [x] examples/ - 示例配置
- [x] TEST_ACCEPTANCE.md - 测试文档

### 部署验收
- [ ] Docker 镜像构建成功
- [ ] K8s 部署示例可运行
- [ ] RBAC 配置正确
- [ ] 日志输出结构化

## 问题记录

在测试过程中发现的问题：

| 序号 | 问题描述 | 严重程度 | 解决状态 | 备注 |
|-----|---------|---------|---------|------|
| 1 | | | | |
| 2 | | | | |
| 3 | | | | |

## 测试结论

**测试日期**: ____________

**测试人员**: ____________

**测试结果**: ☐ 通过  ☐ 不通过

**总体评价**:
_______________________________________________
_______________________________________________
_______________________________________________

**签字**: ____________

---

*注: 请按照此计划逐项进行测试，并记录测试结果*
