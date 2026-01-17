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
	Forecast  ForecastServiceConfig
	Analytics AnalyticsServiceConfig
	MQTT      MQTTConfig
	IoT       IoTConfig
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

// ForecastServiceConfig holds Forecast service integration settings
type ForecastServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// AnalyticsServiceConfig holds Analytics service integration settings
type AnalyticsServiceConfig struct {
	URL     string
	Timeout time.Duration
}

// MQTTConfig holds MQTT broker configuration
type MQTTConfig struct {
	Broker   string
	Port     int
	Username string
	Password string
	ClientID string
	QoS      byte
}

// IoTConfig holds IoT-specific settings
type IoTConfig struct {
	TelemetryBatchSize int
	CommandTimeout     time.Duration
	StateUpdateInterval time.Duration
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
			Port: getEnv("SERVER_PORT", "8083"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "iot_control_service"),
			Timeout:  time.Duration(getEnvAsInt("MONGODB_TIMEOUT", 10)) * time.Second,
		},
		Security: SecurityServiceConfig{
			URL:     getEnv("SECURITY_SERVICE_URL", "http://localhost:8080"),
			Timeout: time.Duration(getEnvAsInt("SECURITY_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		Forecast: ForecastServiceConfig{
			URL:     getEnv("FORECAST_SERVICE_URL", "http://localhost:8082"),
			Timeout: time.Duration(getEnvAsInt("FORECAST_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		Analytics: AnalyticsServiceConfig{
			URL:     getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8084"),
			Timeout: time.Duration(getEnvAsInt("ANALYTICS_SERVICE_TIMEOUT", 10)) * time.Second,
		},
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "localhost"),
			Port:     getEnvAsInt("MQTT_PORT", 1883),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
			ClientID: getEnv("MQTT_CLIENT_ID", "iot-control-service"),
			QoS:      byte(getEnvAsInt("MQTT_QOS", 1)),
		},
		IoT: IoTConfig{
			TelemetryBatchSize:  getEnvAsInt("IOT_TELEMETRY_BATCH_SIZE", 100),
			CommandTimeout:      time.Duration(getEnvAsInt("IOT_COMMAND_TIMEOUT", 30)) * time.Second,
			StateUpdateInterval: time.Duration(getEnvAsInt("IOT_STATE_UPDATE_INTERVAL", 5)) * time.Second,
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
