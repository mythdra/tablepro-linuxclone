package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

var logger *slog.Logger

// init initializes the logger with JSON handler and file rotation
func init() {
	// Create log directory
	logDir := filepath.Join(os.Getenv("HOME"), ".config", "tablepro", "logs")
	os.MkdirAll(logDir, 0755)

	logFile := filepath.Join(logDir, "app.log")

	// Set up lumberjack for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	// Create handler with JSON format
	handler := slog.NewJSONHandler(lumberjackLogger, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger = slog.New(handler)

	// Also log to stdout for development
	slog.SetDefault(logger)
}

// SetLogLevel changes the logging level
func SetLogLevel(level LogLevel) {
	logger.Info("Log level set", "level", level)
	// Note: Changing level at runtime requires handler recreation
	// For now, level is set in init() and can be changed via rebuild
}

// With returns a logger with context fields
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

// DebugContext logs a debug message with context
func DebugContext(ctx context.Context, msg string, args ...any) {
	logger.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context
func InfoContext(ctx context.Context, msg string, args ...any) {
	logger.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context
func WarnContext(ctx context.Context, msg string, args ...any) {
	logger.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context
func ErrorContext(ctx context.Context, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, args...)
}

// DebugEmit logs a debug message and emits it as an event to frontend
func DebugEmit(ctx context.Context, event string, data any) {
	logger.DebugContext(ctx, "Event emitted", "event", event, "data", data)
	runtime.EventsEmit(ctx, "debug:"+event, data)
}
