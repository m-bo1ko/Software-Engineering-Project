# Forecast Service

## Overview

The **Forecast Service** is responsible for energy demand forecasting, peak load prediction, optimization scenario generation, and providing recommendations for the EMSIB platform. It uses historical data, weather information, and tariff data to generate accurate predictions and actionable optimization scenarios.

**Port:** `8082`

---

## Table of Contents

- [Architecture](#architecture)
- [Data Models](#data-models)
- [API Endpoints](#api-endpoints)
- [Authentication](#authentication)
- [Configuration](#configuration)
- [Database Schema](#database-schema)
- [Inter-Service Communication](#inter-service-communication)
- [External APIs](#external-apis)
- [Running the Service](#running-the-service)
- [Health Check](#health-check)
- [Example Usage](#example-usage)
- [Known Limitations](#known-limitations)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Forecast Service                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Handlers  │  │  Middleware │  │   Router    │              │
│  │             │  │             │  │             │              │
│  │ - Forecast  │  │ - Auth      │  │ - v1 routes │              │
│  │ - Optimize  │  │ - CORS      │  │ - Legacy    │              │
│  │             │  │ - Security  │  │   routes    │              │
│  │             │  │ - Logging   │  │             │              │
│  └──────┬──────┘  └─────────────┘  └─────────────┘              │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────┐        │
│  │                    Services                          │        │
│  │          Forecast | Optimization                     │        │
│  └──────┬──────────────────────────────────────────────┘        │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────┐        │
│  │                  Repositories                        │        │
│  │   Forecast | Optimization | PeakLoad | Recommendation│        │
│  └──────┬──────────────────────────────────────────────┘        │
│         │                                                        │
│  ┌──────▼──────┐  ┌──────────────────────────────────┐          │
│  │   MongoDB   │  │       External Integrations       │          │
│  │             │  │  Weather | Tariff | ML | Storage  │          │
│  └─────────────┘  └──────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Internal Layers

| Layer | Description |
|-------|-------------|
| **Handlers** | HTTP request handlers for forecast and optimization endpoints |
| **Middleware** | Authentication, CORS, security headers, logging, recovery |
| **Services** | Business logic for forecasting and optimization |
| **Repositories** | Data access layer for MongoDB operations |
| **Models** | Domain entities (Forecast, Optimization, PeakLoad, etc.) |
| **Integrations** | Clients for Security, IoT, Weather, Tariff, ML, and Storage services |

---

## Data Models

### Forecast

```json
{
  "id": "507f1f77bcf86cd799439011",
  "buildingId": "building-001",
  "deviceId": "device-001",
  "type": "DEMAND",
  "status": "COMPLETED",
  "horizonHours": 24,
  "startTime": "2024-01-20T00:00:00Z",
  "endTime": "2024-01-21T00:00:00Z",
  "predictions": [
    {
      "timestamp": "2024-01-20T00:00:00Z",
      "predictedValue": 150.5,
      "lowerBound": 140.0,
      "upperBound": 161.0,
      "confidenceLevel": 0.95,
      "unit": "kWh"
    }
  ],
  "accuracy": {
    "mae": 5.2,
    "rmse": 7.8,
    "mape": 3.5,
    "score": 92.5
  },
  "modelUsed": "LSTM-v2",
  "createdAt": "2024-01-19T23:00:00Z",
  "createdBy": "user123"
}
```

### Forecast Types

| Type | Description |
|------|-------------|
| `DEMAND` | Energy demand forecast |
| `CONSUMPTION` | Energy consumption forecast |
| `LOAD` | Load profile forecast |

### Forecast Status

| Status | Description |
|--------|-------------|
| `PENDING` | Forecast queued for processing |
| `PROCESSING` | Forecast being generated |
| `COMPLETED` | Forecast successfully generated |
| `FAILED` | Forecast generation failed |

### Optimization Scenario

```json
{
  "id": "507f1f77bcf86cd799439012",
  "buildingId": "building-001",
  "name": "Peak Shaving Scenario",
  "description": "Reduce peak consumption during 14:00-18:00",
  "type": "PEAK_SHAVING",
  "status": "APPROVED",
  "forecastId": "507f1f77bcf86cd799439011",
  "scheduledStart": "2024-01-21T14:00:00Z",
  "scheduledEnd": "2024-01-21T18:00:00Z",
  "actions": [
    {
      "id": "action-001",
      "deviceId": "hvac-001",
      "deviceName": "Main HVAC Unit",
      "deviceType": "HVAC",
      "actionType": "SET_TEMP",
      "currentValue": "22",
      "targetValue": "24",
      "scheduledTime": "2024-01-21T14:00:00Z",
      "duration": 240,
      "status": "PENDING",
      "expectedImpact": 15.5
    }
  ],
  "expectedSavings": {
    "energyKWh": 45.5,
    "costAmount": 12.50,
    "currency": "USD",
    "co2ReductionKg": 18.2,
    "percentReduction": 12.5
  },
  "constraints": {
    "minTemperature": 20.0,
    "maxTemperature": 26.0,
    "occupancyRequired": true,
    "preserveComfort": true
  },
  "priority": 7,
  "createdAt": "2024-01-20T10:00:00Z",
  "createdBy": "user123"
}
```

### Optimization Types

| Type | Description |
|------|-------------|
| `COST_REDUCTION` | Minimize energy costs |
| `PEAK_SHAVING` | Reduce peak demand |
| `LOAD_BALANCING` | Balance load across time |
| `EFFICIENCY` | Improve energy efficiency |
| `COMFORT` | Optimize for comfort |
| `DEMAND_RESPONSE` | Respond to utility signals |

### Device Prediction

```json
{
  "deviceId": "hvac-001",
  "deviceName": "Main HVAC Unit",
  "deviceType": "HVAC",
  "currentConsumption": 5.2,
  "predictedValues": [
    {
      "timestamp": "2024-01-20T01:00:00Z",
      "predictedValue": 5.5,
      "lowerBound": 5.0,
      "upperBound": 6.0,
      "confidenceLevel": 0.90,
      "unit": "kW"
    }
  ],
  "trend": "INCREASING",
  "trendPercentage": 5.8
}
```

---

## API Endpoints

### Forecast Generation

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/forecast/generate` | Generate new forecast | Yes |
| `POST` | `/api/v1/forecast/peak-load` | Generate peak load prediction | Yes |
| `GET` | `/api/v1/forecast/latest` | Get latest forecast | Yes |
| `GET` | `/api/v1/forecast/prediction/:deviceId` | Get device prediction | Yes |
| `GET` | `/api/v1/forecast/optimization/:deviceId` | Get device optimization | Yes |

#### POST /api/v1/forecast/generate

**Request:**
```json
{
  "buildingId": "building-001",
  "deviceId": "device-001",
  "type": "DEMAND",
  "horizonHours": 24,
  "includeWeather": true,
  "includeTariffs": true,
  "historicalDays": 30,
  "metadata": {
    "requestedBy": "dashboard"
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "buildingId": "building-001",
    "type": "DEMAND",
    "status": "PENDING",
    "horizonHours": 24,
    "startTime": "2024-01-20T00:00:00Z",
    "endTime": "2024-01-21T00:00:00Z",
    "predictions": [],
    "modelUsed": "LSTM-v2",
    "createdAt": "2024-01-20T10:00:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8082/api/v1/forecast/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buildingId": "building-001",
    "type": "DEMAND",
    "horizonHours": 48,
    "includeWeather": true,
    "includeTariffs": true
  }'
```

#### POST /api/v1/forecast/peak-load

**Request:**
```json
{
  "buildingId": "building-001",
  "type": "DEMAND",
  "horizonHours": 24
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "buildingId": "building-001",
    "peakTime": "2024-01-20T15:30:00Z",
    "peakValue": 250.5,
    "unit": "kW",
    "thresholdExceeded": true,
    "thresholdValue": 200.0,
    "recommendations": [
      "Consider pre-cooling before peak hours",
      "Shift non-critical loads to off-peak times"
    ]
  }
}
```

#### GET /api/v1/forecast/latest

**Query Parameters:**
- `buildingId` (required) - Building identifier

**Example:**
```bash
curl -X GET "http://localhost:8082/api/v1/forecast/latest?buildingId=building-001" \
  -H "Authorization: Bearer $TOKEN"
```

#### GET /api/v1/forecast/prediction/:deviceId

Get consumption predictions for a specific device.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "deviceId": "hvac-001",
    "deviceName": "Main HVAC Unit",
    "deviceType": "HVAC",
    "currentConsumption": 5.2,
    "predictedValues": [...],
    "trend": "STABLE",
    "trendPercentage": 0.5
  }
}
```

---

### Optimization Scenarios

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/optimization/generate` | Generate optimization scenario | Yes |
| `GET` | `/api/v1/optimization/recommendations/:buildingId` | Get recommendations | Yes |
| `GET` | `/api/v1/optimization/scenario/:scenarioId` | Get scenario details | Yes |
| `POST` | `/api/v1/optimization/send-to-iot` | Send scenario to IoT service | Yes |

#### POST /api/v1/optimization/generate

**Request:**
```json
{
  "buildingId": "building-001",
  "name": "Evening Peak Reduction",
  "type": "PEAK_SHAVING",
  "scheduledStart": "2024-01-21T17:00:00Z",
  "scheduledEnd": "2024-01-21T21:00:00Z",
  "forecastId": "507f1f77bcf86cd799439011",
  "useTariffData": true,
  "useWeatherData": true,
  "constraints": {
    "minTemperature": 20.0,
    "maxTemperature": 26.0,
    "preserveComfort": true,
    "excludeDevices": ["critical-device-001"]
  },
  "priority": 8
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "buildingId": "building-001",
    "name": "Evening Peak Reduction",
    "type": "PEAK_SHAVING",
    "status": "DRAFT",
    "actions": [...],
    "expectedSavings": {
      "energyKWh": 35.0,
      "costAmount": 9.80,
      "currency": "USD",
      "co2ReductionKg": 14.0,
      "percentReduction": 10.5
    },
    "createdAt": "2024-01-20T10:00:00Z"
  }
}
```

#### GET /api/v1/optimization/recommendations/:buildingId

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "rec-001",
      "type": "PEAK_SHAVING",
      "title": "Pre-cooling recommendation",
      "description": "Start cooling 2 hours before peak to reduce demand",
      "potentialSavings": {
        "energyKWh": 25.0,
        "costAmount": 7.00
      },
      "priority": "HIGH",
      "applicableDevices": ["hvac-001", "hvac-002"]
    }
  ]
}
```

#### POST /api/v1/optimization/send-to-iot

Send an approved optimization scenario to the IoT Control Service for execution.

**Request:**
```json
{
  "scenarioId": "507f1f77bcf86cd799439012",
  "executeNow": false,
  "dryRun": false
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "scenarioId": "507f1f77bcf86cd799439012",
  "actionsQueued": 5,
  "actionsSkipped": 1,
  "errors": [],
  "executionId": "exec-001"
}
```

**Example:**
```bash
curl -X POST http://localhost:8082/api/v1/optimization/send-to-iot \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "scenarioId": "507f1f77bcf86cd799439012",
    "executeNow": true
  }'
```

---

## Authentication

All endpoints (except health check) require JWT authentication. The service validates tokens with the Security Service.

**Header Format:**
```
Authorization: Bearer <access_token>
```

**Token Validation Flow:**
1. Request received with Authorization header
2. Service calls Security Service `/api/v1/auth/validate-token`
3. If valid, request proceeds; if invalid, returns 401 Unauthorized

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8082` |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `GIN_MODE` | Gin framework mode | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `forecast_service` |
| `MONGODB_TIMEOUT` | Connection timeout (seconds) | `10` |
| `SECURITY_SERVICE_URL` | Security service URL | `http://localhost:8080` |
| `SECURITY_SERVICE_TIMEOUT` | Security service timeout (seconds) | `10` |
| `IOT_SERVICE_URL` | IoT Control service URL | `http://localhost:8083` |
| `IOT_SERVICE_TIMEOUT` | IoT service timeout (seconds) | `10` |
| `WEATHER_API_URL` | Weather API URL | `http://localhost:8084/external/weather` |
| `TARIFF_API_URL` | Tariff API URL | `http://localhost:8084/external/tariffs` |
| `ML_MODEL_URL` | ML model service URL | `http://localhost:8085/ml/predict` |
| `STORAGE_API_URL` | Storage service URL | `http://localhost:8086/storage` |
| `FORECAST_DEFAULT_HORIZON_HOURS` | Default forecast horizon | `24` |
| `FORECAST_MAX_HORIZON_HOURS` | Maximum forecast horizon | `168` (7 days) |
| `PEAK_LOAD_THRESHOLD_PERCENTAGE` | Peak load warning threshold | `80.0` |
| `LOG_LEVEL` | Logging level | `debug` |
| `LOG_FORMAT` | Log format | `json` |

### Example .env File

```env
SERVER_PORT=8082
GIN_MODE=release

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=forecast_service

SECURITY_SERVICE_URL=http://security-service:8080
IOT_SERVICE_URL=http://iot-control-service:8083

WEATHER_API_URL=https://api.weather.com/v1
TARIFF_API_URL=https://api.utility.com/tariffs
ML_MODEL_URL=http://ml-service:8085/predict

FORECAST_DEFAULT_HORIZON_HOURS=24
FORECAST_MAX_HORIZON_HOURS=168
PEAK_LOAD_THRESHOLD_PERCENTAGE=80

LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Database Schema

### MongoDB Collections

| Collection | Description | Indexes |
|------------|-------------|---------|
| `forecasts` | Generated forecasts | `building_id`, `created_at`, `status` |
| `peak_loads` | Peak load predictions | `building_id`, `peak_time` |
| `optimization_scenarios` | Optimization scenarios | `building_id`, `status`, `scheduled_start` |
| `recommendations` | Generated recommendations | `building_id`, `priority` |

---

## Inter-Service Communication

### Inbound (Other services calling Forecast Service)

| Caller | Endpoint | Purpose |
|--------|----------|---------|
| Analytics Service | `GET /api/v1/forecast/latest` | Get forecast for dashboard |
| IoT Control Service | `GET /api/v1/forecast/prediction/:deviceId` | Get device predictions for optimization |

### Outbound (Forecast Service calling other services)

| Target | Endpoint | Purpose |
|--------|----------|---------|
| Security Service | `GET /api/v1/auth/validate-token` | Validate JWT tokens |
| IoT Control Service | `POST /api/v1/iot/optimization/applySecurity` | Apply optimization scenarios |
| Weather API | External | Get weather data for forecasting |
| Tariff API | External | Get tariff data for cost optimization |
| ML Service | External | Get predictions from ML models |

---

## External APIs

### Weather API Integration

The service fetches weather data to improve forecast accuracy.

**Expected Response Format:**
```json
{
  "temperature": 25.5,
  "humidity": 60.0,
  "cloudCover": 30.0,
  "windSpeed": 10.5,
  "condition": "partly_cloudy",
  "forecastedHigh": 28.0,
  "forecastedLow": 18.0
}
```

### Tariff API Integration

The service fetches tariff data for cost optimization.

**Expected Response Format:**
```json
{
  "region": "US-CA",
  "currentRate": 0.15,
  "peakRate": 0.25,
  "offPeakRate": 0.10,
  "currency": "USD",
  "timeOfUseRates": [
    {
      "name": "Off-Peak",
      "ratePerKWh": 0.10,
      "startHour": 0,
      "endHour": 7
    },
    {
      "name": "Peak",
      "ratePerKWh": 0.25,
      "startHour": 14,
      "endHour": 20
    }
  ]
}
```

### ML Model Service Integration

The service calls ML models for predictions.

**Request:**
```json
{
  "buildingId": "building-001",
  "historicalData": [...],
  "weatherData": {...},
  "horizonHours": 24
}
```

**Expected Response:**
```json
{
  "predictions": [
    {
      "timestamp": "2024-01-20T00:00:00Z",
      "value": 150.5,
      "confidence": 0.95
    }
  ],
  "modelVersion": "LSTM-v2"
}
```

---

## Running the Service

### Prerequisites

- Go 1.21+
- MongoDB 7.0+
- Docker (optional)

### Local Development

```bash
# Navigate to service directory
cd forecast-service

# Install dependencies
go mod download

# Set environment variables
export MONGODB_URI=mongodb://localhost:27017
export SECURITY_SERVICE_URL=http://localhost:8080

# Run the service
go run cmd/main.go
```

### Using Docker

```bash
# Build the image
docker build -t forecast-service ./forecast-service

# Run the container
docker run -p 8082:8082 \
  -e MONGODB_URI=mongodb://host.docker.internal:27017 \
  -e SECURITY_SERVICE_URL=http://security-service:8080 \
  forecast-service
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d forecast-service

# View logs
docker-compose logs -f forecast-service
```

---

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

---

## Example Usage

### Complete Forecasting Workflow

```bash
# 1. Generate a 48-hour demand forecast
FORECAST_RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/forecast/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buildingId": "building-001",
    "type": "DEMAND",
    "horizonHours": 48,
    "includeWeather": true,
    "includeTariffs": true
  }')

FORECAST_ID=$(echo $FORECAST_RESPONSE | jq -r '.data.id')

# 2. Wait for forecast to complete, then get latest
sleep 5
curl -X GET "http://localhost:8082/api/v1/forecast/latest?buildingId=building-001" \
  -H "Authorization: Bearer $TOKEN"

# 3. Check for peak load
curl -X POST http://localhost:8082/api/v1/forecast/peak-load \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"buildingId": "building-001", "horizonHours": 24}'

# 4. Generate optimization scenario based on forecast
curl -X POST http://localhost:8082/api/v1/optimization/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"buildingId\": \"building-001\",
    \"name\": \"Peak Reduction\",
    \"type\": \"PEAK_SHAVING\",
    \"forecastId\": \"$FORECAST_ID\",
    \"useTariffData\": true,
    \"priority\": 8
  }"

# 5. Send approved scenario to IoT service
curl -X POST http://localhost:8082/api/v1/optimization/send-to-iot \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"scenarioId": "scenario-id", "executeNow": true}'
```

### Get Device-Level Predictions

```bash
# Get prediction for specific device
curl -X GET http://localhost:8082/api/v1/forecast/prediction/hvac-001 \
  -H "Authorization: Bearer $TOKEN"

# Get optimization recommendations for device
curl -X GET http://localhost:8082/api/v1/forecast/optimization/hvac-001 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Known Limitations

1. **External API Dependencies:** Weather and tariff APIs must be available for full functionality
2. **ML Model Required:** Advanced predictions require external ML service
3. **Forecast Horizon:** Maximum 168 hours (7 days) forecast horizon
4. **Historical Data:** Requires sufficient historical data for accurate predictions
5. **No Real-time Updates:** Forecasts are point-in-time; no streaming updates
6. **Single Building Focus:** Optimization scenarios are building-specific

---

## Developer Notes

- Forecast generation is asynchronous; poll for completion or use webhooks
- Peak load threshold is configurable per deployment
- Weather data significantly improves forecast accuracy
- Optimization scenarios should be reviewed before sending to IoT service
- Use `dryRun: true` to test optimization scenarios without execution
- Historical data quality directly impacts forecast accuracy
