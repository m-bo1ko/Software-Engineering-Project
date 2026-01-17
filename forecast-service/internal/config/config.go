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
	Server    ServerConfig
	MongoDB   MongoDBConfig
	Security  SecurityServiceConfig
	IoT       IoTServiceConfig
	External  ExternalAPIsConfig
	Forecast  ForecastConfig
	Logging   LoggingConfig
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

// SecurityServiceConfig holds Security service integration settings
type SecurityServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// IoTServiceConfig holds IoT service integration settings
type IoTServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// ExternalAPIsConfig holds external API endpoints
type ExternalAPIsConfig struct {
	WeatherURL string
	TariffURL  string
	MLURL      string
	StorageURL string
}

// ForecastConfig holds forecast-specific settings
type ForecastConfig struct {
	DefaultHorizonHours      int
	MaxHorizonHours          int
	PeakLoadThresholdPercent float64
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
			Port: getEnv("SERVER_PORT", "8082"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "forecast_service"),
			Timeout:  time.Duration(getEnvAsInt("MONGODB_TIMEOUT", 10)) * time.Second,
		},
		Security: SecurityServiceConfig{
			URL:     getEnv("SECURITY_SERVICE_URL", "http://localhost:8080"),
			Timeout: time.Duration(getEnvAsInt("SECURITY_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		IoT: IoTServiceConfig{
			URL:     getEnv("IOT_SERVICE_URL", "http://localhost:8083"),
			Timeout: time.Duration(getEnvAsInt("IOT_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		External: ExternalAPIsConfig{
			WeatherURL: getEnv("WEATHER_API_URL", "http://localhost:8084/external/weather"),
			TariffURL:  getEnv("TARIFF_API_URL", "http://localhost:8084/external/tariffs"),
			MLURL:      getEnv("ML_MODEL_URL", "http://localhost:8085/ml/predict"),
			StorageURL: getEnv("STORAGE_API_URL", "http://localhost:8086/storage"),
		},
		Forecast: ForecastConfig{
			DefaultHorizonHours:      getEnvAsInt("FORECAST_DEFAULT_HORIZON_HOURS", 24),
			MaxHorizonHours:          getEnvAsInt("FORECAST_MAX_HORIZON_HOURS", 168),
			PeakLoadThresholdPercent: getEnvAsFloat("PEAK_LOAD_THRESHOLD_PERCENTAGE", 80.0),
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

// getEnvAsFloat retrieves an environment variable as a float
func getEnvAsFloat(key string, defaultVal float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}
