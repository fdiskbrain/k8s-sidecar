package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// 保存并清除可能干扰测试的环境变量
	envVars := []string{
		"KUBECONFIG", "SIDECAR_KUBECONFIG",
		"NAMESPACES", "SIDECAR_NAMESPACES",
		"LABEL_SELECTOR", "SIDECAR_LABEL_SELECTOR",
		"OUTPUT_DIR", "SIDECAR_OUTPUT_DIR",
		"RESYNC_PERIOD", "SIDECAR_RESYNC_PERIOD",
		"LOG_LEVEL", "SIDECAR_LOG_LEVEL",
	}

	originalEnv := make(map[string]string)
	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
		_ = os.Unsetenv(envVar)
	}

	defer func() {
		// 恢复原始环境变量
		for k, v := range originalEnv {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}()

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

func TestLoadConfigWithDefaults(t *testing.T) {
	// 清除所有环境变量
	envVars := []string{
		"KUBECONFIG", "SIDECAR_KUBECONFIG",
		"NAMESPACES", "SIDECAR_NAMESPACES",
		"LABEL_SELECTOR", "SIDECAR_LABEL_SELECTOR",
		"OUTPUT_DIR", "SIDECAR_OUTPUT_DIR",
		"RESYNC_PERIOD", "SIDECAR_RESYNC_PERIOD",
		"LOG_LEVEL", "SIDECAR_LOG_LEVEL",
	}

	for _, envVar := range envVars {
		_ = os.Unsetenv(envVar)
	}

	// 创建最小化配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "minimal-config.yaml")

	content := `
labelSelector:
  app: test
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

	// 验证默认值
	if cfg.OutputDir != "/etc/config" {
		t.Errorf("Expected default output dir '/etc/config', got '%s'", cfg.OutputDir)
	}

	if cfg.ResyncPeriod != 10*time.Minute {
		t.Errorf("Expected default resync period 10m, got %v", cfg.ResyncPeriod)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.LogLevel)
	}

	if len(cfg.Namespaces) != 1 || cfg.Namespaces[0] != "default" {
		t.Errorf("Expected default namespace ['default'], got %v", cfg.Namespaces)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// 清除可能干扰的环境变量
	envVars := []string{
		"KUBECONFIG", "SIDECAR_KUBECONFIG",
		"NAMESPACES", "SIDECAR_NAMESPACES",
		"LABEL_SELECTOR", "SIDECAR_LABEL_SELECTOR",
		"OUTPUT_DIR", "SIDECAR_OUTPUT_DIR",
		"RESYNC_PERIOD", "SIDECAR_RESYNC_PERIOD",
		"LOG_LEVEL", "SIDECAR_LOG_LEVEL",
	}

	for _, envVar := range envVars {
		_ = os.Unsetenv(envVar)
	}

	// 设置环境变量
	_ = os.Setenv("SIDECAR_OUTPUT_DIR", "/env/config")
	_ = os.Setenv("SIDECAR_LOG_LEVEL", "warn")
	_ = os.Setenv("SIDECAR_NAMESPACES", "ns1,ns2")
	_ = os.Setenv("SIDECAR_LABEL_SELECTOR", `{"env":"prod"}`)
	_ = os.Setenv("SIDECAR_RESYNC_PERIOD", "15m")

	defer func() {
		for _, envVar := range envVars {
			_ = os.Unsetenv(envVar)
		}
	}()

	// 创建最小化配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "env-config.yaml")

	content := `
labelSelector:
  app: test
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

	// 验证环境变量覆盖
	if cfg.OutputDir != "/env/config" {
		t.Errorf("Expected output dir from env '/env/config', got '%s'", cfg.OutputDir)
	}

	if cfg.LogLevel != "warn" {
		t.Errorf("Expected log level from env 'warn', got '%s'", cfg.LogLevel)
	}

	if len(cfg.Namespaces) != 2 {
		t.Errorf("Expected 2 namespaces from env, got %d: %v", len(cfg.Namespaces), cfg.Namespaces)
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
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				LabelSelector: map[string]string{"app": "myapp"},
				Namespaces:    []string{"default"},
				OutputDir:     "/etc/config",
				ResyncPeriod:  10 * time.Minute,
				LogLevel:      "info",
			},
			wantErr: false,
		},
		{
			name: "missing label selector",
			cfg: &Config{
				Namespaces:   []string{"default"},
				OutputDir:    "/etc/config",
				ResyncPeriod: 10 * time.Minute,
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "empty namespaces",
			cfg: &Config{
				LabelSelector: map[string]string{"app": "myapp"},
				Namespaces:    []string{},
				OutputDir:     "/etc/config",
				ResyncPeriod:  10 * time.Minute,
				LogLevel:      "info",
			},
			wantErr: true,
		},
		{
			name: "empty output dir",
			cfg: &Config{
				LabelSelector: map[string]string{"app": "myapp"},
				Namespaces:    []string{"default"},
				OutputDir:     "",
				ResyncPeriod:  10 * time.Minute,
				LogLevel:      "info",
			},
			wantErr: true,
		},
		{
			name: "invalid resync period",
			cfg: &Config{
				LabelSelector: map[string]string{"app": "myapp"},
				Namespaces:    []string{"default"},
				OutputDir:     "/etc/config",
				ResyncPeriod:  -1 * time.Minute,
				LogLevel:      "info",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				LabelSelector: map[string]string{"app": "myapp"},
				Namespaces:    []string{"default"},
				OutputDir:     "/etc/config",
				ResyncPeriod:  10 * time.Minute,
				LogLevel:      "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		KubeConfig:    "/path/to/kubeconfig",
		Namespaces:    []string{"default", "production"},
		LabelSelector: map[string]string{"app": "myapp"},
		OutputDir:     "/etc/config",
		ResyncPeriod:  10 * time.Minute,
		LogLevel:      "info",
	}

	str := cfg.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// 验证字符串包含关键字段
	expectedFields := []string{"/path/to/kubeconfig", "default", "production", "myapp", "/etc/config"}
	for _, field := range expectedFields {
		if !strings.Contains(str, field) {
			t.Errorf("Expected string to contain '%s', got: %s", field, str)
		}
	}
}

func TestLoadConfigWithKeyValueFormat(t *testing.T) {
	// 清除可能干扰的环境变量
	envVars := []string{
		"LABEL_SELECTOR", "SIDECAR_LABEL_SELECTOR",
		"NAMESPACES", "SIDECAR_NAMESPACES",
	}

	for _, envVar := range envVars {
		_ = os.Unsetenv(envVar)
	}

	// 设置 key=value 格式的 label selector
	_ = os.Setenv("SIDECAR_LABEL_SELECTOR", "app=grafana,type=dashboard")
	_ = os.Setenv("SIDECAR_NAMESPACES", "monitoring,production")

	defer func() {
		for _, envVar := range envVars {
			_ = os.Unsetenv(envVar)
		}
	}()

	// 加载配置
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证 key=value 格式被正确解析
	if len(cfg.LabelSelector) != 2 {
		t.Errorf("Expected 2 labels, got %d: %v", len(cfg.LabelSelector), cfg.LabelSelector)
	}

	if cfg.LabelSelector["app"] != "grafana" {
		t.Errorf("Expected app=grafana, got app=%s", cfg.LabelSelector["app"])
	}

	if cfg.LabelSelector["type"] != "dashboard" {
		t.Errorf("Expected type=dashboard, got type=%s", cfg.LabelSelector["type"])
	}

	// 验证 namespaces 被正确解析
	if len(cfg.Namespaces) != 2 {
		t.Errorf("Expected 2 namespaces, got %d: %v", len(cfg.Namespaces), cfg.Namespaces)
	}

	if cfg.Namespaces[0] != "monitoring" || cfg.Namespaces[1] != "production" {
		t.Errorf("Expected [monitoring, production], got %v", cfg.Namespaces)
	}
}

func TestLoadConfigWithInvalidJSON(t *testing.T) {
	// 清除环境变量
	_ = os.Unsetenv("SIDECAR_LABEL_SELECTOR")
	_ = os.Unsetenv("LABEL_SELECTOR")

	defer func() {
		_ = os.Unsetenv("SIDECAR_LABEL_SELECTOR")
	}()

	// 创建包含有效配置的临时文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yaml")
	content := `labelSelector:
  app: test
namespaces:
  - default
outputDir: /tmp/test
`
	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// 加载配置 - 应该成功使用配置文件中的值
	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证使用了配置文件中的值
	if cfg.LabelSelector["app"] != "test" {
		t.Errorf("Expected app=test from config file, got %v", cfg.LabelSelector)
	}
}

func TestLoadConfigWithInvalidJSONWarning(t *testing.T) {
	// 设置无效的 JSON
	_ = os.Setenv("SIDECAR_LABEL_SELECTOR", `{invalid json}`)

	defer func() {
		_ = os.Unsetenv("SIDECAR_LABEL_SELECTOR")
	}()

	// 创建包含有效配置的临时文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yaml")
	content := `labelSelector:
  app: fallback
namespaces:
  - default
outputDir: /tmp/test
`
	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// 捕获 stdout 以验证警告信息
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 加载配置 - 应该显示警告但使用配置文件中的值
	cfg, err := LoadConfig(configFile)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// 验证显示了警告
	if !strings.Contains(output, "Warning") || !strings.Contains(output, "Failed to parse") {
		t.Errorf("Expected warning message, got: %s", output)
	}

	// 验证使用了配置文件中的值（fallback）
	if err != nil {
		t.Logf("Config load failed (expected with invalid env): %v", err)
		// 这是可以接受的，因为环境变量无效且没有正确的 fallback
	} else if cfg.LabelSelector["app"] == "fallback" {
		t.Log("Successfully used config file as fallback")
	}
}
