package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAPILogEnabled    = true
	defaultAPILogFilePath   = "logs/api.log"
	defaultAPILogMaxSizeMB  = 10
	defaultAPILogMaxBackups = 5
	defaultAPILogMaxAgeDays = 30
	defaultAPILogCompress   = true
)

// APILogConfig contains configuration for the dedicated HTTP API access log.
type APILogConfig struct {
	Enabled    bool
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

// LoadAPILogConfig loads API access-log configuration from environment
// variables.
func LoadAPILogConfig() (APILogConfig, error) {
	enabled, err := apiLogBoolEnv(
		"API_LOG_ENABLED",
		defaultAPILogEnabled,
	)
	if err != nil {
		return APILogConfig{}, err
	}

	maxSizeMB, err := apiLogIntEnv(
		"API_LOG_MAX_SIZE_MB",
		defaultAPILogMaxSizeMB,
	)
	if err != nil {
		return APILogConfig{}, err
	}

	maxBackups, err := apiLogIntEnv(
		"API_LOG_MAX_BACKUPS",
		defaultAPILogMaxBackups,
	)
	if err != nil {
		return APILogConfig{}, err
	}

	maxAgeDays, err := apiLogIntEnv(
		"API_LOG_MAX_AGE_DAYS",
		defaultAPILogMaxAgeDays,
	)
	if err != nil {
		return APILogConfig{}, err
	}

	compress, err := apiLogBoolEnv(
		"API_LOG_COMPRESS",
		defaultAPILogCompress,
	)
	if err != nil {
		return APILogConfig{}, err
	}

	apiLogConfig := APILogConfig{
		Enabled: enabled,
		FilePath: apiLogStringEnv(
			"API_LOG_FILE_PATH",
			defaultAPILogFilePath,
		),
		MaxSizeMB:  maxSizeMB,
		MaxBackups: maxBackups,
		MaxAgeDays: maxAgeDays,
		Compress:   compress,
	}

	if err := apiLogConfig.Validate(); err != nil {
		return APILogConfig{}, err
	}

	return apiLogConfig, nil
}

// Validate verifies the API access-log configuration.
func (config APILogConfig) Validate() error {
	if !config.Enabled {
		return nil
	}

	if strings.TrimSpace(config.FilePath) == "" {
		return fmt.Errorf(
			"API_LOG_FILE_PATH cannot be empty when API logging is enabled",
		)
	}

	if config.MaxSizeMB <= 0 {
		return fmt.Errorf(
			"API_LOG_MAX_SIZE_MB must be greater than zero",
		)
	}

	if config.MaxBackups < 0 {
		return fmt.Errorf(
			"API_LOG_MAX_BACKUPS cannot be negative",
		)
	}

	if config.MaxAgeDays < 0 {
		return fmt.Errorf(
			"API_LOG_MAX_AGE_DAYS cannot be negative",
		)
	}

	return nil
}

func apiLogStringEnv(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return defaultValue
	}

	return value
}

func apiLogBoolEnv(key string, defaultValue bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return defaultValue, nil
	}

	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf(
			"%s must be a valid boolean: %w",
			key,
			err,
		)
	}

	return parsedValue, nil
}

func apiLogIntEnv(key string, defaultValue int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return defaultValue, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be a valid integer: %w",
			key,
			err,
		)
	}

	return parsedValue, nil
}
