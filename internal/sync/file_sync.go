package sync

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"k8s.io/api/core/v1"
)

// FileSyncService 文件同步服务
type FileSyncService struct {
	outputDir string
	logger    *zap.Logger
}

// NewFileSyncService 创建文件同步服务
func NewFileSyncService(outputDir string, logger *zap.Logger) *FileSyncService {
	return &FileSyncService{
		outputDir: outputDir,
		logger:    logger,
	}
}

// SyncConfigMap 将 ConfigMap 同步到文件系统
func (fss *FileSyncService) SyncConfigMap(cm *v1.ConfigMap) error {
	startTime := time.Now()

	// 创建 ConfigMap 专属目录: {outputDir}/
	configMapDir := fss.outputDir

	// 确保目录存在
	if err := os.MkdirAll(configMapDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configMapDir, err)
	}

	filesUpdated := 0
	var errors []error

	// 遍历 ConfigMap 的每个 key，将其写入文件
	for key, content := range cm.Data {
		filePath := filepath.Join(configMapDir, key)

		// 原子写入文件
		if err := fss.atomicWriteFile(filePath, []byte(content), 0644); err != nil {
			errors = append(errors, fmt.Errorf("failed to write file %s: %w", filePath, err))
			continue
		}
		filesUpdated++
	}

	// 同时处理 BinaryData
	for key, content := range cm.BinaryData {
		filePath := filepath.Join(configMapDir, key)

		if err := fss.atomicWriteFile(filePath, content, 0644); err != nil {
			errors = append(errors, fmt.Errorf("failed to write binary file %s: %w", filePath, err))
			continue
		}
		filesUpdated++
	}

	duration := time.Since(startTime)

	// 记录日志
	if len(errors) > 0 {
		fss.logger.Warn("ConfigMap synced with errors",
			zap.String("namespace", cm.Namespace),
			zap.String("configmap", cm.Name),
			zap.Int("files_updated", filesUpdated),
			zap.Int("errors", len(errors)),
			zap.Duration("duration", duration),
			zap.Errors("errors", errors),
		)
		return fmt.Errorf("synced with %d errors", len(errors))
	}

	fss.logger.Info("ConfigMap synced successfully",
		zap.String("namespace", cm.Namespace),
		zap.String("configmap", cm.Name),
		zap.Int("files_updated", filesUpdated),
		zap.Duration("duration", duration),
	)

	return nil
}

// DeleteConfigMap 删除 ConfigMap 对应的文件目录
func (fss *FileSyncService) DeleteConfigMap(namespace, name string) error {
	startTime := time.Now()

	// 注意：由于所有文件都在 outputDir 根目录，这里不再删除目录
	// 如果需要清理特定 ConfigMap 的文件，需要额外的元数据追踪
	fss.logger.Debug("Skipping directory deletion - all files are in output root",
		zap.String("namespace", namespace),
		zap.String("configmap", name),
	)

	duration := time.Since(startTime)

	fss.logger.Debug("ConfigMap file cleanup skipped",
		zap.String("namespace", namespace),
		zap.String("configmap", name),
		zap.Duration("duration", duration),
	)

	return nil
}

// atomicWriteFile 原子写入文件（先写临时文件，再重命名）
func (fss *FileSyncService) atomicWriteFile(filePath string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)

	// 创建临时文件
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	// 确保清理临时文件
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpName)
		}
	}()

	// 写入内容
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// 关闭文件
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil

	// 设置权限
	if err := os.Chmod(tmpName, perm); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// 原子重命名
	if err := os.Rename(tmpName, filePath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// CompareAndSync 比较并同步 ConfigMap，只在内容有变化时更新
func (fss *FileSyncService) CompareAndSync(cm *v1.ConfigMap) (bool, error) {
	configMapDir := filepath.Join(fss.outputDir, cm.Namespace, cm.Name)

	// 检查是否有任何变化
	hasChanges := false

	// 检查 Data 中的每个 key
	for key, content := range cm.Data {
		filePath := filepath.Join(configMapDir, key)

		// 读取现有文件内容
		if existingContent, err := os.ReadFile(filePath); err == nil {
			// 比较内容
			if !bytes.Equal(existingContent, []byte(content)) {
				hasChanges = true
				break
			}
		} else {
			// 文件不存在或读取失败，需要更新
			hasChanges = true
			break
		}
	}

	// 如果没有变化，跳过同步
	if !hasChanges {
		fss.logger.Debug("No changes detected, skipping sync",
			zap.String("namespace", cm.Namespace),
			zap.String("configmap", cm.Name),
		)
		return false, nil
	}

	// 有变化，执行同步
	err := fss.SyncConfigMap(cm)
	return true, err
}
