package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	// KubeConfig kubeconfig 文件路径（可选，为空时使用 In Cluster 模式）
	KubeConfig string `mapstructure:"kubeconfig" json:"kubeconfig,omitempty" yaml:"kubeconfig,omitempty"`

	// Namespaces 要监控的命名空间列表
	// 支持特殊值 "*" 表示所有命名空间
	Namespaces []string `mapstructure:"namespaces" json:"namespaces" yaml:"namespaces"`

	// LabelSelector label 选择器
	LabelSelector map[string]string `mapstructure:"labelSelector" json:"labelSelector" yaml:"labelSelector"`

	// OutputDir 配置文件输出目录
	OutputDir string `mapstructure:"outputDir" json:"outputDir" yaml:"outputDir"`

	// ResyncPeriod Informer 重新同步周期（默认 10 分钟）
	ResyncPeriod time.Duration `mapstructure:"resyncPeriod" json:"resyncPeriod" yaml:"resyncPeriod"`

	// LogLevel 日志级别
	LogLevel string `mapstructure:"logLevel" json:"logLevel" yaml:"logLevel"`
}

// LoadConfig 从多个来源加载配置，优先级：命令行参数 > 环境变量 > 配置文件 > 默认值
func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()

	// 设置配置名称和类型
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// 默认在以下位置查找配置文件
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/k8s-sidecar")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.k8s-sidecar")
	}

	// 启用环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置环境变量前缀（可选）
	v.SetEnvPrefix("SIDECAR")

	// 读取配置文件（如果存在）
	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在不是错误，继续使用其他配置源
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 绑定环境变量
	bindEnvVars(v)

	// 设置默认值
	setDefaults(v)

	// 反序列化配置到结构体
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 处理特殊的 Namespaces 格式（支持逗号分隔的字符串）
	if len(cfg.Namespaces) == 0 {
		if ns := v.GetString("namespaces"); ns != "" {
			cfg.Namespaces = strings.Split(ns, ",")
			for i := range cfg.Namespaces {
				cfg.Namespaces[i] = strings.TrimSpace(cfg.Namespaces[i])
			}
		}
	}

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// bindEnvVars 绑定环境变量到配置项
func bindEnvVars(v *viper.Viper) {
	// 显式绑定关键环境变量
	_ = v.BindEnv("kubeconfig", "KUBECONFIG", "SIDECAR_KUBECONFIG")
	_ = v.BindEnv("namespaces", "NAMESPACES", "SIDECAR_NAMESPACES")
	_ = v.BindEnv("labelSelector", "LABEL_SELECTOR", "SIDECAR_LABEL_SELECTOR")
	_ = v.BindEnv("outputDir", "OUTPUT_DIR", "SIDECAR_OUTPUT_DIR")
	_ = v.BindEnv("resyncPeriod", "RESYNC_PERIOD", "SIDECAR_RESYNC_PERIOD")
	_ = v.BindEnv("logLevel", "LOG_LEVEL", "SIDECAR_LOG_LEVEL")
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("outputDir", "/etc/config")
	v.SetDefault("resyncPeriod", "10m")
	v.SetDefault("logLevel", "info")
	v.SetDefault("namespaces", []string{"default"})
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

// String 返回配置的字符串表示（用于调试）
func (cfg *Config) String() string {
	return fmt.Sprintf(
		"Config{KubeConfig: %s, Namespaces: %v, LabelSelector: %v, OutputDir: %s, ResyncPeriod: %v, LogLevel: %s}",
		cfg.KubeConfig,
		cfg.Namespaces,
		cfg.LabelSelector,
		cfg.OutputDir,
		cfg.ResyncPeriod,
		cfg.LogLevel,
	)
}
