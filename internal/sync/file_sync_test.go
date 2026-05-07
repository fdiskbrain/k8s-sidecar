package sync

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"go.uber.org/zap"
)

func TestFileSyncService_SyncConfigMap(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建 logger
	logger, _ := zap.NewDevelopment()

	// 创建文件同步服务
	fss := NewFileSyncService(tmpDir, logger)

	// 创建测试用的 ConfigMap
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"config.txt": "Hello, World!",
			"app.conf":   "[server]\nport = 8080",
		},
	}

	// 同步 ConfigMap
	err := fss.SyncConfigMap(cm)
	if err != nil {
		t.Fatalf("Failed to sync ConfigMap: %v", err)
	}

	// 验证文件是否创建（现在文件直接在 tmpDir 根目录）
	expectedDir := tmpDir
	
	configFile := filepath.Join(expectedDir, "config.txt")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", string(content))
	}

	appConfFile := filepath.Join(expectedDir, "app.conf")
	content, err = os.ReadFile(appConfFile)
	if err != nil {
		t.Fatalf("Failed to read app.conf file: %v", err)
	}
	expectedContent := "[server]\nport = 8080"
	if string(content) != expectedContent {
		t.Errorf("Expected '%s', got '%s'", expectedContent, string(content))
	}
}

func TestFileSyncService_DeleteConfigMap(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建 logger
	logger, _ := zap.NewDevelopment()

	// 创建文件同步服务
	fss := NewFileSyncService(tmpDir, logger)

	// 先创建一个 ConfigMap
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"config.txt": "test",
		},
	}

	err := fss.SyncConfigMap(cm)
	if err != nil {
		t.Fatalf("Failed to sync ConfigMap: %v", err)
	}

	// 验证文件存在（现在文件在 tmpDir 根目录）
	expectedFile := filepath.Join(tmpDir, "config.txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatal("Expected file to exist")
	}

	// 删除 ConfigMap（注意：现在不会删除文件，只会记录日志）
	err = fss.DeleteConfigMap("default", "test-config")
	if err != nil {
		t.Fatalf("Failed to delete ConfigMap: %v", err)
	}

	// 验证文件仍然存在（因为 DeleteConfigMap 不再删除文件）
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Error("File should still exist - DeleteConfigMap no longer removes files in flat structure")
	}
}

func TestFileSyncService_AtomicWrite(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建 logger
	logger, _ := zap.NewDevelopment()

	// 创建文件同步服务
	fss := NewFileSyncService(tmpDir, logger)

	// 原子写入文件
	testFile := filepath.Join(tmpDir, "test.txt")
	err := fss.atomicWriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to atomically write file: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(content))
	}

	// 验证没有临时文件残留
	files, _ := os.ReadDir(tmpDir)
	for _, f := range files {
		if len(f.Name()) >= 4 && f.Name()[:4] == ".tmp" {
			t.Errorf("Found leftover temp file: %s", f.Name())
		}
	}
}
