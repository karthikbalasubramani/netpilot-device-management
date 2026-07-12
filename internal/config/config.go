package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values loaded from environment variables.
type Config struct {
	AppName             string
	AppEnv              string
	AppPort             string
	MongoURI            string
	MongoDatabase       string
	JWTSecret           string
	MQTTBroker          string
	CPUThresholdPercent float64
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
		MongoURI:            getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:       getEnv("MONGO_DATABASE", "netpilot_db"),
		JWTSecret:           getEnv("JWT_SECRET", "change_me"),
		MQTTBroker:          getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		CPUThresholdPercent: getEnvAsFloat("CPU_THRESHOLD_PERCENT", 60),
	}
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
