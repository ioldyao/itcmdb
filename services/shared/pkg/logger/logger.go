package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// Init 初始化日志
func Init(level string) error {
	var err error
	once.Do(func() {
		var zapLevel zapcore.Level
		switch level {
		case "debug":
			zapLevel = zapcore.DebugLevel
		case "info":
			zapLevel = zapcore.InfoLevel
		case "warn":
			zapLevel = zapcore.WarnLevel
		case "error":
			zapLevel = zapcore.ErrorLevel
		default:
			zapLevel = zapcore.InfoLevel
		}

		config := zap.Config{
			Level:            zap.NewAtomicLevelAt(zapLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		logger, err = config.Build()
	})
	return err
}

// Get 获取logger实例
func Get() *zap.Logger {
	if logger == nil {
		Init("info")
	}
	return logger
}

// Debug 日志
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info 日志
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn 日志
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error 日志
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal 日志
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}
