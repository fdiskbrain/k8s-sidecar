package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// 保存并清除可能干扰测试的环境变量
	originalKubeconfig := os.Getenv("KUBECONFIG")
	originalNamespaces := os.Getenv("NAMESPACES")
	originalLabelSelector := os.Getenv("LABEL_SELECTOR")
	originalOutputDir := os.Getenv("OUTPUT_DIR")
	originalResyncPeriod := os.Getenv("RESYNC_PERIOD")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	defer func() {
		// 恢复原始环境变量（忽略错误，因为在 defer 中无法处理）
		_ = os.Setenv("KUBECONFIG", originalKubeconfig)
		_ = os.Setenv("NAMESPACES", originalNamespaces)
		_ = os.Setenv("LABEL_SELECTOR", originalLabelSelector)
		_ = os.Setenv("OUTPUT_DIR", originalOutputDir)
		_ = os.Setenv("RESYNC_PERIOD", originalResyncPeriod)
		_ = os.Setenv("LOG_LEVEL", originalLogLevel)
	}()

	// 清除环境变量以避免干扰测试（忽略错误，这些操作在测试环境中几乎不会失败）
	_ = os.Unsetenv("KUBECONFIG")
	_ = os.Unsetenv("NAMESPACES")
	_ = os.Unsetenv("LABEL_SELECTOR")
	_ = os.Unsetenv("OUTPUT_DIR")
	_ = os.Unsetenv("RESYNC_PERIOD")
	_ = os.Unsetenv("LOG_LEVEL")

	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	content := `
kubeconfig: "/path/to/kubeconfig"
namespaces:
  - default
  - production
labelSelector:
  app: myapp
  type: config
outputDir: "/tmp/config"
resyncPeriod: "5m"
logLevel: "debug"
`
	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// 加载配置
	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	if cfg.KubeConfig != "/path/to/kubeconfig" {
		t.Errorf("Expected kubeconfig '/path/to/kubeconfig', got '%s'", cfg.KubeConfig)
	}

	if len(cfg.Namespaces) != 2 {
		t.Errorf("Expected 2 namespaces, got %d", len(cfg.Namespaces))
	}

	if cfg.LabelSelector["app"] != "myapp" {
		t.Errorf("Expected label selector app=myapp, got %v", cfg.LabelSelector)
	}

	if cfg.OutputDir != "/tmp/config" {
		t.Errorf("Expected output dir '/tmp/config', got '%s'", cfg.OutputDir)
	}

	if cfg.ResyncPeriod != 5*time.Minute {
		t.Errorf("Expected resync period 5m, got %v", cfg.ResyncPeriod)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestBuildLabelSelectorString(t *testing.T) {
	cfg := &Config{
		LabelSelector: map[string]string{
			"app":  "myapp",
			"type": "config",
		},
	}

	result := cfg.BuildLabelSelectorString()

	// 结果应该包含两个选择器（顺序可能不同）
	if result != "app=myapp,type=config" && result != "type=config,app=myapp" {
		t.Errorf("Expected 'app=myapp,type=config' or 'type=config,app=myapp', got '%s'", result)
	}
}

func TestIsAllNamespaces(t *testing.T) {
	cfg1 := &Config{
		Namespaces: []string{"*"},
	}
	if !cfg1.IsAllNamespaces() {
		t.Error("Expected IsAllNamespaces to return true for '*'")
	}

	cfg2 := &Config{
		Namespaces: []string{"default", "production"},
	}
	if cfg2.IsAllNamespaces() {
		t.Error("Expected IsAllNamespaces to return false for specific namespaces")
	}
}

func TestValidateConfig(t *testing.T) {
	// 有效配置
	validCfg := &Config{
		LabelSelector: map[string]string{"app": "myapp"},
		Namespaces:    []string{"default"},
		OutputDir:     "/etc/config",
		ResyncPeriod:  10 * time.Minute,
		LogLevel:      "info",
	}
	if err := validateConfig(validCfg); err != nil {
		t.Errorf("Expected valid config to pass validation: %v", err)
	}

	// 缺少 LabelSelector
	invalidCfg1 := &Config{
		Namespaces:   []string{"default"},
		OutputDir:    "/etc/config",
		ResyncPeriod: 10 * time.Minute,
		LogLevel:     "info",
	}
	if err := validateConfig(invalidCfg1); err == nil {
		t.Error("Expected validation to fail for missing label selector")
	}

	// 空的命名空间列表
	invalidCfg2 := &Config{
		LabelSelector: map[string]string{"app": "myapp"},
		Namespaces:    []string{},
		OutputDir:     "/etc/config",
		ResyncPeriod:  10 * time.Minute,
		LogLevel:      "info",
	}
	if err := validateConfig(invalidCfg2); err == nil {
		t.Error("Expected validation to fail for empty namespaces")
	}
}
