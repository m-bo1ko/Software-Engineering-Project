# IoT Control Service

## Overview

The **IoT Control Service** is the core microservice for managing IoT devices, ingesting telemetry data, executing device commands, and applying optimization scenarios in the EMSIB platform. It supports both HTTP-based and MQTT-based communication with IoT devices.

**Port:** `8083`

---

## Table of Contents

- [Architecture](#architecture)
- [Data Models](#data-models)
- [API Endpoints](#api-endpoints)
- [MQTT Communication](#mqtt-communication)
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
┌─────────────────────────────────────────────────────────────────────┐
│                       IoT Control Service                            │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │   Handlers   │  │  Middleware │  │   Router    │                 │
│  │              │  │             │  │             │                 │
│  │ - Device     │  │ - Auth      │  │ - v1 routes │                 │
│  │ - Telemetry  │  │ - CORS      │  │ - Legacy    │                 │
│  │ - Control    │  │ - Security  │  │   routes    │                 │
│  │ - Optimize   │  │ - Logging   │  │             │                 │
│  │ - State      │  │ - Recovery  │  │             │                 │
│  └──────┬───────┘  └─────────────┘  └─────────────┘                 │
│         │                                                            │
│  ┌──────▼───────────────────────────────────────────────┐           │
│  │                      Services                         │           │
│  │  Device | Telemetry | Control | Optimization | State  │           │
│  └──────┬───────────────────────────────────────────────┘           │
│         │                                                            │
│  ┌──────▼───────────────────────────────────────────────┐           │
│  │                    Repositories                       │           │
│  │      Device | Telemetry | Command | Optimization      │           │
│  └──────┬───────────────────────────────────────────────┘           │
│         │                                                            │
│  ┌──────▼──────┐  ┌──────────────────────────────────────┐          │
│  │   MongoDB   │  │           MQTT Client                 │          │
│  │             │  │  - Subscribe to telemetry             │          │
│  │             │  │  - Subscribe to acks                  │          │
│  │             │  │  - Publish commands                   │          │
│  └─────────────┘  └──────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────────┘
```

### Internal Layers

| Layer | Description |
|-------|-------------|
| **Handlers** | HTTP request handlers for device, telemetry, control, optimization, and state endpoints |
| **Middleware** | Authentication, CORS, security headers, logging, recovery |
| **Services** | Business logic for device management, telemetry processing, command execution |
| **Repositories** | Data access layer for MongoDB operations |
| **Models** | Domain entities (Device, Telemetry, Command, Optimization) |
| **MQTT** | MQTT client for real-time device communication |
| **Integrations** | Clients for Security, Forecast, Analytics, and Storage services |

---

## Data Models

### Device

```json
{
  "id": "507f1f77bcf86cd799439011",
  "deviceId": "hvac-001",
  "type": "HVAC",
  "model": "Carrier 24ACC636A003",
  "location": {
    "buildingId": "building-001",
    "floor": "3",
    "room": "Conference Room A",
    "latitude": 37.7749,
    "longitude": -122.4194
  },
  "capabilities": ["SET_TEMPERATURE", "SET_MODE", "TURN_OFF"],
  "status": "ONLINE",
  "lastSeen": "2024-01-20T10:30:00Z",
  "metadata": {
    "firmwareVersion": "2.1.3",
    "manufacturer": "Carrier"
  },
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-20T10:30:00Z"
}
```

### Device Status

| Status | Description |
|--------|-------------|
| `ONLINE` | Device is connected and responsive |
| `OFFLINE` | Device is not responding |
| `ERROR` | Device has reported an error |
| `MAINTENANCE` | Device is under maintenance |

### Telemetry

```json
{
  "id": "507f1f77bcf86cd799439012",
  "deviceId": "hvac-001",
  "timestamp": "2024-01-20T10:30:00Z",
  "metrics": {
    "temperature": 22.5,
    "humidity": 45.0,
    "power_consumption": 2.3,
    "mode": "cooling",
    "setpoint": 22.0
  },
  "source": "MQTT"
}
```

### Device Command

```json
{
  "id": "507f1f77bcf86cd799439013",
  "commandId": "cmd-001",
  "deviceId": "hvac-001",
  "command": "SET_TEMPERATURE",
  "params": {
    "temperature": 24.0,
    "unit": "celsius"
  },
  "status": "APPLIED",
  "issuedBy": "user123",
  "sentAt": "2024-01-20T10:35:00Z",
  "appliedAt": "2024-01-20T10:35:02Z",
  "createdAt": "2024-01-20T10:35:00Z",
  "updatedAt": "2024-01-20T10:35:02Z"
}
```

### Command Status

| Status | Description |
|--------|-------------|
| `PENDING` | Command created, not yet sent |
| `SENT` | Command sent to device |
| `APPLIED` | Device confirmed command execution |
| `FAILED` | Command execution failed |
| `CANCELLED` | Command was cancelled |
| `TIMEOUT` | No response within timeout period |

### Optimization Scenario (IoT)

```json
{
  "scenarioId": "scenario-001",
  "forecastId": "forecast-001",
  "buildingId": "building-001",
  "actions": [
    {
      "deviceId": "hvac-001",
      "command": "SET_TEMPERATURE",
      "params": {"temperature": 24.0}
    }
  ],
  "executionStatus": "RUNNING",
  "progress": 0.5,
  "createdBy": "user123"
}
```

---

## API Endpoints

### Telemetry

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/iot/telemetry` | Ingest single telemetry | Yes |
| `POST` | `/api/v1/iot/telemetry/bulk` | Ingest bulk telemetry | Yes |
| `GET` | `/api/v1/iot/telemetry/history` | Get telemetry history | Yes |

#### POST /api/v1/iot/telemetry

**Request:**
```json
{
  "deviceId": "hvac-001",
  "timestamp": "2024-01-20T10:30:00Z",
  "metrics": {
    "temperature": 22.5,
    "humidity": 45.0,
    "power_consumption": 2.3
  }
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "deviceId": "hvac-001",
    "timestamp": "2024-01-20T10:30:00Z",
    "metrics": {...},
    "source": "HTTP"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8083/api/v1/iot/telemetry \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "deviceId": "hvac-001",
    "metrics": {
      "temperature": 22.5,
      "humidity": 45.0,
      "power_consumption": 2.3
    }
  }'
```

#### POST /api/v1/iot/telemetry/bulk

**Request:**
```json
{
  "telemetry": [
    {
      "deviceId": "hvac-001",
      "metrics": {"temperature": 22.5}
    },
    {
      "deviceId": "hvac-002",
      "metrics": {"temperature": 23.0}
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "inserted": 2,
    "failed": 0
  }
}
```

#### GET /api/v1/iot/telemetry/history

**Query Parameters:**
- `deviceId` (required) - Device identifier
- `from` - Start timestamp (ISO 8601)
- `to` - End timestamp (ISO 8601)
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 100)

**Example:**
```bash
curl -X GET "http://localhost:8083/api/v1/iot/telemetry/history?deviceId=hvac-001&from=2024-01-20T00:00:00Z&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

---

### Device Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/iot/devices` | List devices | Yes |
| `GET` | `/api/v1/iot/devices/:deviceId` | Get device by ID | Yes |
| `POST` | `/api/v1/iot/devices/register` | Register new device | Yes |

#### GET /api/v1/iot/devices

**Query Parameters:**
- `buildingId` - Filter by building
- `type` - Filter by device type
- `status` - Filter by status
- `page` - Page number
- `limit` - Items per page

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "devices": [
      {
        "id": "507f1f77bcf86cd799439011",
        "deviceId": "hvac-001",
        "type": "HVAC",
        "status": "ONLINE",
        "location": {...},
        "lastSeen": "2024-01-20T10:30:00Z"
      }
    ],
    "total": 25,
    "page": 1,
    "limit": 10
  }
}
```

**Example:**
```bash
curl -X GET "http://localhost:8083/api/v1/iot/devices?buildingId=building-001&status=ONLINE" \
  -H "Authorization: Bearer $TOKEN"
```

#### POST /api/v1/iot/devices/register

**Request:**
```json
{
  "deviceId": "sensor-003",
  "type": "TEMPERATURE_SENSOR",
  "model": "Honeywell T6 Pro",
  "location": {
    "buildingId": "building-001",
    "floor": "2",
    "room": "Server Room"
  },
  "capabilities": ["READ_TEMPERATURE", "READ_HUMIDITY"],
  "metadata": {
    "firmwareVersion": "1.0.0"
  }
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439014",
    "deviceId": "sensor-003",
    "type": "TEMPERATURE_SENSOR",
    "status": "OFFLINE",
    "createdAt": "2024-01-20T11:00:00Z"
  }
}
```

---

### Device Control

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/iot/device-control/:deviceId/command` | Send command | Yes |
| `GET` | `/api/v1/iot/device-control/:deviceId/commands` | List commands | Yes |

#### POST /api/v1/iot/device-control/:deviceId/command

**Request:**
```json
{
  "command": "SET_TEMPERATURE",
  "params": {
    "temperature": 24.0,
    "unit": "celsius"
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439015",
    "commandId": "cmd-002",
    "deviceId": "hvac-001",
    "command": "SET_TEMPERATURE",
    "params": {...},
    "status": "SENT",
    "createdAt": "2024-01-20T11:05:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8083/api/v1/iot/device-control/hvac-001/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "command": "SET_TEMPERATURE",
    "params": {"temperature": 24.0}
  }'
```

#### GET /api/v1/iot/device-control/:deviceId/commands

**Query Parameters:**
- `status` - Filter by command status
- `page` - Page number
- `limit` - Items per page

---

### Optimization

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/iot/optimization/applySecurity` | Apply optimization scenario | Yes |
| `POST` | `/api/v1/iot/optimization/apply` | Apply optimization (legacy) | Yes |
| `GET` | `/api/v1/iot/optimization/status/:scenarioId` | Get optimization status | Yes |

#### POST /api/v1/iot/optimization/applySecurity

Applies an optimization scenario received from the Forecast Service.

**Request:**
```json
{
  "scenarioId": "scenario-001",
  "forecastId": "forecast-001",
  "buildingId": "building-001",
  "actions": [
    {
      "deviceId": "hvac-001",
      "command": "SET_TEMPERATURE",
      "params": {"temperature": 24.0}
    },
    {
      "deviceId": "hvac-002",
      "command": "SET_TEMPERATURE",
      "params": {"temperature": 25.0}
    }
  ]
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "scenarioId": "scenario-001",
    "executionStatus": "PENDING",
    "progress": 0.0,
    "actionsCount": 2
  }
}
```

**Integration Flow:**
1. Forecast Service sends optimization scenario
2. IoT Service fetches device predictions from Forecast Service
3. IoT Service checks for anomalies from Analytics Service
4. IoT Service executes actions with priority based on device trends
5. Progress is tracked and can be queried

**Example:**
```bash
curl -X POST http://localhost:8083/api/v1/iot/optimization/applySecurity \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "scenarioId": "scenario-001",
    "buildingId": "building-001",
    "actions": [
      {"deviceId": "hvac-001", "command": "SET_TEMPERATURE", "params": {"temperature": 24}}
    ]
  }'
```

#### GET /api/v1/iot/optimization/status/:scenarioId

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "scenarioId": "scenario-001",
    "executionStatus": "RUNNING",
    "progress": 0.6,
    "actions": [
      {"deviceId": "hvac-001", "status": "APPLIED", "commandId": "cmd-001"},
      {"deviceId": "hvac-002", "status": "SENT", "commandId": "cmd-002"}
    ]
  }
}
```

---

### State

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/iot/state/live` | Get live device states | Yes |
| `GET` | `/api/v1/iot/state/:deviceId` | Get specific device state | Yes |

#### GET /api/v1/iot/state/live

**Query Parameters:**
- `buildingId` - Filter by building

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "devices": [
      {
        "deviceId": "hvac-001",
        "status": "ONLINE",
        "lastTelemetry": {
          "temperature": 22.5,
          "humidity": 45.0,
          "power_consumption": 2.3
        },
        "lastSeen": "2024-01-20T10:30:00Z"
      }
    ],
    "timestamp": "2024-01-20T10:35:00Z"
  }
}
```

---

## MQTT Communication

The service uses MQTT for real-time bidirectional communication with IoT devices.

### Topic Structure

| Topic | Direction | Description |
|-------|-----------|-------------|
| `mqtt/iot/{deviceId}/telemetry` | Device → Service | Device sends telemetry data |
| `mqtt/iot/{deviceId}/command` | Service → Device | Service sends commands to device |
| `mqtt/iot/{deviceId}/ack` | Device → Service | Device acknowledges command |
| `mqtt/iot/broadcast/announcement` | Service → All | Broadcast to all devices |

### Telemetry Message Format

```json
{
  "deviceId": "hvac-001",
  "timestamp": "2024-01-20T10:30:00Z",
  "metrics": {
    "temperature": 22.5,
    "humidity": 45.0,
    "power_consumption": 2.3
  }
}
```

### Command Message Format

```json
{
  "commandId": "cmd-001",
  "command": "SET_TEMPERATURE",
  "params": {
    "temperature": 24.0
  },
  "timestamp": "2024-01-20T10:35:00Z"
}
```

### Acknowledgment Message Format

```json
{
  "commandId": "cmd-001",
  "deviceId": "hvac-001",
  "status": "APPLIED",
  "errorMsg": "",
  "timestamp": "2024-01-20T10:35:02Z"
}
```

### MQTT Configuration

```yaml
Broker: mqtt-broker
Port: 1883
ClientID: iot-control-service
QoS: 1 (At least once delivery)
```

---

## Authentication

All HTTP endpoints (except health check) require JWT authentication. MQTT communication uses service-level credentials.

**HTTP Header Format:**
```
Authorization: Bearer <access_token>
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8083` |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `GIN_MODE` | Gin framework mode | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `iot_control_service` |
| `MONGODB_TIMEOUT` | Connection timeout (seconds) | `10` |
| `SECURITY_SERVICE_URL` | Security service URL | `http://localhost:8080` |
| `SECURITY_SERVICE_TIMEOUT` | Security service timeout (seconds) | `10` |
| `FORECAST_SERVICE_URL` | Forecast service URL | `http://localhost:8082` |
| `FORECAST_SERVICE_TIMEOUT` | Forecast service timeout (seconds) | `10` |
| `ANALYTICS_SERVICE_URL` | Analytics service URL | `http://localhost:8084` |
| `ANALYTICS_SERVICE_TIMEOUT` | Analytics service timeout (seconds) | `10` |
| `STORAGE_SERVICE_URL` | Storage service URL | `http://localhost:8086/storage` |
| `STORAGE_SERVICE_TIMEOUT` | Storage service timeout (seconds) | `10` |
| `MQTT_BROKER` | MQTT broker host | `localhost` |
| `MQTT_PORT` | MQTT broker port | `1883` |
| `MQTT_USERNAME` | MQTT username | - |
| `MQTT_PASSWORD` | MQTT password | - |
| `MQTT_CLIENT_ID` | MQTT client identifier | `iot-control-service` |
| `MQTT_QOS` | MQTT Quality of Service level | `1` |
| `IOT_TELEMETRY_BATCH_SIZE` | Max telemetry batch size | `100` |
| `IOT_COMMAND_TIMEOUT` | Command timeout (seconds) | `30` |
| `IOT_STATE_UPDATE_INTERVAL` | State update interval (seconds) | `5` |
| `LOG_LEVEL` | Logging level | `debug` |
| `LOG_FORMAT` | Log format | `json` |

### Example .env File

```env
SERVER_PORT=8083
GIN_MODE=release

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=iot_control_service

SECURITY_SERVICE_URL=http://security-service:8080
FORECAST_SERVICE_URL=http://forecast-service:8082
ANALYTICS_SERVICE_URL=http://analytics-service:8084

MQTT_BROKER=mqtt-broker
MQTT_PORT=1883
MQTT_CLIENT_ID=iot-control-service
MQTT_QOS=1

IOT_TELEMETRY_BATCH_SIZE=100
IOT_COMMAND_TIMEOUT=30

LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Database Schema

### MongoDB Collections

| Collection | Description | Indexes |
|------------|-------------|---------|
| `devices` | Device registry | `device_id` (unique), `location.building_id`, `status` |
| `telemetry` | Telemetry data | `device_id`, `timestamp` |
| `device_commands` | Command history | `command_id`, `device_id`, `status`, `created_at` |
| `optimization_scenarios` | Optimization executions | `scenario_id`, `building_id`, `execution_status` |

---

## Inter-Service Communication

### Inbound (Other services calling IoT Service)

| Caller | Endpoint | Purpose |
|--------|----------|---------|
| Forecast Service | `POST /api/v1/iot/optimization/applySecurity` | Apply optimization scenarios |
| Analytics Service | `GET /api/v1/iot/devices` | Get device list for dashboards |
| Analytics Service | `GET /api/v1/iot/telemetry/history` | Get telemetry for analysis |

### Outbound (IoT Service calling other services)

| Target | Endpoint | Purpose |
|--------|----------|---------|
| Security Service | `GET /api/v1/auth/validate-token` | Validate JWT tokens |
| Forecast Service | `GET /api/v1/forecast/prediction/:deviceId` | Get device predictions for optimization |
| Analytics Service | `GET /api/v1/analytics/anomalies` | Check for device anomalies |
| Storage Service | Various | Persist telemetry and command data |

---

## Running the Service

### Prerequisites

- Go 1.21+
- MongoDB 7.0+
- MQTT Broker (Eclipse Mosquitto 2.0+)
- Docker (optional)

### Local Development

```bash
# Start MQTT broker
docker run -d -p 1883:1883 eclipse-mosquitto:2.0

# Navigate to service directory
cd iot-control-service

# Install dependencies
go mod download

# Set environment variables
export MONGODB_URI=mongodb://localhost:27017
export MQTT_BROKER=localhost
export SECURITY_SERVICE_URL=http://localhost:8080

# Run the service
go run cmd/main.go
```

### Using Docker

```bash
# Build the image
docker build -t iot-control-service ./iot-control-service

# Run the container
docker run -p 8083:8083 \
  -e MONGODB_URI=mongodb://host.docker.internal:27017 \
  -e MQTT_BROKER=host.docker.internal \
  -e SECURITY_SERVICE_URL=http://security-service:8080 \
  iot-control-service
```

### Using Docker Compose

```bash
# Start all services including MQTT broker
docker-compose up -d mqtt-broker iot-control-service

# View logs
docker-compose logs -f iot-control-service
```

---

## Health Check

```bash
curl http://localhost:8083/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "iot-control-service"
}
```

---

## Example Usage

### Complete Device Workflow

```bash
# 1. Register a new device
curl -X POST http://localhost:8083/api/v1/iot/devices/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "deviceId": "hvac-003",
    "type": "HVAC",
    "model": "Carrier 24ACC636A003",
    "location": {"buildingId": "building-001", "floor": "3", "room": "Office 301"},
    "capabilities": ["SET_TEMPERATURE", "SET_MODE", "TURN_OFF"]
  }'

# 2. Send telemetry (simulating device)
curl -X POST http://localhost:8083/api/v1/iot/telemetry \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "deviceId": "hvac-003",
    "metrics": {"temperature": 23.5, "humidity": 50.0, "power_consumption": 2.5}
  }'

# 3. Send command to device
curl -X POST http://localhost:8083/api/v1/iot/device-control/hvac-003/command \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"command": "SET_TEMPERATURE", "params": {"temperature": 22.0}}'

# 4. Check device state
curl -X GET http://localhost:8083/api/v1/iot/state/hvac-003 \
  -H "Authorization: Bearer $TOKEN"

# 5. View command history
curl -X GET "http://localhost:8083/api/v1/iot/device-control/hvac-003/commands?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### MQTT Telemetry Ingestion (Python Example)

```python
import paho.mqtt.client as mqtt
import json
from datetime import datetime

def on_connect(client, userdata, flags, rc):
    print(f"Connected with result code {rc}")

client = mqtt.Client()
client.on_connect = on_connect
client.connect("localhost", 1883, 60)

# Send telemetry
telemetry = {
    "deviceId": "hvac-001",
    "timestamp": datetime.utcnow().isoformat() + "Z",
    "metrics": {
        "temperature": 22.5,
        "humidity": 45.0,
        "power_consumption": 2.3
    }
}

client.publish("mqtt/iot/hvac-001/telemetry", json.dumps(telemetry))
```

### Apply Optimization Scenario

```bash
curl -X POST http://localhost:8083/api/v1/iot/optimization/applySecurity \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "scenarioId": "scenario-001",
    "buildingId": "building-001",
    "forecastId": "forecast-001",
    "actions": [
      {"deviceId": "hvac-001", "command": "SET_TEMPERATURE", "params": {"temperature": 24}},
      {"deviceId": "hvac-002", "command": "SET_TEMPERATURE", "params": {"temperature": 25}},
      {"deviceId": "lighting-001", "command": "DIM", "params": {"level": 70}}
    ]
  }'

# Check optimization status
curl -X GET http://localhost:8083/api/v1/iot/optimization/status/scenario-001 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Known Limitations

1. **MQTT Dependency:** Real-time features require MQTT broker availability
2. **Command Timeout:** Commands timeout after 30 seconds by default
3. **No Device Authentication:** Devices authenticate via MQTT broker, not per-device tokens
4. **Telemetry Retention:** No automatic telemetry data cleanup (implement TTL indexes)
5. **Single MQTT Broker:** No MQTT broker clustering support
6. **No Batch Commands:** Commands are sent individually, not batched

---

## Developer Notes

- MQTT subscriptions are set up automatically on service startup
- Device `lastSeen` is updated automatically on telemetry receipt
- Command status is updated asynchronously via MQTT acknowledgments
- Optimization scenarios use device predictions to prioritize actions
- Bulk telemetry ingestion is more efficient for high-frequency data
- Use QoS 1 for reliable command delivery
- Monitor MQTT connection status for device health
