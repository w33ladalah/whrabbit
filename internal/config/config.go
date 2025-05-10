package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	// Build-time variables
	AppName    string
	AppVersion string
	APIKey     string
	BaseURL    string
	ServerPort string
)

type Config struct {
	AppName    string
	AppVersion string
	APIKey     string
	BaseURL    string
	ServerPort string
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// If build-time variables are not set (development mode), use environment variables
	if APIKey == "" {
		APIKey = getEnv("API_KEY", "")
	}
	if BaseURL == "" {
		BaseURL = getEnv("BASE_URL", "http://localhost:8080")
	}
	if ServerPort == "" {
		ServerPort = getEnv("PORT", "8080")
	}
	if AppName == "" {
		AppName = getEnv("APP_NAME", "whrabbit")
	}
	if AppVersion == "" {
		AppVersion = getEnv("APP_VERSION", "dev")
	}

	return &Config{
		AppName:    AppName,
		AppVersion: AppVersion,
		APIKey:     APIKey,
		BaseURL:    BaseURL,
		ServerPort: ServerPort,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
