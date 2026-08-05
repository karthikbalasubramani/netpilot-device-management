package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPReadHeaderTimeoutSeconds = 5
	defaultHTTPReadTimeoutSeconds       = 15
	defaultHTTPWriteTimeoutSeconds      = 30
	defaultHTTPIdleTimeoutSeconds       = 60
	defaultHTTPMaxHeaderBytes           = 1 << 20
	defaultHTTPMaxRequestBodyBytes      = 1 << 20
)

// HTTPServerConfig contains timeout and request-size settings for the
// NetPilot HTTP server.
type HTTPServerConfig struct {
	ReadHeaderTimeout   time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	MaxHeaderBytes      int
	MaxRequestBodyBytes int64
}

// LoadHTTPServerConfig loads HTTP server configuration from environment
// variables.
//
// config.Load must run before this function so godotenv has already loaded
// values from .env into the process environment.
func LoadHTTPServerConfig() (HTTPServerConfig, error) {
	readHeaderTimeoutSeconds, err := httpServerIntEnv(
		"HTTP_READ_HEADER_TIMEOUT_SECONDS",
		defaultHTTPReadHeaderTimeoutSeconds,
	)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	readTimeoutSeconds, err := httpServerIntEnv(
		"HTTP_READ_TIMEOUT_SECONDS",
		defaultHTTPReadTimeoutSeconds,
	)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	writeTimeoutSeconds, err := httpServerIntEnv(
		"HTTP_WRITE_TIMEOUT_SECONDS",
		defaultHTTPWriteTimeoutSeconds,
	)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	idleTimeoutSeconds, err := httpServerIntEnv(
		"HTTP_IDLE_TIMEOUT_SECONDS",
		defaultHTTPIdleTimeoutSeconds,
	)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	maxHeaderBytes, err := httpServerIntEnv(
		"HTTP_MAX_HEADER_BYTES",
		defaultHTTPMaxHeaderBytes,
	)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	maxRequestBodyBytes, err := httpServerInt64Env(
		"HTTP_MAX_REQUEST_BODY_BYTES",
		defaultHTTPMaxRequestBodyBytes,
	)
	if err != nil {
		return HTTPServerConfig{}, err
	}

	serverConfig := HTTPServerConfig{
		ReadHeaderTimeout: time.Duration(
			readHeaderTimeoutSeconds,
		) * time.Second,
		ReadTimeout: time.Duration(
			readTimeoutSeconds,
		) * time.Second,
		WriteTimeout: time.Duration(
			writeTimeoutSeconds,
		) * time.Second,
		IdleTimeout: time.Duration(
			idleTimeoutSeconds,
		) * time.Second,
		MaxHeaderBytes:      maxHeaderBytes,
		MaxRequestBodyBytes: maxRequestBodyBytes,
	}

	if err := serverConfig.Validate(); err != nil {
		return HTTPServerConfig{}, err
	}

	return serverConfig, nil
}

// Validate verifies that the HTTP server configuration is safe and valid.
func (config HTTPServerConfig) Validate() error {
	if config.ReadHeaderTimeout <= 0 {
		return fmt.Errorf(
			"HTTP_READ_HEADER_TIMEOUT_SECONDS must be greater than zero",
		)
	}

	if config.ReadTimeout <= 0 {
		return fmt.Errorf(
			"HTTP_READ_TIMEOUT_SECONDS must be greater than zero",
		)
	}

	if config.WriteTimeout <= 0 {
		return fmt.Errorf(
			"HTTP_WRITE_TIMEOUT_SECONDS must be greater than zero",
		)
	}

	if config.IdleTimeout <= 0 {
		return fmt.Errorf(
			"HTTP_IDLE_TIMEOUT_SECONDS must be greater than zero",
		)
	}

	if config.MaxHeaderBytes <= 0 {
		return fmt.Errorf(
			"HTTP_MAX_HEADER_BYTES must be greater than zero",
		)
	}

	if config.MaxRequestBodyBytes <= 0 {
		return fmt.Errorf(
			"HTTP_MAX_REQUEST_BODY_BYTES must be greater than zero",
		)
	}

	if config.ReadHeaderTimeout > config.ReadTimeout {
		return fmt.Errorf(
			"HTTP_READ_HEADER_TIMEOUT_SECONDS cannot be greater " +
				"than HTTP_READ_TIMEOUT_SECONDS",
		)
	}

	return nil
}

func httpServerIntEnv(
	key string,
	defaultValue int,
) (int, error) {
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

func httpServerInt64Env(
	key string,
	defaultValue int64,
) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return defaultValue, nil
	}

	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be a valid integer: %w",
			key,
			err,
		)
	}

	return parsedValue, nil
}
