package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Init 初始化日志
func Init(level, format string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext 带上下文的日志
func WithContext(ctx context.Context) *slog.Logger {
	// 可以从 context 中提取 trace_id, user_id 等
	return defaultLogger
}

// Debug 调试日志
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info 信息日志
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn 警告日志
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error 错误日志
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}
