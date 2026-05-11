#!/bin/bash

# Grafana Dashboard 自动加载演示脚本
# 此脚本展示如何使用 k8s-sidecar 实现 Grafana Dashboard 的自动管理

set -e

echo "========================================="
echo "Grafana Dashboard 自动加载演示"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查 kubectl 是否可用
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}错误: kubectl 未安装${NC}"
    exit 1
fi

# 检查 Kubernetes 集群连接
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}错误: 无法连接到 Kubernetes 集群${NC}"
    echo "请确保已配置 kubeconfig 并且集群可访问"
    exit 1
fi

echo -e "${GREEN}✓ Kubernetes 集群连接正常${NC}"
echo ""

# 步骤 1: 创建命名空间
echo -e "${YELLOW}步骤 1: 创建 monitoring 命名空间${NC}"
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✓ 命名空间创建完成${NC}"
echo ""

# 步骤 2: 部署 RBAC
echo -e "${YELLOW}步骤 2: 部署 RBAC 权限${NC}"
kubectl apply -f examples/rbac-grafana.yaml
echo -e "${GREEN}✓ RBAC 配置完成${NC}"
echo ""

# 步骤 3: 创建示例 Dashboard
echo -e "${YELLOW}步骤 3: 创建示例 Dashboard ConfigMap${NC}"
kubectl apply -f examples/grafana-dashboard.yaml
echo -e "${GREEN}✓ Dashboard ConfigMap 创建完成${NC}"
echo ""

# 验证 ConfigMap
echo "已创建的 Dashboard ConfigMap:"
kubectl get configmaps -n monitoring -l app=grafana,type=dashboard
echo ""

# 步骤 4: 部署 Grafana with Sidecar
echo -e "${YELLOW}步骤 4: 部署 Grafana + Sidecar${NC}"
kubectl apply -f examples/deployment-grafana.yaml
echo -e "${GREEN}✓ Grafana 部署完成${NC}"
echo ""

# 等待 Pod 就绪
echo -e "${YELLOW}等待 Pod 就绪...${NC}"
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=120s || {
    echo -e "${RED}Pod 启动超时，请检查日志:${NC}"
    kubectl get pods -n monitoring -l app=grafana
    exit 1
}
echo -e "${GREEN}✓ Pod 已就绪${NC}"
echo ""

# 显示 Pod 状态
echo "Pod 状态:"
kubectl get pods -n monitoring -l app=grafana
echo ""

# 显示 Service 信息
echo "Service 信息:"
kubectl get svc -n monitoring -l app=grafana
echo ""

# 查看 sidecar 日志（最近 10 行）
echo -e "${YELLOW}Sidecar 同步日志（最近 10 行）:${NC}"
kubectl logs -n monitoring -l app=grafana -c dashboard-sidecar --tail=10 || true
echo ""

# 检查同步的文件
echo -e "${YELLOW}已同步的 Dashboard 文件:${NC}"
POD_NAME=$(kubectl get pod -n monitoring -l app=grafana -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n monitoring $POD_NAME -c dashboard-sidecar -- ls -la /var/lib/grafana/dashboards/ || true
echo ""

echo "========================================="
echo -e "${GREEN}部署完成！${NC}"
echo "========================================="
echo ""
echo "访问 Grafana:"
echo "  1. 运行以下命令进行端口转发:"
echo "     kubectl port-forward -n monitoring svc/grafana 3000:3000"
echo ""
echo "  2. 浏览器访问: http://localhost:3000"
echo "     用户名: admin"
echo "     密码: admin123"
echo ""
echo "添加新的 Dashboard:"
echo "  1. 创建带有标签 app=grafana,type=dashboard 的 ConfigMap"
echo "  2. 在 data 字段中包含 dashboard JSON"
echo "  3. 应用 ConfigMap: kubectl apply -f your-dashboard.yaml"
echo "  4. Sidecar 会自动同步到 Grafana（最多 5 分钟）"
echo ""
echo "查看详细文档:"
echo "  - 完整指南: examples/GRAFANA_INTEGRATION.md"
echo "  - 快速参考: examples/GRAFANA_QUICKREF.md"
echo ""
echo "清理资源:"
echo "  kubectl delete -f examples/deployment-grafana.yaml"
echo "  kubectl delete -f examples/grafana-dashboard.yaml"
echo "  kubectl delete -f examples/rbac-grafana.yaml"
echo ""
