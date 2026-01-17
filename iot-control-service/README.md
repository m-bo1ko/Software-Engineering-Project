# IoT & Control Component Service

A microservice responsible for telemetry ingestion, device management, real-time device state, command execution, and applying optimization scenarios in the EMSIB system.

## Features

- **Telemetry Ingestion**: HTTP and MQTT-based telemetry ingestion
- **Device Management**: Device registration, listing, and status tracking
- **Device Control**: Command execution via MQTT
- **Optimization Scenarios**: Apply and track optimization scenarios from Forecast service
- **Real-Time State**: Live device state monitoring
- **Historical Data**: Telemetry history queries

## Technology Stack

- **Language**: Go 1.21
- **HTTP Framework**: Gin
- **Database**: MongoDB
- **Messaging**: MQTT (Eclipse Paho)
- **Authentication**: JWT via Security service

## Project Structure

```
iot-control-service/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── config/                    # Configuration
│   ├── handlers/                  # HTTP handlers
│   ├── middleware/                # Auth and common middleware
│   ├── models/                    # Domain models
│   ├── repository/                # MongoDB repositories
│   ├── service/                   # Business logic
│   ├── integrations/              # External service clients
│   └── mqtt/                      # MQTT client
├── Dockerfile
├── .env.example
└── README.md
```

## API Endpoints

### Telemetry
- `POST /iot/telemetry` - Ingest single telemetry message
- `POST /iot/telemetry/bulk` - Ingest batch telemetry
- `GET /iot/telemetry/history` - Get telemetry history

### Devices
- `GET /iot/devices` - List devices
- `GET /iot/devices/{deviceId}` - Get device details
- `POST /iot/devices/register` - Register new device

### Control
- `POST /iot/device-control/{deviceId}/command` - Send command to device
- `GET /iot/device-control/{deviceId}/commands` - List device commands

### Optimization
- `POST /iot/optimization/apply` - Apply optimization scenario
- `GET /iot/optimization/status/{scenarioId}` - Get optimization status

### State
- `GET /iot/state/live` - Get live state for all devices
- `GET /iot/state/{deviceId}` - Get device state

## MQTT Topics

- `mqtt/iot/{deviceId}/telemetry` - Device → Service (telemetry data)
- `mqtt/iot/{deviceId}/command` - Service → Device (commands)
- `mqtt/iot/{deviceId}/ack` - Device → Service (command acknowledgments)
- `mqtt/iot/broadcast/announcement` - Service → All devices (broadcasts)

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key configuration:
- `SERVER_PORT`: HTTP server port (default: 8083)
- `MONGODB_URI`: MongoDB connection string
- `MQTT_BROKER`: MQTT broker address
- `SECURITY_SERVICE_URL`: Security service URL for authentication
- `FORECAST_SERVICE_URL`: Forecast service URL
- `ANALYTICS_SERVICE_URL`: Analytics service URL

## Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Start MongoDB and MQTT broker (via docker-compose or locally)

3. Run the service:
```bash
go run cmd/main.go
```

## Docker

Build and run with Docker:

```bash
docker build -t iot-control-service .
docker run -p 8083:8083 --env-file .env iot-control-service
```

## Example Requests

### Register Device
```bash
curl -X POST http://localhost:8083/iot/devices/register \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "device-001",
    "type": "HVAC",
    "model": "Model-X",
    "location": {
      "buildingId": "building-001"
    },
    "capabilities": ["temperature", "humidity"]
  }'
```

### Ingest Telemetry
```bash
curl -X POST http://localhost:8083/iot/telemetry \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "device-001",
    "metrics": {
      "temperature": 22.5,
      "humidity": 45.0
    }
  }'
```

### Send Command
```bash
curl -X POST http://localhost:8083/iot/device-control/device-001/command \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "command": "SET_TEMPERATURE",
    "params": {
      "temperature": 20.0
    }
  }'
```

## MQTT Examples

### Publish Telemetry (Device)
```json
{
  "deviceId": "device-001",
  "timestamp": "2024-01-15T10:00:00Z",
  "metrics": {
    "temperature": 22.5,
    "humidity": 45.0
  }
}
```
Topic: `mqtt/iot/device-001/telemetry`

### Command Acknowledgment (Device)
```json
{
  "commandId": "cmd-123",
  "deviceId": "device-001",
  "status": "APPLIED",
  "timestamp": "2024-01-15T10:00:05Z"
}
```
Topic: `mqtt/iot/device-001/ack`

## Integration

This service integrates with:
- **Security Service**: Authentication and audit logging
- **Forecast Service**: Device predictions and optimization recommendations
- **Analytics Service**: Anomaly detection

## Testing

Run tests:
```bash
go test ./...
```

## License

Part of the EMSIB system.
