package config

import (
	"os"
)

var (
	// Build-time variables
	APIKey     string
	BaseURL    string
	ServerPort string
)

type Config struct {
	APIKey     string
	BaseURL    string
	ServerPort string
}

func LoadConfig() *Config {
	return &Config{
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
