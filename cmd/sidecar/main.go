package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"k8s-sidecar/internal/client"
	"k8s-sidecar/internal/config"
	"k8s-sidecar/internal/informer"
	"k8s-sidecar/internal/logger"
	"k8s-sidecar/internal/sync"
)

var (
	Version = "latest"
)

func main() {
	// 解析命令行参数
	var (
		kubeconfig    string
		namespaces    string
		labelSelector string
		outputDir     string
		resyncPeriod  string
		logLevel      string
		configFile    string
		version       bool
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	flag.StringVar(&namespaces, "namespaces", "", "Comma-separated list of namespaces to watch (use '*' for all)")
	flag.StringVar(&labelSelector, "label-selector", "", "Label selector (e.g., 'app=myapp,type=config')")
	flag.StringVar(&outputDir, "output-dir", "/etc/config", "Directory to write config files")
	flag.StringVar(&resyncPeriod, "resync-period", "10m0s", "Informer resync period")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&configFile, "config", "/etc/sidecar/config.yaml", "Path to config file")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.Parse()

	if version {
		fmt.Printf("k8s-configmap-sidecar version: %s\n", Version)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := loadConfig(configFile, kubeconfig, namespaces, labelSelector, outputDir, resyncPeriod, logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.InitLogger(cfg.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log := logger.GetLogger()
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Starting k8s-configmap-sidecar",
		zap.String("version", Version),
		zap.Strings("namespaces", cfg.Namespaces),
		zap.String("labelSelector", cfg.BuildLabelSelectorString()),
		zap.String("outputDir", cfg.OutputDir),
	)

	// 创建 Kubernetes 客户端
	k8sClient, err := client.NewKubernetesClient(cfg.KubeConfig)
	if err != nil {
		log.Fatal("Failed to create Kubernetes client", zap.Error(err))
	}

	// 验证连接
	if err := k8sClient.VerifyConnection(); err != nil {
		log.Fatal("Failed to connect to Kubernetes API server", zap.Error(err))
	}

	log.Info("Connected to Kubernetes API server")

	// 创建文件同步服务
	fileSyncSvc := sync.NewFileSyncService(cfg.OutputDir, log)

	// 创建 Informer 管理器
	informerMgr := informer.NewInformerManager(
		k8sClient.GetClientSet(),
		cfg,
		fileSyncSvc,
		log,
	)

	// 启动 Informer
	if err := informerMgr.Start(); err != nil {
		log.Fatal("Failed to start informers", zap.Error(err))
	}

	log.Info("All components started successfully")

	// 设置信号处理，优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			log.Info("Received signal, shutting down", zap.String("signal", sig.String()))
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	// 等待关闭
	informerMgr.WaitForShutdown(ctx.Done())

	log.Info("k8s-configmap-sidecar stopped")
}

// loadConfig 加载配置，合并配置文件和命令行参数
func loadConfig(configFile, kubeconfig, namespaces, labelSelector, outputDir, resyncPeriod, logLevel string) (*config.Config, error) {
	// 从配置文件加载
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 命令行参数优先级高于配置文件
	if kubeconfig != "" {
		cfg.KubeConfig = kubeconfig
	}

	if namespaces != "" {
		nsList := strings.Split(namespaces, ",")
		for i := range nsList {
			nsList[i] = strings.TrimSpace(nsList[i])
		}
		cfg.Namespaces = nsList
	}

	if labelSelector != "" {
		// 解析 label selector 字符串为 map
		selectorMap := make(map[string]string)
		pairs := strings.Split(labelSelector, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				selectorMap[parts[0]] = parts[1]
			}
		}
		cfg.LabelSelector = selectorMap
	}

	if outputDir != "/etc/config" { // 如果不是默认值
		cfg.OutputDir = outputDir
	}

	if resyncPeriod != "10m0s" { // 如果不是默认值
		// 这里简化处理，实际应该解析 duration 字符串
		// 可以使用 time.ParseDuration
	}

	if logLevel != "info" { // 如果不是默认值
		cfg.LogLevel = logLevel
	}

	return cfg, nil
}
