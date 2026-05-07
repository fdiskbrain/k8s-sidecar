package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// InitLogger 初始化日志系统
func InitLogger(level string) error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	var err error
	log, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// GetLogger 获取日志记录器
func GetLogger() *zap.Logger {
	if log == nil {
		// 如果未初始化，返回默认 logger
		return zap.L()
	}
	return log
}

// WithNamespace 添加命名空间字段到日志
func WithNamespace(namespace string) *zap.Logger {
	return GetLogger().With(zap.String("namespace", namespace))
}

// WithConfigMap 添加 ConfigMap 名称字段到日志
func WithConfigMap(name string) *zap.Logger {
	return GetLogger().With(zap.String("configmap", name))
}

// WithFields 添加多个字段到日志
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}
