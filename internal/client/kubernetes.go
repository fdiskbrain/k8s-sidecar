package client

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// KubernetesClient Kubernetes 客户端封装
type KubernetesClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewKubernetesClient 创建 Kubernetes 客户端
func NewKubernetesClient(kubeconfig string) (*KubernetesClient, error) {
	config, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	// 设置 QPS 和 Burst 限制，减少 API Server 压力
	config.QPS = 5.0
	config.Burst = 10

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &KubernetesClient{
		clientset: clientset,
		config:    config,
	}, nil
}

// GetClientSet 获取 Kubernetes ClientSet
func (kc *KubernetesClient) GetClientSet() *kubernetes.Clientset {
	return kc.clientset
}

// GetConfig 获取 REST Config
func (kc *KubernetesClient) GetConfig() *rest.Config {
	return kc.config
}

// buildConfig 构建 Kubernetes 配置
func buildConfig(kubeconfig string) (*rest.Config, error) {
	// 如果提供了 kubeconfig 路径，使用它
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// 尝试从环境变量获取 KUBECONFIG
	if envKubeconfig := clientcmd.RecommendedHomeFile; envKubeconfig != "" {
		if _, err := rest.InClusterConfig(); err != nil {
			// 不在集群内，尝试使用本地 kubeconfig
			if home := homedir.HomeDir(); home != "" {
				kubeconfigPath := filepath.Join(home, ".kube", "config")
				return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			}
		}
	}

	// 使用 In Cluster 配置
	return rest.InClusterConfig()
}

// VerifyConnection 验证与 API Server 的连接
func (kc *KubernetesClient) VerifyConnection() error {
	_, err := kc.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes API server: %w", err)
	}
	return nil
}
