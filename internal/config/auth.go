package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	defaultAuthBcryptCost           = 12
	defaultJWTIssuer                = "netpilot"
	defaultJWTAudience              = "netpilot-api"
	defaultJWTAccessTokenTTLMinutes = 15
	minimumJWTSecretLengthBytes     = 32
)

// AuthConfig contains authentication-related application configuration.
type AuthConfig struct {
	BcryptCost int

	JWTSecret                string
	JWTIssuer                string
	JWTAudience              string
	JWTAccessTokenTTLMinutes time.Duration
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
	accessTokenTTLMinutes, err := authIntEnv(
		"JWT_ACCESS_TOKEN_TTL_MINUTES",
		defaultJWTAccessTokenTTLMinutes,
	)
	if err != nil {
		return AuthConfig{}, err
	}

	authConfig := AuthConfig{
		BcryptCost: bcryptCost,

		JWTSecret: strings.TrimSpace(
			os.Getenv("JWT_SECRET"),
		),
		JWTIssuer: authStringEnv(
			"JWT_ISSUER",
			defaultJWTIssuer,
		),
		JWTAudience: authStringEnv(
			"JWT_AUDIENCE",
			defaultJWTAudience,
		),
		JWTAccessTokenTTLMinutes: time.Duration(
			accessTokenTTLMinutes,
		) * time.Minute,
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
	if len([]byte(config.JWTSecret)) <
		minimumJWTSecretLengthBytes {
		return fmt.Errorf(
			"JWT_SECRET must contain at least %d bytes",
			minimumJWTSecretLengthBytes,
		)
	}
	if strings.TrimSpace(config.JWTIssuer) == "" {
		return fmt.Errorf(
			"JWT_ISSUER is required",
		)
	}
	if strings.TrimSpace(config.JWTAudience) == "" {
		return fmt.Errorf(
			"JWT_AUDIENCE is required",
		)
	}
	if config.JWTAccessTokenTTLMinutes <= 0 {
		return fmt.Errorf(
			"JWT_ACCESS_TOKEN_TTL_MINUTES must be greater than zero",
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

func authStringEnv(
	key string,
	defaultValue string,
) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return defaultValue
	}

	return value
}
