// Package config handles application configuration loading from environment variables
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server       ServerConfig
	MongoDB      MongoDBConfig
	JWT          JWTConfig
	Encryption   EncryptionConfig
	Notification NotificationConfig
	Energy       EnergyProviderConfig
	Storage      StorageServiceConfig
	Logging      LoggingConfig
}

// StorageServiceConfig holds Storage service integration settings
type StorageServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
	Mode string
}

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// EncryptionConfig holds encryption settings
type EncryptionConfig struct {
	Key string
}

// NotificationConfig holds notification service URLs
type NotificationConfig struct {
	EmailURL string
	SMSURL   string
	PushURL  string
}

// EnergyProviderConfig holds external energy provider settings
type EnergyProviderConfig struct {
	BaseURL      string
	APIKey       string
	ClientID     string
	ClientSecret string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load reads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "security_service"),
			Timeout:  time.Duration(getEnvAsInt("MONGODB_TIMEOUT", 10)) * time.Second,
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "default-secret-change-me"),
			AccessTokenExpiry:  parseDuration(getEnv("JWT_ACCESS_TOKEN_EXPIRY", "15m")),
			RefreshTokenExpiry: parseDuration(getEnv("JWT_REFRESH_TOKEN_EXPIRY", "168h")), // 7 days
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", "32-byte-encryption-key-here!!!!"),
		},
		Notification: NotificationConfig{
			EmailURL: getEnv("NOTIFICATION_EMAIL_URL", "http://localhost:8081/external/notifications/email"),
			SMSURL:   getEnv("NOTIFICATION_SMS_URL", "http://localhost:8081/external/notifications/sms"),
			PushURL:  getEnv("NOTIFICATION_PUSH_URL", "http://localhost:8081/external/notifications/push"),
		},
		Energy: EnergyProviderConfig{
			BaseURL:      getEnv("ENERGY_PROVIDER_BASE_URL", "https://api.energy-provider.com"),
			APIKey:       getEnv("ENERGY_PROVIDER_API_KEY", ""),
			ClientID:     getEnv("ENERGY_PROVIDER_CLIENT_ID", ""),
			ClientSecret: getEnv("ENERGY_PROVIDER_CLIENT_SECRET", ""),
		},
		Storage: StorageServiceConfig{
			URL:     getEnv("STORAGE_SERVICE_URL", "http://localhost:8086/storage"),
			Timeout: time.Duration(getEnvAsInt("STORAGE_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// getEnv retrieves an environment variable with a default fallback
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt retrieves an environment variable as an integer
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// parseDuration parses a duration string with fallback
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute // Default fallback
	}
	return d
}
