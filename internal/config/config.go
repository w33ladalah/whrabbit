package config

import (
	"os"
)

var (
	// Build-time variables
	AppName    string
	AppVersion string
	APIKey     string
	BaseURL    string
	ServerPort string
	AppEnv     string
)

func GetAppName() string {
	if AppName != "" {
		return AppName
	}
	return os.Getenv("APP_NAME")
}

func GetAppVersion() string {
	if AppVersion != "" {
		return AppVersion
	}
	return os.Getenv("APP_VERSION")
}

func GetAPIKey() string {
	if APIKey != "" {
		return APIKey
	}
	return os.Getenv("API_KEY")
}

func GetBaseURL() string {
	if BaseURL != "" {
		return BaseURL
	}
	return os.Getenv("BASE_URL")
}

func GetServerPort() string {
	if ServerPort != "" {
		return ServerPort
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	return port
}

func GetAppEnv() string {
	if AppEnv != "" {
		return AppEnv
	}
	return os.Getenv("APP_ENV")
}
