package config

import (
	"encoding/json"
	"fmt"
	"os"
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

	// 预处理环境变量中的特殊格式
	preprocessEnvVars(v)

	// 设置默认值
	setDefaults(v)

	// 反序列化配置到结构体
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
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

// preprocessEnvVars 在 Unmarshal 之前预处理环境变量中的特殊格式
//
// 由于 Viper 的 Unmarshal 无法自动处理某些类型转换（如 JSON 字符串到 map），
// 需要在反序列化之前手动解析并设置到 Viper 实例中。
//
// 支持的转换：
// - LABEL_SELECTOR: JSON 字符串或 key=value 格式 → map[string]string
// - NAMESPACES: 逗号分隔字符串 → []string
func preprocessEnvVars(v *viper.Viper) {
	// 处理 LABEL_SELECTOR 环境变量（JSON 字符串或 key=value 格式 → map）
	labelSelectorEnv := []string{"LABEL_SELECTOR", "SIDECAR_LABEL_SELECTOR"}
	for _, envVar := range labelSelectorEnv {
		if val := os.Getenv(envVar); val != "" {
			var selector map[string]string
			
			// 尝试解析 JSON 格式
			if val[0] == '{' {
				if err := json.Unmarshal([]byte(val), &selector); err != nil {
					fmt.Printf("Warning: Failed to parse %s as JSON: %v. Expected format: '{\"key\":\"value\"}'\n", envVar, err)
					// 设置为空 map，避免 Unmarshal 失败
					v.Set("labelSelector", make(map[string]string))
					return
				}
			} else {
				// 解析 key=value 格式 (如: app=grafana,type=dashboard)
				selector = parseKeyValueFormat(val)
			}
			
			v.Set("labelSelector", selector)
			break
		}
	}

	// 处理 NAMESPACES 环境变量（逗号分隔字符串 → 数组）
	namespacesEnv := []string{"NAMESPACES", "SIDECAR_NAMESPACES"}
	for _, envVar := range namespacesEnv {
		if val := os.Getenv(envVar); val != "" {
			nsList := strings.Split(val, ",")
			for i := range nsList {
				nsList[i] = strings.TrimSpace(nsList[i])
			}
			v.Set("namespaces", nsList)
			break
		}
	}
}

// parseKeyValueFormat 解析 key=value 格式的字符串为 map
// 例如: "app=grafana,type=dashboard" -> {"app": "grafana", "type": "dashboard"}
func parseKeyValueFormat(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
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
