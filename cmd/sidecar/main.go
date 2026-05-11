package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	flag.StringVar(&outputDir, "output-dir", "", "Directory to write config files")
	flag.StringVar(&resyncPeriod, "resync-period", "", "Informer resync period (e.g., 10m, 5s)")
	flag.StringVar(&logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flag.StringVar(&configFile, "config", "/etc/sidecar/config.yaml", "Path to config file")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.Parse()

	if version {
		fmt.Printf("k8s-sidecar version: %s\n", Version)
		os.Exit(0)
	}

	// 如果提供了命令行参数，设置对应的环境变量以便 Viper 读取
	setupEnvOverrides(kubeconfig, namespaces, labelSelector, outputDir, resyncPeriod, logLevel)

	// 加载配置（Viper 会自动合并配置文件、环境变量和默认值）
	cfg, err := config.LoadConfig(configFile)
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

	log.Info("Starting k8s-sidecar",
		zap.String("version", Version),
		zap.Strings("namespaces", cfg.Namespaces),
		zap.String("labelSelector", cfg.BuildLabelSelectorString()),
		zap.String("outputDir", cfg.OutputDir),
		zap.Duration("resyncPeriod", cfg.ResyncPeriod),
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

	log.Info("k8s-sidecar stopped")
}

// setupEnvOverrides 将命令行参数转换为环境变量，让 Viper 能够读取
// 命令行参数的优先级高于配置文件和环境变量
func setupEnvOverrides(kubeconfig, namespaces, labelSelector, outputDir, resyncPeriod, logLevel string) {
	if kubeconfig != "" {
		_ = os.Setenv("SIDECAR_KUBECONFIG", kubeconfig)
	}

	if namespaces != "" {
		_ = os.Setenv("SIDECAR_NAMESPACES", namespaces)
	}

	if labelSelector != "" {
		// 直接设置原始字符串，由 LoadConfig 统一解析
		// 支持两种格式：
		// 1. JSON 格式: {"app":"grafana","type":"dashboard"}
		// 2. key=value 格式: app=grafana,type=dashboard
		_ = os.Setenv("SIDECAR_LABEL_SELECTOR", labelSelector)
	}

	if outputDir != "" {
		_ = os.Setenv("SIDECAR_OUTPUT_DIR", outputDir)
	}

	if resyncPeriod != "" {
		// 验证 duration 格式
		if _, err := time.ParseDuration(resyncPeriod); err == nil {
			_ = os.Setenv("SIDECAR_RESYNC_PERIOD", resyncPeriod)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Invalid resync period '%s': %v\n", resyncPeriod, err)
		}
	}

	if logLevel != "" {
		_ = os.Setenv("SIDECAR_LOG_LEVEL", logLevel)
	}
}

// parseLabelSelector 解析 label selector 字符串为 map
func parseLabelSelector(labelSelector string) map[string]string {
	selectorMap := make(map[string]string)
	pairs := strings.Split(labelSelector, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			selectorMap[parts[0]] = parts[1]
		}
	}
	return selectorMap
}

// labelSelectorToJSON 将 label selector map 转换为 JSON 字符串
func labelSelectorToJSON(selector map[string]string) string {
	if len(selector) == 0 {
		return "{}"
	}

	result := "{"
	first := true
	for k, v := range selector {
		if !first {
			result += ","
		}
		result += fmt.Sprintf(`"%s":"%s"`, k, v)
		first = false
	}
	result += "}"
	return result
}
