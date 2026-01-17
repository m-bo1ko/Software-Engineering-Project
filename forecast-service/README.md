# Forecast & Optimization Service

A comprehensive microservice for energy demand forecasting, peak load prediction, and optimization scenario generation. This service integrates seamlessly with the Security & External Integration service for authentication, authorization, and audit logging.

## Features

- **Energy Demand Forecasting**: Generate accurate forecasts using historical data, weather, and tariffs
- **Peak Load Prediction**: Identify periods of abnormally high energy consumption
- **Optimization Scenarios**: Create optimal control scenarios for energy savings
- **Recommendations**: Provide actionable energy-saving recommendations
- **Device-Level Predictions**: Get consumption predictions and optimization suggestions for specific devices
- **IoT Integration**: Send optimization scenarios to IoT & Control service for execution

## Technology Stack

- **Language**: Go (Golang) 1.21+
- **Framework**: Gin
- **Database**: MongoDB
- **Authentication**: JWT via Security Service
- **Containerization**: Docker

## Project Structure

```
forecast-service/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── handlers/
│   │   ├── forecast_handler.go    # Forecast endpoints
│   │   ├── optimization_handler.go # Optimization endpoints
│   │   └── router.go              # Route configuration
│   ├── middleware/
│   │   ├── auth.go                # JWT authentication middleware
│   │   └── common.go              # Common middleware (CORS, logging, etc.)
│   ├── models/                    # Data models
│   ├── repository/                # Database operations
│   ├── service/                   # Business logic
│   └── integrations/              # External service clients
├── pkg/
│   └── utils/                     # Utility functions
├── tests/                         # Test files
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.21+
- MongoDB 7.0+
- Security Service running (for authentication)
- Docker & Docker Compose (optional)

### Running with Docker

```bash
# Clone the repository
cd forecast-service

# Copy and configure environment
cp .env.example .env
# Edit .env with your configuration

# Start services
docker-compose up -d

# View logs
docker-compose logs -f forecast-service
```

### Running Locally

```bash
# Install dependencies
go mod download

# Set environment variables (or create .env file)
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=forecast_service
export SECURITY_SERVICE_URL=http://localhost:8080
export IOT_SERVICE_URL=http://localhost:8083

# Run the application
go run cmd/main.go
```

## API Endpoints

### Forecasting

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/forecast/generate` | Generate energy demand forecast | Yes |
| POST | `/forecast/peak-load` | Predict peak load periods | Yes |
| GET | `/forecast/latest?buildingId=` | Get latest forecast for building | Yes |
| GET | `/forecast/prediction/:deviceId` | Get device consumption prediction | Yes |
| GET | `/forecast/optimization/:deviceId` | Get device optimization recommendations | Yes |

### Optimization

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/optimization/generate` | Generate optimization scenario | Yes |
| GET | `/optimization/recommendations/:buildingId` | Get energy-saving recommendations | Yes |
| GET | `/optimization/scenario/:scenarioId` | Get scenario details | Yes |
| POST | `/optimization/send-to-iot` | Send scenario to IoT service | Yes |

## Example Requests

### Generate Forecast

```bash
curl -X POST http://localhost:8082/forecast/generate \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "buildingId": "building-123",
    "type": "DEMAND",
    "horizonHours": 24,
    "includeWeather": true,
    "includeTariffs": true,
    "historicalDays": 30
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Forecast generated successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "buildingId": "building-123",
    "type": "DEMAND",
    "status": "COMPLETED",
    "horizonHours": 24,
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "2024-01-02T00:00:00Z",
    "predictions": [
      {
        "timestamp": "2024-01-01T00:00:00Z",
        "predictedValue": 75.5,
        "lowerBound": 65.2,
        "upperBound": 85.8,
        "confidenceLevel": 0.95,
        "unit": "kW"
      }
    ],
    "accuracy": {
      "mae": 15.5,
      "rmse": 20.3,
      "mape": 8.2,
      "score": 78.0
    },
    "modelUsed": "STATISTICAL",
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

### Generate Peak Load Prediction

```bash
curl -X POST http://localhost:8082/forecast/peak-load \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "buildingId": "building-123",
    "analysisFromDate": "2024-01-01T00:00:00Z",
    "analysisToDate": "2024-01-02T00:00:00Z",
    "thresholdPercent": 80.0,
    "includeWeather": true
  }'
```

### Generate Optimization Scenario

```bash
curl -X POST http://localhost:8082/optimization/generate \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "buildingId": "building-123",
    "type": "COST_REDUCTION",
    "scheduledStart": "2024-01-01T14:00:00Z",
    "scheduledEnd": "2024-01-01T18:00:00Z",
    "useTariffData": true,
    "useWeatherData": true,
    "priority": 5,
    "constraints": {
      "preserveComfort": true,
      "minTemperature": 20.0,
      "maxTemperature": 26.0
    }
  }'
```

### Get Device Optimization

```bash
curl -X GET http://localhost:8082/forecast/optimization/device-123 \
  -H "Authorization: Bearer <access_token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "deviceId": "device-123",
    "deviceName": "Device device-123",
    "currentState": "ON",
    "optimalState": "ECO_MODE",
    "potentialSavings": 24.0,
    "savingsUnit": "kWh/day",
    "recommendations": [
      "Consider increasing setpoint by 1-2°C during peak hours",
      "Enable pre-cooling before peak tariff periods"
    ],
    "scheduledActions": [
      {
        "time": "2024-01-01T16:00:00Z",
        "action": "REDUCE_POWER",
        "targetState": "ECO_MODE",
        "reason": "Peak tariff period approaching"
      }
    ],
    "priority": "MEDIUM"
  }
}
```

### Send Scenario to IoT Service

```bash
curl -X POST http://localhost:8082/optimization/send-to-iot \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "scenarioId": "507f1f77bcf86cd799439011",
    "executeNow": false,
    "dryRun": false
  }'
```

### Get Recommendations

```bash
curl -X GET http://localhost:8082/optimization/recommendations/building-123 \
  -H "Authorization: Bearer <access_token>"
```

## Configuration

The application can be configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8082` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `forecast_service` |
| `SECURITY_SERVICE_URL` | Security service URL | `http://localhost:8080` |
| `IOT_SERVICE_URL` | IoT service URL | `http://localhost:8083` |
| `WEATHER_API_URL` | Weather API URL | `http://localhost:8084/external/weather` |
| `TARIFF_API_URL` | Tariff API URL | `http://localhost:8084/external/tariffs` |
| `ML_MODEL_URL` | ML model service URL | `http://localhost:8085/ml/predict` |
| `STORAGE_API_URL` | Storage service URL | `http://localhost:8086/storage` |
| `FORECAST_DEFAULT_HORIZON_HOURS` | Default forecast horizon | `24` |
| `FORECAST_MAX_HORIZON_HOURS` | Maximum forecast horizon | `168` |
| `PEAK_LOAD_THRESHOLD_PERCENTAGE` | Peak load threshold | `80.0` |

## Integration with Security Service

All endpoints require JWT authentication via the Security & External Integration service. The service:

1. Validates tokens using `/auth/validate-token`
2. Logs audit events to `/audit/log`
3. Checks permissions using `/auth/check-permissions` (optional)

## Integration with IoT Service

The service integrates with IoT & Control service to:

1. Get device states via `/iot/state/:deviceId`
2. Get devices by building via `/iot/devices?buildingId=`
3. Apply optimizations via `/iot/optimization/apply`

## Database Collections

- `forecasts` - Generated forecasts and predictions
- `peak_loads` - Peak load predictions
- `optimization_scenarios` - Optimization scenarios and actions
- `recommendations` - Energy-saving recommendations
- `devices` - Device metadata (optional)

## Testing

```bash
# Run all tests
go test ./tests/... -v

# Run specific test
go test ./tests/... -v -run TestForecastGeneration
```

## Health Check

```bash
curl http://localhost:8082/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "forecast-service"
}
```

## Development

### Building

```bash
go build -o forecast-service ./cmd/main.go
```

### Running Tests

```bash
go test ./... -v -cover
```

## License

This project is part of the EMSIB microservices infrastructure.

