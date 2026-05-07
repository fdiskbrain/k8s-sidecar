# 快速开始指南

## 前置要求

- **Go 1.25+**
- Kubernetes 集群（用于测试）或本地 kubeconfig
- Docker（可选，用于容器化部署）

## 步骤 1: 下载依赖并构建

```bash
cd /data/fbdn/k8s-sidecar

# 配置 Go 代理（如果在中国）
export GOPROXY=https://goproxy.io,direct

# 下载依赖
go mod download
go mod tidy

# 构建项目
make build
```

或者使用构建脚本：
```bash
./build.sh
```

## 步骤 2: 本地测试运行

### 2.1 查看帮助

```bash
./bin/k8s-sidecar --help
```

### 2.2 运行 Sidecar

#### 方式 1: 使用命令行参数

```bash
./bin/k8s-sidecar \
  --kubeconfig=$HOME/.kube/config \
  --namespaces=default \
  --label-selector=app=myapp,type=config \
  --output-dir=/tmp/test-config \
  --log-level=debug
```

#### 方式 2: 使用配置文件

```
# 创建配置目录
mkdir -p /etc/sidecar

# 复制示例配置
cp examples/config.yaml.example /etc/sidecar/config.yaml

# 编辑配置
vim /etc/sidecar/config.yaml

# 运行
./bin/k8s-sidecar --config=/etc/sidecar/config.yaml
```

### 方式 3: 使用环境变量

```bash
export KUBECONFIG=$HOME/.kube/config
export NAMESPACES="default,production"
export LABEL_SELECTOR='{"app":"myapp","type":"config"}'
export OUTPUT_DIR="/tmp/test-config"
export LOG_LEVEL="info"

./bin/k8s-sidecar
```

## 步骤 3: 在 Kubernetes 中部署

### 3.1 准备测试用的 ConfigMap

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
  labels:
    app: myapp
    type: config
data:
  app.conf: |
    [server]
    port = 8080
    host = 0.0.0.0
  config.json: |
    {
      "debug": true,
      "timeout": "30s"
    }
EOF
```

### 3.2 应用 RBAC

```bash
kubectl apply -f examples/rbac.yaml
```

### 3.3 部署应用

```bash
kubectl apply -f examples/deployment.yaml
```

### 3.4 验证运行状态

```bash
# 查看 Pod 状态
kubectl get pods -l app=myapp

# 查看 Sidecar 日志
kubectl logs <pod-name> -c configmap-sidecar

# 进入主容器查看配置文件（现在文件直接在输出根目录）
kubectl exec -it <pod-name> -c main-app -- ls -la /etc/config/
kubectl exec -it <pod-name> -c main-app -- cat /etc/config/app.conf
```

## 步骤 4: 测试动态更新

### 更新 ConfigMap

```bash
kubectl edit configmap test-config
```

修改内容，例如将 `port = 8080` 改为 `port = 9090`

### 观察日志

```bash
# 在另一个终端观察 sidecar 日志
kubectl logs -f <pod-name> -c configmap-sidecar
```

应该能看到类似这样的日志：
```json
{
  "level": "info",
  "timestamp": "2024-01-01T00:00:00Z",
  "message": "ConfigMap updated event received",
  "namespace": "default",
  "configmap": "test-config"
}
{
  "level": "info",
  "timestamp": "2024-01-01T00:00:00Z",
  "message": "ConfigMap synced successfully",
  "namespace": "default",
  "configmap": "test-config",
  "files_updated": 2,
  "duration": 15000000
}
```

### 验证文件已更新

```
kubectl exec -it <pod-name> -c main-app -- cat /etc/config/app.conf
```

应该看到更新后的内容。

## 步骤 5: 运行单元测试

```bash
# 运行所有测试
make test

# 或
./test.sh

# 查看覆盖率
go test -cover ./...
```

## 故障排查

### 问题 1: 无法连接 API Server

**症状**: `Failed to connect to Kubernetes API server`

**解决**:
```bash
# 检查 kubeconfig
kubectl cluster-info

# 如果在 Pod 内运行，检查 ServiceAccount
kubectl get sa
kubectl describe role configmap-watcher
```

### 问题 2: ConfigMap 没有同步

**症状**: 文件目录为空或没有更新

**检查清单**:
- Label 选择器是否正确？
- RBAC 权限是否足够？
- Informer 是否启动成功？

```bash
# 查看详细日志
kubectl logs <pod-name> -c configmap-sidecar | jq 'select(.level == "error")'
```

### 问题 3: 权限错误

**症状**: `configmaps is forbidden`

**解决**:
```bash
# 检查 Role
kubectl get role configmap-watcher -o yaml

# 检查 RoleBinding
kubectl get rolebinding configmap-watcher-binding -o yaml

# 重新应用 RBAC
kubectl apply -f examples/rbac.yaml
```

## 常见问题

### Q: 如何查看帮助信息？

```bash
./bin/k8s-sidecar --help
```

## 常用命令速查

```bash
# 构建
make build

# 测试
make test

# 清理
make clean

# 格式化代码
make fmt

# 代码检查
make vet

# 构建 Docker 镜像
make docker-build VERSION=v1.0.0

# 查看帮助
./bin/k8s-sidecar --help
```

## 下一步

- 📖 阅读 [DESIGN.md](DESIGN.md) 了解详细设计
- 📋 查看 [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) 了解项目结构
- 🚀 在生产环境部署前进行充分测试
- 🔧 根据实际需求调整配置和参数

## 获取帮助

如有问题，请查看：
- README.md - 完整文档
- DESIGN.md - 设计文档
- 示例配置 - examples/ 目录

祝使用愉快！🎉
