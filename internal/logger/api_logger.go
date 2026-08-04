package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// APILoggerConfig contains settings for the dedicated API access logger.
type APILoggerConfig struct {
	Enabled    bool
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

var (
	apiLoggerMutex sync.RWMutex
	apiLogger      *slog.Logger
	apiLogWriter   *lumberjack.Logger
)

// InitAPILogger initializes the dedicated API access logger.
//
// API access logs are written only to the configured file. They are not
// written to the terminal.
func InitAPILogger(config APILoggerConfig) error {
	if !config.Enabled {
		return disableAPILogger()
	}

	filePath := strings.TrimSpace(config.FilePath)
	logDirectory := filepath.Dir(filePath)

	if err := os.MkdirAll(logDirectory, 0755); err != nil {
		return fmt.Errorf(
			"create API log directory %q: %w",
			logDirectory,
			err,
		)
	}

	if err := verifyAPILogFile(filePath); err != nil {
		return err
	}

	newWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    config.MaxSizeMB,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAgeDays,
		LocalTime:  true,
		Compress:   config.Compress,
	}

	newLogger := slog.New(
		slog.NewJSONHandler(
			newWriter,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	apiLoggerMutex.Lock()

	previousWriter := apiLogWriter
	apiLogger = newLogger
	apiLogWriter = newWriter

	apiLoggerMutex.Unlock()

	if previousWriter != nil {
		if err := previousWriter.Close(); err != nil {
			return fmt.Errorf(
				"close previous API log writer: %w",
				err,
			)
		}
	}

	return nil
}

// CloseAPILogger closes the dedicated API log writer.
func CloseAPILogger() error {
	apiLoggerMutex.Lock()

	writer := apiLogWriter
	apiLogger = nil
	apiLogWriter = nil

	apiLoggerMutex.Unlock()

	if writer == nil {
		return nil
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close API log writer: %w", err)
	}

	return nil
}

// APIInfo writes an informational HTTP access event to the API log file.
func APIInfo(message string, attributes ...any) {
	writeAPILog(slog.LevelInfo, message, attributes...)
}

// APIWarn writes a client-error HTTP access event to the API log file.
func APIWarn(message string, attributes ...any) {
	writeAPILog(slog.LevelWarn, message, attributes...)
}

// APIError writes a server-error HTTP access event to the API log file.
func APIError(message string, attributes ...any) {
	writeAPILog(slog.LevelError, message, attributes...)
}

func writeAPILog(
	level slog.Level,
	message string,
	attributes ...any,
) {
	apiLoggerMutex.RLock()
	defer apiLoggerMutex.RUnlock()

	if apiLogger == nil {
		return
	}

	apiLogger.Log(
		context.Background(),
		level,
		message,
		attributes...,
	)
}

func verifyAPILogFile(filePath string) error {
	file, err := os.OpenFile(
		filePath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return fmt.Errorf(
			"open API log file %q: %w",
			filePath,
			err,
		)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf(
			"close API log file validation handle %q: %w",
			filePath,
			err,
		)
	}

	return nil
}

func disableAPILogger() error {
	apiLoggerMutex.Lock()

	previousWriter := apiLogWriter
	apiLogger = nil
	apiLogWriter = nil

	apiLoggerMutex.Unlock()

	if previousWriter == nil {
		return nil
	}

	if err := previousWriter.Close(); err != nil {
		return fmt.Errorf(
			"close disabled API log writer: %w",
			err,
		)
	}

	return nil
}
