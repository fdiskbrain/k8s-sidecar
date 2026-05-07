package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	// KubeConfig kubeconfig 文件路径（可选，为空时使用 In Cluster 模式）
	KubeConfig string `json:"kubeconfig,omitempty" yaml:"kubeconfig,omitempty"`

	// Namespaces 要监控的命名空间列表
	// 支持特殊值 "*" 表示所有命名空间
	Namespaces []string `json:"namespaces" yaml:"namespaces"`

	// LabelSelector label 选择器
	LabelSelector map[string]string `json:"labelSelector" yaml:"labelSelector"`

	// OutputDir 配置文件输出目录
	OutputDir string `json:"outputDir" yaml:"outputDir" default:"/etc/config"`

	// ResyncPeriod Informer 重新同步周期（默认 10 分钟）
	ResyncPeriod time.Duration `json:"resyncPeriod" yaml:"resyncPeriod" default:"10m"`

	// LogLevel 日志级别
	LogLevel string `json:"logLevel" yaml:"logLevel" default:"info"`
}

// LoadConfig 从多个来源加载配置，优先级：环境变量 > 配置文件 > 命令行参数 > 默认值
func LoadConfig(configFile string) (*Config, error) {
	cfg := &Config{}

	// 1. 应用默认值
	applyDefaults(cfg)

	// 2. 从配置文件加载（如果存在）
	if configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			fmt.Printf("Warning: Failed to load config file %s: %v\n", configFile, err)
		}
	}

	// 3. 从环境变量覆盖
	loadFromEnv(cfg)

	// 4. 验证配置
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// applyDefaults 应用默认值
func applyDefaults(cfg *Config) {
	if cfg.OutputDir == "" {
		cfg.OutputDir = "/etc/config"
	}
	if cfg.ResyncPeriod == 0 {
		cfg.ResyncPeriod = 10 * time.Minute
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if len(cfg.Namespaces) == 0 {
		cfg.Namespaces = []string{"default"}
	}
}

// loadFromFile 从 YAML 配置文件加载
func loadFromFile(cfg *Config, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	// KUBECONFIG
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		cfg.KubeConfig = kubeconfig
	}

	// NAMESPACES
	if namespaces := os.Getenv("NAMESPACES"); namespaces != "" {
		cfg.Namespaces = strings.Split(namespaces, ",")
		for i := range cfg.Namespaces {
			cfg.Namespaces[i] = strings.TrimSpace(cfg.Namespaces[i])
		}
	}

	// LABEL_SELECTOR (JSON format)
	if labelSelector := os.Getenv("LABEL_SELECTOR"); labelSelector != "" {
		var selector map[string]string
		if err := json.Unmarshal([]byte(labelSelector), &selector); err == nil {
			cfg.LabelSelector = selector
		}
	}

	// OUTPUT_DIR
	if outputDir := os.Getenv("OUTPUT_DIR"); outputDir != "" {
		cfg.OutputDir = outputDir
	}

	// RESYNC_PERIOD
	if resyncPeriod := os.Getenv("RESYNC_PERIOD"); resyncPeriod != "" {
		if period, err := time.ParseDuration(resyncPeriod); err == nil {
			cfg.ResyncPeriod = period
		}
	}

	// LOG_LEVEL
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 验证 LabelSelector
	if len(cfg.LabelSelector) == 0 {
		return fmt.Errorf("labelSelector is required")
	}

	// 验证 Namespaces
	if len(cfg.Namespaces) == 0 {
		return fmt.Errorf("at least one namespace must be specified")
	}

	// 验证 OutputDir
	if cfg.OutputDir == "" {
		return fmt.Errorf("outputDir is required")
	}

	// 验证 ResyncPeriod
	if cfg.ResyncPeriod <= 0 {
		return fmt.Errorf("resyncPeriod must be positive")
	}

	// 验证 LogLevel
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	return nil
}

// BuildLabelSelectorString 将 LabelSelector map 转换为 Kubernetes 标签选择器字符串
func (cfg *Config) BuildLabelSelectorString() string {
	var selectors []string
	for k, v := range cfg.LabelSelector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(selectors, ",")
}

// IsAllNamespaces 检查是否监控所有命名空间
func (cfg *Config) IsAllNamespaces() bool {
	return len(cfg.Namespaces) == 1 && cfg.Namespaces[0] == "*"
}
