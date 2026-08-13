package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const defaultAuthBcryptCost = 12

// AuthConfig contains authentication-related application configuration.
type AuthConfig struct {
	BcryptCost int
}

// LoadAuthConfig loads authentication configuration from environment
// variables.
func LoadAuthConfig() (AuthConfig, error) {
	bcryptCost, err := authIntEnv(
		"AUTH_BCRYPT_COST",
		defaultAuthBcryptCost,
	)
	if err != nil {
		return AuthConfig{}, err
	}

	authConfig := AuthConfig{
		BcryptCost: bcryptCost,
	}

	if err := authConfig.Validate(); err != nil {
		return AuthConfig{}, err
	}

	return authConfig, nil
}

// Validate verifies authentication configuration values.
func (config AuthConfig) Validate() error {
	if config.BcryptCost < bcrypt.MinCost ||
		config.BcryptCost > bcrypt.MaxCost {
		return fmt.Errorf(
			"AUTH_BCRYPT_COST must be between %d and %d",
			bcrypt.MinCost,
			bcrypt.MaxCost,
		)
	}

	return nil
}

func authIntEnv(
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
