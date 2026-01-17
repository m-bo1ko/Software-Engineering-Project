# Analytics Component Service

A microservice responsible for processing telemetry data, calculating KPIs, generating reports, detecting anomalies, and exposing aggregated metrics for dashboards in the EMSIB system.

## Features

- **Report Generation**: Generate analytical reports (energy consumption, device performance, anomaly summary)
- **Anomaly Detection**: Detect and track anomalies in telemetry data
- **Time-Series Analysis**: Query and aggregate historical telemetry data
- **KPI Calculation**: Calculate and track Key Performance Indicators
- **Dashboard Metrics**: System-wide and building-specific dashboard data

## Technology Stack

- **Language**: Go 1.21
- **HTTP Framework**: Gin
- **Database**: MongoDB
- **Authentication**: JWT via Security service

## Project Structure

```
analytics-service/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── config/                    # Configuration
│   ├── handlers/                  # HTTP handlers
│   ├── middleware/                # Auth and common middleware
│   ├── models/                    # Domain models
│   ├── repository/                # MongoDB repositories
│   ├── service/                   # Business logic
│   └── integrations/              # External service clients
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## API Endpoints

### Reports
- `GET /analytics/reports` - List available reports
- `GET /analytics/reports/{reportId}` - Get full report content
- `POST /analytics/reports/generate` - Trigger report generation

### Anomalies
- `GET /analytics/anomalies` - List all detected anomalies
- `GET /analytics/anomalies/{anomalyId}` - Get detailed anomaly info
- `POST /analytics/anomalies/acknowledge` - Mark anomalies as acknowledged

### Time-Series
- `POST /analytics/time-series/query` - Aggregate historical telemetry data

### KPIs
- `GET /analytics/kpi` - System-wide KPIs
- `GET /analytics/kpi/{buildingId}` - KPIs for a specific building
- `POST /analytics/kpi/calculate` - Calculate KPIs

### Dashboards
- `GET /analytics/dashboards/overview` - System-wide aggregated metrics
- `GET /analytics/dashboards/building/{buildingId}` - Building-specific metrics

## Configuration

Key configuration (via environment variables):
- `SERVER_PORT`: HTTP server port (default: 8084)
- `MONGODB_URI`: MongoDB connection string
- `SECURITY_SERVICE_URL`: Security service URL for authentication
- `IOT_SERVICE_URL`: IoT service URL
- `FORECAST_SERVICE_URL`: Forecast service URL

## Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Start MongoDB (via docker-compose or locally)

3. Run the service:
```bash
go run cmd/main.go
```

## Docker

Build and run with Docker:

```bash
docker build -t analytics-service .
docker run -p 8084:8084 --env-file .env analytics-service
```

## Example Requests

### Generate Report
```bash
curl -X POST http://localhost:8084/analytics/reports/generate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "ENERGY_CONSUMPTION",
    "buildingId": "building-001",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z"
  }'
```

### List Anomalies
```bash
curl -X GET "http://localhost:8084/analytics/anomalies?status=NEW&severity=HIGH" \
  -H "Authorization: Bearer <token>"
```

### Query Time-Series
```bash
curl -X POST http://localhost:8084/analytics/time-series/query \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceIds": ["device-001"],
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z",
    "aggregationType": "DAILY"
  }'
```

### Get KPIs
```bash
curl -X GET "http://localhost:8084/analytics/kpi/building-001?period=DAILY" \
  -H "Authorization: Bearer <token>"
```

### Get Dashboard Overview
```bash
curl -X GET http://localhost:8084/analytics/dashboards/overview \
  -H "Authorization: Bearer <token>"
```

## Integration

This service integrates with:
- **Security Service**: Authentication and audit logging
- **IoT Service**: Telemetry data and device information
- **Forecast Service**: Forecast data for dashboards

## Testing

Run tests:
```bash
go test ./...
```

## License

Part of the EMSIB system.
