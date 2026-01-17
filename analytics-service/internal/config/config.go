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
	Forecast  ForecastServiceConfig
	Storage   StorageServiceConfig
	Analytics AnalyticsConfig
	Logging   LoggingConfig
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

// ForecastServiceConfig holds Forecast service integration settings
type ForecastServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// AnalyticsConfig holds analytics-specific settings
type AnalyticsConfig struct {
	AnomalyDetectionEnabled       bool
	KPICalculationInterval        time.Duration
	ReportRetentionDays           int
	TimeSeriesAggregationInterval time.Duration
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
			Port: getEnv("SERVER_PORT", "8084"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "analytics_service"),
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
		Forecast: ForecastServiceConfig{
			URL:     getEnv("FORECAST_SERVICE_URL", "http://localhost:8082"),
			Timeout: time.Duration(getEnvAsInt("FORECAST_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		Storage: StorageServiceConfig{
			URL:     getEnv("STORAGE_SERVICE_URL", "http://localhost:8086/storage"),
			Timeout: time.Duration(getEnvAsInt("STORAGE_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		Analytics: AnalyticsConfig{
			AnomalyDetectionEnabled:       getEnvAsBool("ANALYTICS_ANOMALY_DETECTION_ENABLED", true),
			KPICalculationInterval:        time.Duration(getEnvAsInt("ANALYTICS_KPI_CALCULATION_INTERVAL", 60)) * time.Minute,
			ReportRetentionDays:           getEnvAsInt("ANALYTICS_REPORT_RETENTION_DAYS", 90),
			TimeSeriesAggregationInterval: time.Duration(getEnvAsInt("ANALYTICS_TIME_SERIES_AGGREGATION_INTERVAL", 60)) * time.Minute,
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

// getEnvAsBool retrieves an environment variable as a boolean
func getEnvAsBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
