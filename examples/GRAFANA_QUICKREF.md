# Grafana Dashboard 自动加载 - 快速参考

## 一键部署命令

```bash
# 1. 应用 RBAC 权限
kubectl apply -f examples/rbac-grafana.yaml

# 2. 创建示例 Dashboard
kubectl apply -f examples/grafana-dashboard.yaml

# 3. 部署 Grafana with Sidecar
kubectl apply -f examples/deployment-grafana.yaml

# 4. 等待 Pod 就绪
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=120s

# 5. 访问 Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

浏览器访问: http://localhost:3000  
用户名: `admin`  
密码: `admin123`

## 添加新的 Dashboard

### 方法 1: 从 Grafana UI 导出

1. 在 Grafana 中创建/编辑 Dashboard
2. 点击 ⚙️ → JSON Model
3. 复制 JSON 内容
4. 创建 ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-custom-dashboard
  namespace: monitoring
  labels:
    app: grafana
    type: dashboard
data:
  my-dashboard.json: |
    {
      "title": "My Custom Dashboard",
      "uid": "my-custom",
      ...
    }
```

5. 应用配置:

```bash
kubectl apply -f my-dashboard.yaml
```

Sidecar 会自动检测并同步到 Grafana（最多 5 分钟）。

### 方法 2: 使用 kubectl 直接创建

```bash
kubectl create configmap my-dashboard \
  --from-file=dashboard.json=./my-dashboard.json \
  -n monitoring \
  --dry-run=client -o yaml | \
  kubectl label --local -f - app=grafana type=dashboard -o yaml | \
  kubectl apply -f -
```

## 常用操作

### 查看已同步的 Dashboard

```bash
# 列出所有 Dashboard ConfigMap
kubectl get configmaps -n monitoring -l app=grafana,type=dashboard

# 查看 sidecar 同步日志
kubectl logs -n monitoring -l app=grafana -c dashboard-sidecar -f

# 检查文件是否已同步
kubectl exec -n monitoring $(kubectl get pod -n monitoring -l app=grafana -o jsonpath='{.items[0].metadata.name}') \
  -c dashboard-sidecar -- ls -la /var/lib/grafana/dashboards/
```

### 更新 Dashboard

```bash
# 修改 ConfigMap
kubectl edit configmap grafana-dashboard-app-metrics -n monitoring

# Sidecar 会自动检测变化并更新文件
# 等待 5 分钟或手动触发重新同步
```

### 删除 Dashboard

```bash
kubectl delete configmap my-dashboard -n monitoring
```

Grafana 会在下次刷新时自动移除该 Dashboard。

### 强制立即同步

```bash
# 重启 sidecar 容器
kubectl rollout restart deployment/grafana-with-sidecar -n monitoring
```

## 配置参数速查

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--namespaces` | `default` | 监控的命名空间列表 |
| `--label-selector` | 必填 | Label 选择器 |
| `--output-dir` | `/etc/config` | 输出目录 |
| `--resync-period` | `10m` | 重新同步周期 |
| `--log-level` | `info` | 日志级别 |

### Grafana 特定配置

```yaml
# 推荐配置
namespaces: [monitoring]
labelSelector:
  app: grafana
  type: dashboard
outputDir: /var/lib/grafana/dashboards
resyncPeriod: 5m
logLevel: info
```

## 故障排查速查

### Sidecar 无法连接 API Server

```bash
# 检查权限
kubectl auth can-i list configmaps \
  --as=system:serviceaccount:monitoring:grafana-sidecar \
  -n monitoring

# 查看日志
kubectl logs -n monitoring -l app=grafana -c dashboard-sidecar --tail=50
```

### Dashboard 未显示

```bash
# 1. 验证 ConfigMap 标签
kubectl get configmaps -n monitoring -l app=grafana,type=dashboard

# 2. 检查文件是否存在
kubectl exec -n monitoring <pod-name> -c dashboard-sidecar -- \
  ls /var/lib/grafana/dashboards/

# 3. 验证 JSON 格式
kubectl get configmap <name> -n monitoring \
  -o jsonpath='{.data.<file>}' | jq .

# 4. 检查 Grafana 日志
kubectl logs -n monitoring -l app=grafana -c grafana | grep -i error
```

### 权限问题

```bash
# 重新应用 RBAC
kubectl apply -f examples/rbac-grafana.yaml

# 验证 ServiceAccount
kubectl get serviceaccount grafana-sidecar -n monitoring
```

## 环境变量配置

也可以在 Deployment 中使用环境变量配置 sidecar：

```yaml
env:
- name: NAMESPACES
  value: "monitoring,production"
- name: LABEL_SELECTOR
  value: '{"app":"grafana","type":"dashboard"}'
- name: OUTPUT_DIR
  value: "/var/lib/grafana/dashboards"
- name: RESYNC_PERIOD
  value: "5m"
- name: LOG_LEVEL
  value: "info"
```

## 多环境部署

### 开发环境

```yaml
args:
- --namespaces=dev
- --label-selector=app=grafana,type=dashboard
- --output-dir=/var/lib/grafana/dashboards
- --resync-period=1m  # 更快响应
- --log-level=debug   # 详细日志
```

### 生产环境

```yaml
args:
- --namespaces=prod
- --label-selector=app=grafana,type=dashboard,env=prod
- --output-dir=/var/lib/grafana/dashboards
- --resync-period=10m  # 减少 API 调用
- --log-level=warn     # 减少日志噪音
```

## 完整工作流程

```
1. 创建 Dashboard JSON
         ↓
2. 创建 ConfigMap (带 label)
         ↓
3. kubectl apply
         ↓
4. Sidecar 检测到变化
         ↓
5. 同步到文件系统
         ↓
6. Grafana 自动加载
         ↓
7. Dashboard 可用 ✓
```

## 相关文档

- 📖 [完整集成指南](GRAFANA_INTEGRATION.md)
- 📖 [项目主文档](../README.md)
- 📖 [设计文档](../DESIGN.md)

## 示例文件清单

| 文件 | 用途 |
|------|------|
| `grafana-dashboard.yaml` | Dashboard ConfigMap 示例 |
| `deployment-grafana.yaml` | Grafana + Sidecar 部署 |
| `rbac-grafana.yaml` | RBAC 权限配置 |
| `config-grafana.yaml.example` | Sidecar 配置模板 |
| `GRAFANA_INTEGRATION.md` | 详细集成文档 |
