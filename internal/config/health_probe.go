package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultHealthReadinessTimeoutSeconds = 3

// HealthProbeConfig contains settings for application health probes.
type HealthProbeConfig struct {
	ReadinessTimeout time.Duration
}

// LoadHealthProbeConfig loads health-probe settings from environment
// variables.
//
// config.Load must run before this function so godotenv has already loaded
// values from .env into the process environment.
func LoadHealthProbeConfig() (HealthProbeConfig, error) {
	readinessTimeoutSeconds, err := healthProbeIntEnv(
		"HEALTH_READINESS_TIMEOUT_SECONDS",
		defaultHealthReadinessTimeoutSeconds,
	)
	if err != nil {
		return HealthProbeConfig{}, err
	}

	healthProbeConfig := HealthProbeConfig{
		ReadinessTimeout: time.Duration(
			readinessTimeoutSeconds,
		) * time.Second,
	}

	if err := healthProbeConfig.Validate(); err != nil {
		return HealthProbeConfig{}, err
	}

	return healthProbeConfig, nil
}

// Validate verifies that health-probe configuration values are valid.
func (config HealthProbeConfig) Validate() error {
	if config.ReadinessTimeout <= 0 {
		return fmt.Errorf(
			"HEALTH_READINESS_TIMEOUT_SECONDS must be greater than zero",
		)
	}

	return nil
}

func healthProbeIntEnv(
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
