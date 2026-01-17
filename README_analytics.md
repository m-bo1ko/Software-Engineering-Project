# Analytics Service

## Overview

The **Analytics Service** processes telemetry data, calculates Key Performance Indicators (KPIs), generates reports, detects anomalies, and provides dashboard data for the EMSIB platform. It aggregates data from IoT devices and enriches it with forecast information to provide comprehensive insights.

**Port:** `8084`

---

## Table of Contents

- [Architecture](#architecture)
- [Data Models](#data-models)
- [API Endpoints](#api-endpoints)
- [Authentication](#authentication)
- [Configuration](#configuration)
- [Database Schema](#database-schema)
- [Inter-Service Communication](#inter-service-communication)
- [Running the Service](#running-the-service)
- [Health Check](#health-check)
- [Example Usage](#example-usage)
- [Known Limitations](#known-limitations)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Analytics Service                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Handlers   │  │  Middleware │  │   Router    │             │
│  │              │  │             │  │             │             │
│  │ - Report     │  │ - Auth      │  │ - v1 routes │             │
│  │ - Anomaly    │  │ - CORS      │  │ - Legacy    │             │
│  │ - TimeSeries │  │ - Security  │  │   routes    │             │
│  │ - KPI        │  │ - Logging   │  │             │             │
│  │ - Dashboard  │  │ - Recovery  │  │             │             │
│  └──────┬───────┘  └─────────────┘  └─────────────┘             │
│         │                                                        │
│  ┌──────▼───────────────────────────────────────────────┐       │
│  │                      Services                         │       │
│  │  Report | Anomaly | TimeSeries | KPI | Dashboard      │       │
│  └──────┬───────────────────────────────────────────────┘       │
│         │                                                        │
│  ┌──────▼───────────────────────────────────────────────┐       │
│  │                    Repositories                       │       │
│  │      Report | Anomaly | TimeSeries | KPI              │       │
│  └──────┬───────────────────────────────────────────────┘       │
│         │                                                        │
│  ┌──────▼──────┐  ┌──────────────────────────────────┐          │
│  │   MongoDB   │  │        External Integrations      │          │
│  │             │  │   Security | IoT | Forecast       │          │
│  └─────────────┘  └──────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Internal Layers

| Layer | Description |
|-------|-------------|
| **Handlers** | HTTP request handlers for reports, anomalies, time-series, KPIs, and dashboards |
| **Middleware** | Authentication, CORS, security headers, logging, recovery |
| **Services** | Business logic for analytics, anomaly detection, KPI calculation |
| **Repositories** | Data access layer for MongoDB operations |
| **Models** | Domain entities (Report, Anomaly, TimeSeries, KPI, Dashboard) |
| **Integrations** | Clients for Security, IoT, Forecast, and Storage services |

---

## Data Models

### Report

```json
{
  "id": "507f1f77bcf86cd799439011",
  "reportId": "rpt-001",
  "buildingId": "building-001",
  "type": "ENERGY_CONSUMPTION",
  "status": "COMPLETED",
  "content": {
    "summary": {
      "totalConsumption": 15000.5,
      "peakDemand": 250.0,
      "averageDaily": 500.0
    },
    "breakdown": {
      "hvac": 60.5,
      "lighting": 25.0,
      "equipment": 14.5
    },
    "recommendations": [
      "Consider upgrading HVAC systems",
      "Implement smart lighting controls"
    ]
  },
  "generatedAt": "2024-01-20T10:00:00Z",
  "generatedBy": "user123",
  "createdAt": "2024-01-20T09:55:00Z"
}
```

### Report Types

| Type | Description |
|------|-------------|
| `ENERGY_CONSUMPTION` | Energy consumption analysis |
| `COST_ANALYSIS` | Energy cost breakdown |
| `EFFICIENCY` | Efficiency metrics |
| `SUSTAINABILITY` | Carbon footprint and sustainability |
| `COMPARISON` | Period-over-period comparison |
| `CUSTOM` | Custom report type |

### Report Status

| Status | Description |
|--------|-------------|
| `PENDING` | Report queued for generation |
| `GENERATING` | Report being generated |
| `COMPLETED` | Report successfully generated |
| `FAILED` | Report generation failed |

### Anomaly

```json
{
  "id": "507f1f77bcf86cd799439012",
  "anomalyId": "anom-001",
  "deviceId": "hvac-001",
  "buildingId": "building-001",
  "type": "CONSUMPTION_SPIKE",
  "severity": "HIGH",
  "status": "NEW",
  "details": {
    "expectedValue": 5.0,
    "actualValue": 15.0,
    "deviation": 200.0,
    "metric": "power_consumption",
    "unit": "kW"
  },
  "detectedAt": "2024-01-20T10:30:00Z",
  "createdAt": "2024-01-20T10:30:00Z"
}
```

### Anomaly Severity

| Severity | Description |
|----------|-------------|
| `LOW` | Minor deviation, informational |
| `MEDIUM` | Noticeable deviation, monitor |
| `HIGH` | Significant deviation, investigate |
| `CRITICAL` | Major deviation, immediate action |

### Anomaly Status

| Status | Description |
|--------|-------------|
| `NEW` | Newly detected anomaly |
| `ACKNOWLEDGED` | Anomaly has been reviewed |
| `RESOLVED` | Anomaly has been resolved |
| `FALSE_POSITIVE` | Marked as false positive |

### KPI

```json
{
  "id": "507f1f77bcf86cd799439013",
  "buildingId": "building-001",
  "calculatedAt": "2024-01-20T00:00:00Z",
  "metrics": {
    "energyIntensity": 150.5,
    "peakDemandRatio": 0.75,
    "equipmentEfficiency": 0.92,
    "renewableShare": 0.15,
    "carbonIntensity": 0.45,
    "costPerSquareMeter": 12.50
  },
  "period": "DAILY",
  "createdAt": "2024-01-20T00:05:00Z"
}
```

### KPI Periods

| Period | Description |
|--------|-------------|
| `DAILY` | Daily KPI calculation |
| `WEEKLY` | Weekly KPI calculation |
| `MONTHLY` | Monthly KPI calculation |

### Time Series

```json
{
  "id": "507f1f77bcf86cd799439014",
  "deviceId": "hvac-001",
  "buildingId": "building-001",
  "metricName": "power_consumption",
  "aggregation": "AVG",
  "interval": "1h",
  "dataPoints": [
    {"timestamp": "2024-01-20T00:00:00Z", "value": 5.2},
    {"timestamp": "2024-01-20T01:00:00Z", "value": 4.8},
    {"timestamp": "2024-01-20T02:00:00Z", "value": 4.5}
  ],
  "unit": "kW"
}
```

### Dashboard Overview

```json
{
  "totalDevices": 150,
  "onlineDevices": 142,
  "totalBuildings": 5,
  "activeAnomalies": 3,
  "kpis": {
    "totalConsumption": 15000.5,
    "averageEfficiency": 0.89,
    "costToDate": 4500.00
  },
  "recentAnomalies": [...],
  "updatedAt": "2024-01-20T10:35:00Z"
}
```

### Building Dashboard

```json
{
  "buildingId": "building-001",
  "deviceCount": 45,
  "onlineDeviceCount": 43,
  "activeAnomalies": 1,
  "kpis": {
    "energyIntensity": 150.5,
    "efficiency": 0.92,
    "forecast_available": true,
    "forecast_horizon_hours": 24,
    "forecast_confidence": 0.95
  },
  "forecastSummary": {
    "nextPeakTime": "2024-01-20T15:00:00Z",
    "nextPeakValue": 250.0,
    "predictions": [...]
  },
  "recentTelemetry": [...],
  "updatedAt": "2024-01-20T10:35:00Z"
}
```

---

## API Endpoints

### Reports

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/analytics/reports` | List reports | Yes |
| `GET` | `/api/v1/analytics/reports/:reportId` | Get report by ID | Yes |
| `POST` | `/api/v1/analytics/reports/generate` | Generate new report | Yes |

#### POST /api/v1/analytics/reports/generate

**Request:**
```json
{
  "buildingId": "building-001",
  "type": "ENERGY_CONSUMPTION",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z",
  "options": {
    "includeBreakdown": true,
    "includeRecommendations": true,
    "compareWithPreviousPeriod": true
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "reportId": "rpt-002",
    "buildingId": "building-001",
    "type": "ENERGY_CONSUMPTION",
    "status": "PENDING",
    "createdAt": "2024-01-20T10:00:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8084/api/v1/analytics/reports/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buildingId": "building-001",
    "type": "ENERGY_CONSUMPTION",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z"
  }'
```

#### GET /api/v1/analytics/reports

**Query Parameters:**
- `buildingId` - Filter by building
- `type` - Filter by report type
- `status` - Filter by status
- `page` - Page number
- `limit` - Items per page

**Example:**
```bash
curl -X GET "http://localhost:8084/api/v1/analytics/reports?buildingId=building-001&type=ENERGY_CONSUMPTION&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

---

### Anomalies

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/analytics/anomalies` | List anomalies | Yes |
| `GET` | `/api/v1/analytics/anomalies/:anomalyId` | Get anomaly by ID | Yes |
| `POST` | `/api/v1/analytics/anomalies/acknowledge` | Acknowledge anomaly | Yes |

#### GET /api/v1/analytics/anomalies

**Query Parameters:**
- `deviceId` - Filter by device
- `buildingId` - Filter by building
- `type` - Filter by anomaly type
- `severity` - Filter by severity (LOW, MEDIUM, HIGH, CRITICAL)
- `status` - Filter by status (NEW, ACKNOWLEDGED, RESOLVED, FALSE_POSITIVE)
- `page` - Page number
- `limit` - Items per page

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "anomalies": [
      {
        "id": "507f1f77bcf86cd799439012",
        "anomalyId": "anom-001",
        "deviceId": "hvac-001",
        "buildingId": "building-001",
        "type": "CONSUMPTION_SPIKE",
        "severity": "HIGH",
        "status": "NEW",
        "details": {...},
        "detectedAt": "2024-01-20T10:30:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "limit": 10
  }
}
```

**Example:**
```bash
curl -X GET "http://localhost:8084/api/v1/analytics/anomalies?severity=HIGH&status=NEW" \
  -H "Authorization: Bearer $TOKEN"
```

#### POST /api/v1/analytics/anomalies/acknowledge

**Request:**
```json
{
  "anomalyId": "anom-001"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "anomalyId": "anom-001",
    "status": "ACKNOWLEDGED",
    "acknowledgedAt": "2024-01-20T11:00:00Z",
    "acknowledgedBy": "user123"
  }
}
```

---

### Time Series

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/analytics/time-series/query` | Query time series data | Yes |

#### POST /api/v1/analytics/time-series/query

**Request:**
```json
{
  "deviceId": "hvac-001",
  "buildingId": "building-001",
  "metricName": "power_consumption",
  "aggregation": "AVG",
  "interval": "1h",
  "from": "2024-01-20T00:00:00Z",
  "to": "2024-01-20T23:59:59Z"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "deviceId": "hvac-001",
    "metricName": "power_consumption",
    "aggregation": "AVG",
    "interval": "1h",
    "dataPoints": [
      {"timestamp": "2024-01-20T00:00:00Z", "value": 5.2},
      {"timestamp": "2024-01-20T01:00:00Z", "value": 4.8},
      {"timestamp": "2024-01-20T02:00:00Z", "value": 4.5}
    ],
    "unit": "kW",
    "count": 24
  }
}
```

**Aggregation Types:**
- `AVG` - Average value
- `SUM` - Sum of values
- `MIN` - Minimum value
- `MAX` - Maximum value
- `COUNT` - Count of data points

**Example:**
```bash
curl -X POST http://localhost:8084/api/v1/analytics/time-series/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "deviceId": "hvac-001",
    "metricName": "power_consumption",
    "aggregation": "AVG",
    "interval": "1h",
    "from": "2024-01-20T00:00:00Z",
    "to": "2024-01-20T23:59:59Z"
  }'
```

---

### KPIs

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/analytics/kpi` | Get system-wide KPIs | Yes |
| `GET` | `/api/v1/analytics/kpi/:buildingId` | Get building KPIs | Yes |
| `POST` | `/api/v1/analytics/kpi/calculate` | Trigger KPI calculation | Yes |

#### GET /api/v1/analytics/kpi/:buildingId

**Query Parameters:**
- `period` - KPI period (DAILY, WEEKLY, MONTHLY)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "buildingId": "building-001",
    "calculatedAt": "2024-01-20T00:00:00Z",
    "metrics": {
      "energyIntensity": 150.5,
      "peakDemandRatio": 0.75,
      "equipmentEfficiency": 0.92,
      "renewableShare": 0.15,
      "carbonIntensity": 0.45,
      "costPerSquareMeter": 12.50
    },
    "period": "DAILY"
  }
}
```

#### POST /api/v1/analytics/kpi/calculate

**Request:**
```json
{
  "buildingId": "building-001",
  "period": "DAILY",
  "force": true
}
```

---

### Dashboards

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/analytics/dashboards/overview` | Get system overview | Yes |
| `GET` | `/api/v1/analytics/dashboards/building/:buildingId` | Get building dashboard | Yes |

#### GET /api/v1/analytics/dashboards/overview

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "totalDevices": 150,
    "onlineDevices": 142,
    "totalBuildings": 5,
    "activeAnomalies": 3,
    "kpis": {
      "totalConsumption": 15000.5,
      "averageEfficiency": 0.89,
      "costToDate": 4500.00
    },
    "recentAnomalies": [...],
    "updatedAt": "2024-01-20T10:35:00Z"
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8084/api/v1/analytics/dashboards/overview \
  -H "Authorization: Bearer $TOKEN"
```

#### GET /api/v1/analytics/dashboards/building/:buildingId

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "buildingId": "building-001",
    "deviceCount": 45,
    "onlineDeviceCount": 43,
    "activeAnomalies": 1,
    "kpis": {
      "energyIntensity": 150.5,
      "efficiency": 0.92,
      "forecast_available": true,
      "forecast_horizon_hours": 24,
      "forecast_confidence": 0.95
    },
    "forecastSummary": {
      "nextPeakTime": "2024-01-20T15:00:00Z",
      "nextPeakValue": 250.0,
      "predictions": [...]
    },
    "recentTelemetry": [...],
    "updatedAt": "2024-01-20T10:35:00Z"
  }
}
```

**Integration Notes:**
- Dashboard fetches device data from IoT Service
- Dashboard fetches forecast data from Forecast Service
- KPIs are enriched with forecast confidence scores

---

## Authentication

All endpoints (except health check) require JWT authentication. The service validates tokens with the Security Service.

**Header Format:**
```
Authorization: Bearer <access_token>
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8084` |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `GIN_MODE` | Gin framework mode | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `analytics_service` |
| `MONGODB_TIMEOUT` | Connection timeout (seconds) | `10` |
| `SECURITY_SERVICE_URL` | Security service URL | `http://localhost:8080` |
| `SECURITY_SERVICE_TIMEOUT` | Security service timeout (seconds) | `10` |
| `IOT_SERVICE_URL` | IoT service URL | `http://localhost:8083` |
| `IOT_SERVICE_TIMEOUT` | IoT service timeout (seconds) | `10` |
| `FORECAST_SERVICE_URL` | Forecast service URL | `http://localhost:8082` |
| `FORECAST_SERVICE_TIMEOUT` | Forecast service timeout (seconds) | `10` |
| `STORAGE_SERVICE_URL` | Storage service URL | `http://localhost:8086/storage` |
| `STORAGE_SERVICE_TIMEOUT` | Storage service timeout (seconds) | `10` |
| `ANALYTICS_ANOMALY_DETECTION_ENABLED` | Enable anomaly detection | `true` |
| `ANALYTICS_KPI_CALCULATION_INTERVAL` | KPI calculation interval (minutes) | `60` |
| `ANALYTICS_REPORT_RETENTION_DAYS` | Report retention period (days) | `90` |
| `ANALYTICS_TIME_SERIES_AGGREGATION_INTERVAL` | Time-series aggregation (minutes) | `60` |
| `LOG_LEVEL` | Logging level | `debug` |
| `LOG_FORMAT` | Log format | `json` |

### Example .env File

```env
SERVER_PORT=8084
GIN_MODE=release

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=analytics_service

SECURITY_SERVICE_URL=http://security-service:8080
IOT_SERVICE_URL=http://iot-control-service:8083
FORECAST_SERVICE_URL=http://forecast-service:8082

ANALYTICS_ANOMALY_DETECTION_ENABLED=true
ANALYTICS_KPI_CALCULATION_INTERVAL=60
ANALYTICS_REPORT_RETENTION_DAYS=90

LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Database Schema

### MongoDB Collections

| Collection | Description | Indexes |
|------------|-------------|---------|
| `reports` | Generated reports | `report_id`, `building_id`, `type`, `created_at` |
| `anomalies` | Detected anomalies | `anomaly_id`, `device_id`, `building_id`, `severity`, `status` |
| `time_series` | Aggregated time-series data | `device_id`, `metric_name`, `timestamp` |
| `kpis` | Calculated KPIs | `building_id`, `period`, `calculated_at` |

---

## Inter-Service Communication

### Inbound (Other services calling Analytics Service)

| Caller | Endpoint | Purpose |
|--------|----------|---------|
| IoT Control Service | `GET /api/v1/analytics/anomalies` | Check for device anomalies before optimization |

### Outbound (Analytics Service calling other services)

| Target | Endpoint | Purpose |
|--------|----------|---------|
| Security Service | `GET /api/v1/auth/validate-token` | Validate JWT tokens |
| IoT Service | `GET /api/v1/iot/devices` | Get device list for dashboards |
| IoT Service | `GET /api/v1/iot/telemetry/history` | Get telemetry for analysis |
| Forecast Service | `GET /api/v1/forecast/latest` | Get forecast data for dashboards |

---

## Running the Service

### Prerequisites

- Go 1.21+
- MongoDB 7.0+
- Docker (optional)

### Local Development

```bash
# Navigate to service directory
cd analytics-service

# Install dependencies
go mod download

# Set environment variables
export MONGODB_URI=mongodb://localhost:27017
export SECURITY_SERVICE_URL=http://localhost:8080
export IOT_SERVICE_URL=http://localhost:8083

# Run the service
go run cmd/main.go
```

### Using Docker

```bash
# Build the image
docker build -t analytics-service ./analytics-service

# Run the container
docker run -p 8084:8084 \
  -e MONGODB_URI=mongodb://host.docker.internal:27017 \
  -e SECURITY_SERVICE_URL=http://security-service:8080 \
  -e IOT_SERVICE_URL=http://iot-control-service:8083 \
  analytics-service
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d analytics-service

# View logs
docker-compose logs -f analytics-service
```

---

## Health Check

```bash
curl http://localhost:8084/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "analytics-service"
}
```

---

## Example Usage

### Complete Analytics Workflow

```bash
# 1. Generate an energy consumption report
REPORT_RESPONSE=$(curl -s -X POST http://localhost:8084/api/v1/analytics/reports/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buildingId": "building-001",
    "type": "ENERGY_CONSUMPTION",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z",
    "options": {"includeBreakdown": true}
  }')

REPORT_ID=$(echo $REPORT_RESPONSE | jq -r '.data.reportId')

# 2. Wait for report generation, then retrieve
sleep 5
curl -X GET "http://localhost:8084/api/v1/analytics/reports/$REPORT_ID" \
  -H "Authorization: Bearer $TOKEN"

# 3. Check for anomalies
curl -X GET "http://localhost:8084/api/v1/analytics/anomalies?buildingId=building-001&severity=HIGH" \
  -H "Authorization: Bearer $TOKEN"

# 4. Query time-series data
curl -X POST http://localhost:8084/api/v1/analytics/time-series/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buildingId": "building-001",
    "metricName": "power_consumption",
    "aggregation": "AVG",
    "interval": "1h",
    "from": "2024-01-20T00:00:00Z",
    "to": "2024-01-20T23:59:59Z"
  }'

# 5. Get building dashboard with forecast data
curl -X GET http://localhost:8084/api/v1/analytics/dashboards/building/building-001 \
  -H "Authorization: Bearer $TOKEN"
```

### Monitor and Acknowledge Anomalies

```bash
# List new critical anomalies
curl -X GET "http://localhost:8084/api/v1/analytics/anomalies?status=NEW&severity=CRITICAL" \
  -H "Authorization: Bearer $TOKEN"

# Acknowledge an anomaly
curl -X POST http://localhost:8084/api/v1/analytics/anomalies/acknowledge \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"anomalyId": "anom-001"}'
```

### Get System-Wide Overview

```bash
# Get overview dashboard
curl -X GET http://localhost:8084/api/v1/analytics/dashboards/overview \
  -H "Authorization: Bearer $TOKEN"

# Get system-wide KPIs
curl -X GET http://localhost:8084/api/v1/analytics/kpi \
  -H "Authorization: Bearer $TOKEN"

# Trigger KPI recalculation
curl -X POST http://localhost:8084/api/v1/analytics/kpi/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"buildingId": "building-001", "period": "DAILY", "force": true}'
```

---

## Known Limitations

1. **Report Generation:** Reports are generated asynchronously; poll for completion
2. **Anomaly Detection:** Basic threshold-based detection; no ML-based detection
3. **Time Series:** Limited aggregation intervals (minimum 1 minute)
4. **Data Retention:** Reports retained for 90 days by default
5. **Real-time Updates:** Dashboards are point-in-time; no WebSocket streaming
6. **Cross-Building Analysis:** Limited support for multi-building comparisons

---

## Developer Notes

- Anomaly detection runs automatically when telemetry is ingested
- KPI calculation runs on configurable intervals (default: hourly)
- Dashboard data is aggregated from multiple services for comprehensive view
- Forecast integration enriches building dashboards with prediction data
- Report generation is CPU-intensive; consider background workers for high volume
- Time-series queries support flexible aggregation intervals
- Use appropriate indexes for time-series queries performance
