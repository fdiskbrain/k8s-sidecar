# Grafana Operator 与 K8s Sidecar 集成指南

本文档介绍如何使用 k8s-sidecar 实现 Grafana Dashboard 的自动加载和管理。

## 架构概述

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                     │
│                                                          │
│  ┌──────────────┐         ┌──────────────────────────┐  │
│  │  ConfigMap   │         │    Grafana Pod           │  │
│  │              │         │                          │  │
│  │ app: grafana │         │  ┌────────────────────┐  │  │
│  │ type:        │ Watch   │  │   Grafana Container│  │  │
│  │ dashboard    │────────▶│  │                    │  │  │
│  │              │         │  │ /var/lib/grafana/  │  │  │
│  └──────────────┘         │  │   dashboards/      │  │  │
│                           │  └─────────▲──────────┘  │  │
│                           │            │ Shared       │  │
│                           │  ┌─────────┴──────────┐  │  │
│                           │  │  Sidecar Container │  │  │
│                           │  │                    │  │  │
│                           │  │ Sync ConfigMaps to │  │  │
│                           │  │   File System      │  │  │
│                           │  └────────────────────┘  │  │
│                           └──────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 快速开始

### 1. 部署 RBAC 权限

```bash
kubectl apply -f examples/rbac-grafana.yaml
```

这将创建：
- ServiceAccount: `grafana-sidecar`
- Role: 允许读取 monitoring 命名空间中的 ConfigMap
- RoleBinding: 绑定权限到 ServiceAccount

### 2. 创建 Dashboard ConfigMap

```bash
kubectl apply -f examples/grafana-dashboard.yaml
```

这会创建示例 Dashboard：
- `app-metrics.json` - 应用监控面板
- `system-resources.json` - 系统资源监控面板

### 3. 部署 Grafana with Sidecar

```bash
kubectl apply -f examples/deployment-grafana.yaml
```

### 4. 验证部署

```bash
# 检查 Pod 状态
kubectl get pods -n monitoring -l app=grafana

# 查看 sidecar 日志
kubectl logs -n monitoring -l app=grafana -c dashboard-sidecar -f

# 查看 Grafana 日志
kubectl logs -n monitoring -l app=grafana -c grafana -f

# 端口转发访问 Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
```

访问 http://localhost:3000，使用 admin/admin123 登录。

## 配置说明

### Label Selector

Sidecar 通过 label selector 识别需要监控的 ConfigMap：

```yaml
labelSelector:
  app: grafana
  type: dashboard
```

只有带有这些标签的 ConfigMap 才会被同步。

### Output Directory

Dashboard JSON 文件会被写入：

```
/var/lib/grafana/dashboards/
```

这是 Grafana 默认的文件系统 provisioning 目录。

### 文件命名规则

ConfigMap 中的每个 key 都会被转换为一个独立的 JSON 文件：

```
ConfigMap: grafana-dashboard-app-metrics
  data:
    app-metrics.json: {...}    → /var/lib/grafana/dashboards/app-metrics.json
    
ConfigMap: grafana-dashboard-system-resources
  data:
    system-resources.json: {...} → /var/lib/grafana/dashboards/system-resources.json
```

## Dashboard ConfigMap 格式

### 基本结构

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-dashboard
  namespace: monitoring
  labels:
    app: grafana        # 必须匹配 label selector
    type: dashboard     # 必须匹配 label selector
data:
  dashboard-name.json: |
    {
      "title": "My Dashboard",
      "uid": "my-dashboard",
      "panels": [...],
      ...
    }
```

### 从现有 Dashboard 导出

1. 在 Grafana UI 中打开 Dashboard
2. 点击齿轮图标（Dashboard Settings）
3. 选择 "JSON Model"
4. 复制 JSON 内容
5. 创建 ConfigMap，将 JSON 放入 data 字段

### 使用 jsonnet 生成 Dashboard

对于复杂的 Dashboard，可以使用 [grafonnet](https://github.com/grafana/grafonnet) 生成：

```jsonnet
local grafana = import 'grafonnet/grafana.libsonnet';

grafana.dashboard.new(
  'Application Metrics',
  uid='app-metrics'
)
.addPanel(
  grafana.graphPanel.new(
    'Request Rate',
    datasource='prometheus',
    targets=[
      grafana.target.new(
        expr='rate(http_requests_total[5m])'
      )
    ]
  )
)
```

## 高级配置

### 多命名空间监控

如果需要从多个命名空间加载 Dashboard：

```yaml
# deployment-grafana.yaml
args:
- --namespaces=monitoring,production,staging
- --label-selector=app=grafana,type=dashboard
- --output-dir=/var/lib/grafana/dashboards
```

**注意**: 需要在每个命名空间中创建相应的 Role 和 RoleBinding，或使用 ClusterRole。

### 自定义 Resync Period

调整同步频率以平衡响应速度和资源消耗：

```yaml
# 快速响应变化（适合开发环境）
resyncPeriod: "1m"

# 标准配置（适合生产环境）
resyncPeriod: "5m"

# 降低 API Server 负载
resyncPeriod: "15m"
```

### 使用配置文件

创建 `config-grafana.yaml`：

```yaml
kubeconfig: ""
namespaces: 
  - monitoring
labelSelector:
  app: grafana
  type: dashboard
outputDir: "/var/lib/grafana/dashboards"
resyncPeriod: "5m"
logLevel: "info"
```

在 Deployment 中挂载：

```yaml
volumeMounts:
- name: sidecar-config
  mountPath: /etc/sidecar/config.yaml
volumes:
- name: sidecar-config
  configMap:
    name: sidecar-config
```

启动时指定配置文件：

```yaml
args:
- --config=/etc/sidecar/config.yaml
```

## Grafana 配置

### 启用文件系统 Provisioning

Grafana 会自动检测 `/var/lib/grafana/dashboards` 目录下的 JSON 文件。

在 `grafana.ini` 中确认配置：

```ini
[dashboards]
# 设置默认首页 Dashboard
default_home_dashboard_path = /var/lib/grafana/dashboards/app-metrics.json

[paths]
# Dashboard provisioning 目录
provisioning = /etc/grafana/provisioning
```

### 使用 Provisioning 配置文件（推荐）

创建 provisioning 配置以更好地控制 Dashboard 加载：

```yaml
# configmap-grafana-provisioning.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-provisioning-dashboards
  namespace: monitoring
data:
  dashboards.yaml: |
    apiVersion: 1
    
    providers:
    - name: 'sidecar-dashboards'
      orgId: 1
      folder: ''
      folderUid: ''
      type: file
      disableDeletion: false
      editable: true
      updateIntervalSeconds: 10
      allowUiUpdates: true
      options:
        path: /var/lib/grafana/dashboards
        foldersFromFilesStructure: false
```

在 Deployment 中挂载：

```yaml
volumeMounts:
- name: provisioning-dashboards
  mountPath: /etc/grafana/provisioning/dashboards
volumes:
- name: provisioning-dashboards
  configMap:
    name: grafana-provisioning-dashboards
```

## 故障排查

### Sidecar 无法连接 API Server

```bash
# 检查 ServiceAccount 是否正确配置
kubectl get serviceaccount grafana-sidecar -n monitoring

# 检查 RBAC 权限
kubectl auth can-i list configmaps --as=system:serviceaccount:monitoring:grafana-sidecar -n monitoring

# 查看 sidecar 日志
kubectl logs -n monitoring -l app=grafana -c dashboard-sidecar --tail=100
```

### Dashboard 未出现在 Grafana 中

```bash
# 1. 检查 ConfigMap 是否有正确的标签
kubectl get configmaps -n monitoring -l app=grafana,type=dashboard

# 2. 检查 sidecar 是否正在同步文件
kubectl exec -n monitoring <grafana-pod> -c dashboard-sidecar -- ls -la /var/lib/grafana/dashboards/

# 3. 检查文件内容
kubectl exec -n monitoring <grafana-pod> -c dashboard-sidecar -- cat /var/lib/grafana/dashboards/app-metrics.json

# 4. 检查 Grafana 日志
kubectl logs -n monitoring -l app=grafana -c grafana | grep -i dashboard

# 5. 验证 provisioning 配置
kubectl exec -n monitoring <grafana-pod> -c grafana -- cat /etc/grafana/provisioning/dashboards/dashboards.yaml
```

### 文件权限问题

确保 sidecar 有权限写入共享卷：

```yaml
securityContext:
  runAsUser: 472  # Grafana 默认用户 ID
  runAsGroup: 472
  fsGroup: 472
```

### Dashboard JSON 格式错误

```bash
# 验证 JSON 格式
kubectl get configmap grafana-dashboard-app-metrics -n monitoring -o jsonpath='{.data.app-metrics\.json}' | jq .

# 检查 sidecar 日志中的错误
kubectl logs -n monitoring -l app=grafana -c dashboard-sidecar | grep -i error
```

## 最佳实践

### 1. Dashboard 版本控制

将 Dashboard JSON 存储在 Git 仓库中，使用 CI/CD 自动部署：

```yaml
# .gitlab-ci.yml 示例
deploy-dashboards:
  stage: deploy
  script:
    - kubectl apply -f dashboards/
  only:
    - main
```

### 2. 使用 Folder 组织 Dashboard

在 Dashboard JSON 中指定 folder：

```json
{
  "meta": {
    "folderId": 1,
    "folderUid": "infrastructure"
  },
  "dashboard": {
    "title": "System Resources",
    ...
  }
}
```

### 3. 添加 Dashboard 标签

便于分类和搜索：

```json
{
  "tags": ["kubernetes", "monitoring", "production"],
  ...
}
```

### 4. 设置合适的刷新间隔

```json
{
  "refresh": "5s",  // 根据数据更新频率调整
  ...
}
```

### 5. 使用 Template Variables

使 Dashboard 更灵活：

```json
{
  "templating": {
    "list": [
      {
        "name": "namespace",
        "type": "query",
        "datasource": "prometheus",
        "query": "label_values(kube_pod_info, namespace)"
      }
    ]
  }
}
```

## 与 Grafana Operator 集成

如果使用 [Grafana Operator](https://github.com/grafana-operator/grafana-operator)，可以结合使用：

### 方案 1: Sidecar + Grafana CR

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
spec:
  config:
    security:
      admin_user: admin
    auth.anonymous:
      enabled: false
  dashboardLabelSelector:
    - matchExpressions:
      - key: app
        operator: In
        values:
        - grafana
```

Sidecar 仍然负责将 ConfigMap 同步到文件系统，Grafana Operator 管理 Grafana 实例的生命周期。

### 方案 2: 完全使用 Grafana Operator

Grafana Operator 也支持直接从 ConfigMap 加载 Dashboard：

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: app-metrics
  labels:
    app: grafana
spec:
  instanceSelector:
    matchLabels:
      dashboards: grafana
  configMapRef:
    name: grafana-dashboard-app-metrics
    key: app-metrics.json
```

**对比**：
- **Sidecar 方案**: 更轻量，不依赖 Operator，适合简单场景
- **Operator 方案**: 功能更强大，支持更多 CRD，适合复杂场景

可以根据需求选择或组合使用两种方案。

## 安全考虑

### 1. 最小权限原则

只授予 sidecar 必要的权限：

```yaml
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]  # 不需要 create/update/delete
```

### 2. 使用 Secret 管理敏感信息

```yaml
env:
- name: GF_SECURITY_ADMIN_PASSWORD
  valueFrom:
    secretKeyRef:
      name: grafana-admin-password
      key: password
```

### 3. 网络策略

限制 Grafana 的网络访问：

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: grafana-network-policy
spec:
  podSelector:
    matchLabels:
      app: grafana
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 3000
```

## 性能优化

### 1. 减少 Resync 频率

对于稳定的生产环境，可以延长 resync period：

```yaml
resyncPeriod: "15m"
```

### 2. 限制监控的命名空间

只监控必要的命名空间：

```yaml
namespaces:
  - monitoring
```

而不是使用 `*` 监控所有命名空间。

### 3. 使用精确的 Label Selector

避免匹配过多的 ConfigMap：

```yaml
labelSelector:
  app: grafana
  type: dashboard
  environment: production  # 额外的过滤条件
```

## 完整示例

查看所有示例文件：

```bash
# 目录结构
examples/
├── grafana-dashboard.yaml          # Dashboard ConfigMap 示例
├── deployment-grafana.yaml         # Grafana + Sidecar 部署
├── rbac-grafana.yaml               # RBAC 权限配置
└── config-grafana.yaml.example     # Sidecar 配置文件模板

# 一键部署
kubectl apply -f examples/rbac-grafana.yaml
kubectl apply -f examples/grafana-dashboard.yaml
kubectl apply -f examples/deployment-grafana.yaml
```

## 参考资料

- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [Grafana File Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/#dashboards)
- [Kubernetes ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Grafana Operator](https://github.com/grafana-operator/grafana-operator)
- [k8s-sidecar 项目文档](../README.md)
