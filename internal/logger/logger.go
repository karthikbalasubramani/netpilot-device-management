package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu        sync.RWMutex
	appLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
)

// Init initializes the global application logger based on the configured log level.
// This should be called once during application startup.
func Init(loglevel string) {
	level := parsedLogLevel(loglevel)

	handler := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	mu.Lock()
	appLogger = handler
	mu.Unlock()

	slog.SetDefault(handler)
}

// Get returns the global application logger instance.
func GetLogger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return appLogger
}

// Info logs an info-level message.
func Info(message string, args ...any) {
	GetLogger().Info(message, args...)
}

// Debug logs an info-level message.
func Debug(message string, args ...any) {
	GetLogger().Debug(message, args...)
}

// Warn logs an info-level message.
func Warn(message string, args ...any) {
	GetLogger().Warn(message, args...)
}

// Error logs an info-level message.
func Error(message string, args ...any) {
	GetLogger().Error(message, args...)
}

// parseLogLevel converts the configured log level string into slog level.
func parsedLogLevel(loglevel string) slog.Level {
	switch strings.ToLower(loglevel) {
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
