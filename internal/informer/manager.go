package informer

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"k8s-sidecar/internal/config"
	"k8s-sidecar/internal/sync"
)

// InformerManager 管理多个命名空间的 Informer
type InformerManager struct {
	clientset    kubernetes.Interface
	cfg          *config.Config
	fileSyncSvc  *sync.FileSyncService
	logger       *zap.Logger
	informers    map[string]cache.SharedIndexInformer // key: namespace
	stopChannels map[string]chan struct{}
}

// NewInformerManager 创建 Informer 管理器
func NewInformerManager(
	clientset kubernetes.Interface,
	cfg *config.Config,
	fileSyncSvc *sync.FileSyncService,
	logger *zap.Logger,
) *InformerManager {
	return &InformerManager{
		clientset:    clientset,
		cfg:          cfg,
		fileSyncSvc:  fileSyncSvc,
		logger:       logger,
		informers:    make(map[string]cache.SharedIndexInformer),
		stopChannels: make(map[string]chan struct{}),
	}
}

// Start 启动所有命名空间的 Informer
func (im *InformerManager) Start() error {
	namespaces := im.cfg.Namespaces

	// 如果是所有命名空间,获取集群中所有命名空间列表
	if im.cfg.IsAllNamespaces() {
		nsList, err := im.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}
		namespaces = make([]string, len(nsList.Items))
		for i, ns := range nsList.Items {
			namespaces[i] = ns.Name
		}
		im.logger.Info("Monitoring all namespaces", zap.Int("count", len(namespaces)))
	}

	// 为每个命名空间启动一个 Informer
	for _, namespace := range namespaces {
		if err := im.startNamespaceInformer(namespace); err != nil {
			im.logger.Error("Failed to start informer for namespace",
				zap.String("namespace", namespace),
				zap.Error(err),
			)
			// 继续启动其他命名空间，不因为一个失败而停止
		}
	}

	im.logger.Info("All informers started", zap.Int("namespaces", len(namespaces)))
	return nil
}

// startNamespaceInformer 启动单个命名空间的 Informer
func (im *InformerManager) startNamespaceInformer(namespace string) error {
	logger := im.logger.With(zap.String("namespace", namespace))

	// 创建带 Label Selector 的 ListWatch
	labelSelector := im.cfg.BuildLabelSelectorString()
	logger.Info("Starting ConfigMap informer with label selector",
		zap.String("labelSelector", labelSelector),
	)

	// 创建 SharedInformerFactory
	factory := informers.NewSharedInformerFactoryWithOptions(
		im.clientset,
		im.cfg.ResyncPeriod,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector
		}),
	)

	// 获取 ConfigMap Informer
	configMapInformer := factory.Core().V1().ConfigMaps().Informer()

	// 注册事件处理器
	configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.handleAdd,
		UpdateFunc: im.handleUpdate,
		DeleteFunc: im.handleDelete,
	})

	// 创建停止通道
	stopCh := make(chan struct{})

	// 启动 Informer
	go func() {
		logger.Info("Starting informer")
		configMapInformer.Run(stopCh)
		logger.Info("Informer stopped")
	}()

	// 等待 Informer 同步
	if !cache.WaitForCacheSync(stopCh, configMapInformer.HasSynced) {
		return fmt.Errorf("timed out waiting for informer cache sync")
	}

	// 保存引用
	im.informers[namespace] = configMapInformer
	im.stopChannels[namespace] = stopCh

	logger.Info("Informer started successfully")
	return nil
}

// handleAdd 处理 ConfigMap 添加事件
func (im *InformerManager) handleAdd(obj interface{}) {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		im.logger.Error("Failed to convert object to ConfigMap")
		return
	}

	logger := im.logger.With(
		zap.String("namespace", cm.Namespace),
		zap.String("configmap", cm.Name),
	)

	logger.Info("ConfigMap added event received")

	// 同步到文件系统
	if err := im.fileSyncSvc.SyncConfigMap(cm); err != nil {
		logger.Error("Failed to sync ConfigMap", zap.Error(err))
	}
}

// handleUpdate 处理 ConfigMap 更新事件
func (im *InformerManager) handleUpdate(oldObj, newObj interface{}) {
	oldCm, ok := oldObj.(*v1.ConfigMap)
	if !ok {
		im.logger.Error("Failed to convert old object to ConfigMap")
		return
	}

	newCm, ok := newObj.(*v1.ConfigMap)
	if !ok {
		im.logger.Error("Failed to convert new object to ConfigMap")
		return
	}

	logger := im.logger.With(
		zap.String("namespace", newCm.Namespace),
		zap.String("configmap", newCm.Name),
	)

	// 检查 ResourceVersion 是否变化
	if oldCm.ResourceVersion == newCm.ResourceVersion {
		logger.Debug("ConfigMap unchanged (same ResourceVersion)")
		return
	}

	logger.Info("ConfigMap updated event received")

	// 同步到文件系统（覆盖写入）
	if err := im.fileSyncSvc.SyncConfigMap(newCm); err != nil {
		logger.Error("Failed to sync ConfigMap", zap.Error(err))
	}
}

// handleDelete 处理 ConfigMap 删除事件
func (im *InformerManager) handleDelete(obj interface{}) {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		// 处理 DeletedFinalStateUnknown
		deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			im.logger.Error("Failed to convert object to ConfigMap or DeletedFinalStateUnknown")
			return
		}
		cm, ok = deletedState.Obj.(*v1.ConfigMap)
		if !ok {
			im.logger.Error("Failed to convert DeletedFinalStateUnknown.Obj to ConfigMap")
			return
		}
	}

	logger := im.logger.With(
		zap.String("namespace", cm.Namespace),
		zap.String("configmap", cm.Name),
	)

	logger.Info("ConfigMap deleted event received")

	// 从文件系统删除
	if err := im.fileSyncSvc.DeleteConfigMap(cm.Namespace, cm.Name); err != nil {
		logger.Error("Failed to delete ConfigMap files", zap.Error(err))
	}
}

// Stop 停止所有 Informer
func (im *InformerManager) Stop() {
	im.logger.Info("Stopping all informers")

	for namespace, stopCh := range im.stopChannels {
		close(stopCh)
		im.logger.Info("Stopped informer", zap.String("namespace", namespace))
	}

	im.logger.Info("All informers stopped")
}

// WaitForShutdown 等待关闭信号
func (im *InformerManager) WaitForShutdown(stopCh <-chan struct{}) {
	<-stopCh
	im.Stop()
}
