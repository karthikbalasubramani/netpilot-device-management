package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName       string
	AppEnv        string
	AppPort       string
	MongoURI      string
	MongoDatabase string
	JWTSecret     string
	MQTTBroker    string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, Using System Environment Variables")
	}

	return &Config{
		AppName:       getEnv("APP_NAME", "NetPilot"),
		AppEnv:        getEnv("APP_ENV", "local"),
		AppPort:       getEnv("APP_PORT", "8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "netpilot_db"),
		JWTSecret:     getEnv("JWT_SECRET", "Secret_Token_NetPilot_31011998"),
		MQTTBroker:    getEnv("MQTT_BROKER", "tcp://localhost:1883"),
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
