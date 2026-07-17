package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values loaded from environment variables.
type Config struct {
	AppName             string
	AppEnv              string
	AppPort             string
	LogLevel            string
	MongoURI            string
	MongoDatabase       string
	JWTSecret           string
	MQTTBroker          string
	CPUThresholdPercent float64
	DiskPath            string
}

// Load loads application configuration from the .env file or system environment variables.
// If any environment variable is missing, a default value is used.
func Load() *Config {
	// Load environment variables from .env file for local development.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Build and return application configuration with fallback default values.
	return &Config{
		AppName:             getEnv("APP_NAME", "NetPilot"),
		AppEnv:              getEnv("APP_ENV", "local"),
		AppPort:             getEnv("APP_PORT", "8080"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		MongoURI:            getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:       getEnv("MONGO_DATABASE", "netpilot_db"),
		JWTSecret:           getEnv("JWT_SECRET", "change_me"),
		MQTTBroker:          getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		CPUThresholdPercent: getEnvAsFloat("CPU_THRESHOLD_PERCENT", 60),
		DiskPath:            getEnv("DISK_PATH", "/"),
	}
}

// Env Values Validation function
// Returns []string for validation errors
func (c *Config) ValidateEnvConfiguration() error {
	var validationErrors []string

	// App Name Validation
	if strings.TrimSpace(c.AppName) == "" {
		validationErrors = append(validationErrors, "Configured APP Name is invalid in Env file")
	}

	// App Env Validation
	if !VerifyAppEnv(c.AppEnv) {
		validationErrors = append(validationErrors, "Configured APP Env is invalid in Env file")
	}

	// App Port Number Validation
	if err := VerifyAppPort(c.AppPort); err != nil {
		validationErrors = append(
			validationErrors,
			fmt.Sprintf("Configured App Port Number is failed with error: %v", err),
		)
	}

	// Log Level Validation
	if !VerifyLoggerLogLevel(c.LogLevel) {
		validationErrors = append(validationErrors, "Configured Log Level is invalid in Env file")
	}

	// MongoURI Validation
	if strings.TrimSpace(c.MongoURI) == "" {
		validationErrors = append(validationErrors, "Cnfigured MongoURI is invalid in Env file")
	}

	// Mongo Database Validation
	if strings.TrimSpace(c.MongoDatabase) == "" {
		validationErrors = append(validationErrors, "Configured Mongo database value is invalid")
	}

	// JWT_Secret Validation
	if strings.TrimSpace(c.JWTSecret) == "" {
		validationErrors = append(validationErrors, "JWT Secret Token Value is invalid in Env file")
	}

	// MQTT Broker Validation
	if strings.TrimSpace(c.MQTTBroker) == "" {
		validationErrors = append(validationErrors, "MQTT Broker value is invalid in Env file")
	}

	// JWT Secret Validation in Production
	if c.AppEnv == "production" && c.JWTSecret == "change_me" {
		validationErrors = append(validationErrors, "JWT_SECRET must not use default value in production")
	}

	// CPU Threshold Percent Validation
	if c.CPUThresholdPercent <= 0 || c.CPUThresholdPercent > 100 {
		validationErrors = append(
			validationErrors,
			"CPU_THRESHOLD_PERCENT must be greater than 0 and less than or equal to 100",
		)
	}

	// Disk Path Validation
	if strings.TrimSpace(c.DiskPath) == "" {
		validationErrors = append(validationErrors, "DISK_PATH is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("invalid configuration: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// Validate App Env in Configuration
func VerifyAppEnv(appName string) bool {
	switch strings.ToLower(appName) {
	case "dev", "production", "test", "staging", "local":
		return true
	default:
		return false
	}
}

// Validate Logger Log Level in Configuration
func VerifyLoggerLogLevel(logLevel string) bool {
	switch strings.ToLower(logLevel) {
	case "info", "warn", "debug", "error":
		return true
	default:
		return false
	}
}

// Validate App Port Number in Configuration
func VerifyAppPort(appport string) error {
	if strings.TrimSpace(appport) == "" {
		return fmt.Errorf("App Port Number cannot be empty")
	}
	portNumber, err := strconv.Atoi(appport)
	if err != nil {
		return fmt.Errorf("App Port Number is not valid: %v", err)
	}
	if portNumber < 0 || portNumber > 65535 {
		return fmt.Errorf("App Port Number is not in range")
	}
	return nil
}

// getEnv returns the value of the environment variable identified by key.
// If the variable is not set or has an empty value, it returns defaultValue.
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

// getEnvAsFloat returns the environment variable value as float64.
// If the variable is missing or invalid, it returns defaultValue.
func getEnvAsFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Printf("Invalid value for %s, using default value %.2f", key, defaultValue)
		return defaultValue
	}

	return parsedValue
}
