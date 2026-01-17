# EMSIB - Enterprise Microservices Infrastructure for Buildings

## Overview

**EMSIB** is a comprehensive microservices-based platform for managing building energy systems, IoT devices, forecasting energy demand, and providing real-time analytics. The system enables smart building management through telemetry collection, automated optimization, anomaly detection, and actionable insights.

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Microservices](#microservices)
- [Inter-Service Communication](#inter-service-communication)
- [Data Flow](#data-flow)
- [Authentication & Security](#authentication--security)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              EMSIB Platform                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                           API Gateway / Load Balancer                    │    │
│  └───────────────────────────────────┬─────────────────────────────────────┘    │
│                                      │                                           │
│  ┌───────────────┬───────────────────┼───────────────────┬───────────────┐      │
│  │               │                   │                   │               │      │
│  ▼               ▼                   ▼                   ▼               ▼      │
│ ┌───────────┐ ┌───────────┐ ┌───────────────┐ ┌───────────┐ ┌───────────┐      │
│ │ Security  │ │ Forecast  │ │  IoT Control  │ │ Analytics │ │  Storage  │      │
│ │  Service  │ │  Service  │ │    Service    │ │  Service  │ │  Service  │      │
│ │  :8080    │ │  :8082    │ │    :8083      │ │  :8084    │ │  :8086    │      │
│ └─────┬─────┘ └─────┬─────┘ └───────┬───────┘ └─────┬─────┘ └───────────┘      │
│       │             │               │               │                           │
│       └─────────────┴───────┬───────┴───────────────┘                           │
│                             │                                                    │
│  ┌──────────────────────────┴──────────────────────────┐                        │
│  │                       MongoDB                        │                        │
│  │                       :27018                         │                        │
│  └──────────────────────────────────────────────────────┘                        │
│                             │                                                    │
│  ┌──────────────────────────┴──────────────────────────┐                        │
│  │                    MQTT Broker                       │                        │
│  │                 :1883 / :9001                        │                        │
│  └──────────────────────────────────────────────────────┘                        │
│                             │                                                    │
│  ┌──────────────────────────┴──────────────────────────┐                        │
│  │                    IoT Devices                       │                        │
│  │           HVAC | Sensors | Lighting | Meters         │                        │
│  └──────────────────────────────────────────────────────┘                        │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                        External Services                                 │    │
│  │         Weather API | Tariff API | ML Service | Notification Service    │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Microservices

| Service | Port | Description | Documentation |
|---------|------|-------------|---------------|
| **Security Service** | 8080 | Authentication, authorization, user/role management, audit logging | [README_security.md](./README_security.md) |
| **Forecast Service** | 8082 | Energy demand forecasting, peak load prediction, optimization scenarios | [README_forecast.md](./README_forecast.md) |
| **IoT Control Service** | 8083 | Device management, telemetry ingestion, command execution, MQTT communication | [README_iot.md](./README_iot.md) |
| **Analytics Service** | 8084 | Reports, KPIs, anomaly detection, dashboards, time-series analysis | [README_analytics.md](./README_analytics.md) |

### Infrastructure Services

| Service | Port | Description |
|---------|------|-------------|
| **MongoDB** | 27018 | Primary database for all microservices |
| **MQTT Broker** | 1883, 9001 | Message broker for IoT device communication |

---

## Inter-Service Communication

### Service Dependency Graph

```
                    ┌─────────────────┐
                    │ Security Service│
                    │     :8080       │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Forecast Service│ │IoT Control Svc  │ │Analytics Service│
│     :8082       │ │     :8083       │ │     :8084       │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         │         ┌─────────┴─────────┐         │
         │         │                   │         │
         └─────────┼───────────────────┼─────────┘
                   │                   │
                   ▼                   ▼
            ┌───────────┐       ┌───────────┐
            │  MongoDB  │       │MQTT Broker│
            └───────────┘       └───────────┘
```

### Integration Matrix

| From \ To | Security | Forecast | IoT Control | Analytics | Storage |
|-----------|----------|----------|-------------|-----------|---------|
| **Security** | - | - | - | - | Credentials, Audit |
| **Forecast** | Token validation | - | Apply optimization | - | Forecasts |
| **IoT Control** | Token validation | Get predictions | - | Get anomalies | Telemetry, Commands |
| **Analytics** | Token validation | Get forecasts | Get devices/telemetry | - | Reports, KPIs |

### Key Integration Flows

#### 1. Token Validation Flow
All services validate JWT tokens with the Security Service:
```
Any Service → GET /api/v1/auth/validate-token → Security Service
```

#### 2. Optimization Execution Flow
```
1. Forecast Service generates optimization scenario
2. Forecast Service → POST /api/v1/iot/optimization/applySecurity → IoT Control Service
3. IoT Control Service → GET /api/v1/forecast/prediction/:deviceId → Forecast Service
4. IoT Control Service → GET /api/v1/analytics/anomalies → Analytics Service
5. IoT Control Service executes actions on devices via MQTT
```

#### 3. Dashboard Data Aggregation Flow
```
1. Client → GET /api/v1/analytics/dashboards/building/:id → Analytics Service
2. Analytics Service → GET /api/v1/iot/devices → IoT Control Service
3. Analytics Service → GET /api/v1/forecast/latest → Forecast Service
4. Analytics Service aggregates and returns dashboard data
```

---

## Data Flow

### Telemetry Flow

```
┌──────────┐    MQTT     ┌───────────────┐   Store    ┌─────────┐
│  Device  │────────────▶│  IoT Control  │───────────▶│ MongoDB │
└──────────┘             │    Service    │            └─────────┘
                         └───────┬───────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
                    ▼                         ▼
            ┌───────────────┐        ┌───────────────┐
            │   Analytics   │        │   Forecast    │
            │    Service    │        │    Service    │
            │               │        │               │
            │ - Anomaly Det │        │ - Predictions │
            │ - KPI Calc    │        │ - Peak Load   │
            │ - Reports     │        │ - Optimization│
            └───────────────┘        └───────────────┘
```

### Command Flow

```
┌────────┐   Request    ┌───────────────┐   MQTT    ┌──────────┐
│ Client │─────────────▶│  IoT Control  │─────────▶│  Device  │
└────────┘              │    Service    │          └────┬─────┘
                        └───────┬───────┘               │
                                │                       │ ACK
                                │◀──────────────────────┘
                                │
                        ┌───────▼───────┐
                        │    Update     │
                        │ Command Status│
                        └───────────────┘
```

---

## Authentication & Security

### Authentication Flow

```
┌────────┐   Login     ┌──────────────┐
│ Client │────────────▶│   Security   │
└────────┘             │   Service    │
    │                  └──────┬───────┘
    │                         │
    │◀────────────────────────┘
    │  JWT (Access + Refresh)
    │
    │   API Request + Bearer Token
    ▼
┌──────────────────────────────────────────────┐
│              Any Microservice                 │
│                                               │
│  1. Extract token from Authorization header   │
│  2. Validate with Security Service            │
│  3. Process request if valid                  │
└──────────────────────────────────────────────┘
```

### JWT Token Structure

```json
{
  "userId": "507f1f77bcf86cd799439011",
  "username": "john.doe",
  "email": "john.doe@example.com",
  "roles": ["admin", "building_manager"],
  "exp": 1705750800,
  "iat": 1705749900
}
```

### Role-Based Access Control (RBAC)

| Role | Description | Typical Permissions |
|------|-------------|---------------------|
| `admin` | Full system access | All resources, all actions |
| `user` | Basic access | Read access to most resources |
| `building_manager` | Building management | Read/write buildings, devices |
| `energy_analyst` | Read-only analytics | Read reports, analytics, forecasts |

---

## Getting Started

### Prerequisites

- **Go** 1.21 or higher
- **Docker** 20.10 or higher
- **Docker Compose** 2.0 or higher
- **MongoDB** 7.0 (provided via Docker)
- **MQTT Broker** Eclipse Mosquitto 2.0 (provided via Docker)

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd Software-Engineering-Project

# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health  # Security
curl http://localhost:8082/health  # Forecast
curl http://localhost:8083/health  # IoT Control
curl http://localhost:8084/health  # Analytics

# View logs
docker-compose logs -f
```

### Default Credentials

After first startup, the system creates a default admin user:

| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `admin123` |
| Roles | `admin` |

**Important:** Change the default password immediately in production!

### First Steps

```bash
# 1. Login to get a token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.accessToken')

# 2. Register a device
curl -X POST http://localhost:8083/api/v1/iot/devices/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "deviceId": "hvac-001",
    "type": "HVAC",
    "model": "Carrier 24ACC636A003",
    "location": {"buildingId": "building-001", "floor": "1", "room": "Lobby"},
    "capabilities": ["SET_TEMPERATURE", "SET_MODE", "TURN_OFF"]
  }'

# 3. Send telemetry
curl -X POST http://localhost:8083/api/v1/iot/telemetry \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "deviceId": "hvac-001",
    "metrics": {"temperature": 22.5, "humidity": 45.0, "power_consumption": 2.3}
  }'

# 4. Generate a forecast
curl -X POST http://localhost:8082/api/v1/forecast/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "buildingId": "building-001",
    "type": "DEMAND",
    "horizonHours": 24,
    "includeWeather": true
  }'

# 5. View analytics dashboard
curl http://localhost:8084/api/v1/analytics/dashboards/overview \
  -H "Authorization: Bearer $TOKEN"
```

---

## Configuration

### Environment Variables by Service

#### Global Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GIN_MODE` | Gin framework mode (`debug`/`release`) | `debug` |
| `LOG_LEVEL` | Logging level | `debug` |
| `LOG_FORMAT` | Log format (`json`/`text`) | `json` |

#### MongoDB Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | Service-specific |
| `MONGODB_TIMEOUT` | Connection timeout (seconds) | `10` |

#### Service URLs

| Variable | Description | Default |
|----------|-------------|---------|
| `SECURITY_SERVICE_URL` | Security service URL | `http://localhost:8080` |
| `FORECAST_SERVICE_URL` | Forecast service URL | `http://localhost:8082` |
| `IOT_SERVICE_URL` | IoT Control service URL | `http://localhost:8083` |
| `ANALYTICS_SERVICE_URL` | Analytics service URL | `http://localhost:8084` |
| `STORAGE_SERVICE_URL` | Storage service URL | `http://localhost:8086/storage` |

#### MQTT Configuration (IoT Control Service)

| Variable | Description | Default |
|----------|-------------|---------|
| `MQTT_BROKER` | MQTT broker host | `localhost` |
| `MQTT_PORT` | MQTT broker port | `1883` |
| `MQTT_CLIENT_ID` | MQTT client identifier | `iot-control-service` |
| `MQTT_QOS` | Quality of Service level | `1` |

### Sample Docker Compose Override

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  security-service:
    environment:
      - JWT_SECRET=your-production-secret-key
      - GIN_MODE=release
      - LOG_LEVEL=info
```

---

## API Reference

### Base URLs

| Service | URL |
|---------|-----|
| Security | `http://localhost:8080/api/v1` |
| Forecast | `http://localhost:8082/api/v1` |
| IoT Control | `http://localhost:8083/api/v1` |
| Analytics | `http://localhost:8084/api/v1` |

### Quick Reference

#### Authentication

```bash
# Login
POST /api/v1/auth/login
Body: {"username": "user", "password": "pass"}

# Refresh token
POST /api/v1/auth/refresh
Body: {"refreshToken": "..."}

# Validate token (internal)
GET /api/v1/auth/validate-token
Header: Authorization: Bearer <token>
```

#### Devices

```bash
# List devices
GET /api/v1/iot/devices?buildingId=...&status=ONLINE

# Register device
POST /api/v1/iot/devices/register
Body: {"deviceId": "...", "type": "...", "location": {...}}

# Send command
POST /api/v1/iot/device-control/:deviceId/command
Body: {"command": "SET_TEMPERATURE", "params": {"temperature": 24}}
```

#### Telemetry

```bash
# Ingest telemetry
POST /api/v1/iot/telemetry
Body: {"deviceId": "...", "metrics": {...}}

# Get history
GET /api/v1/iot/telemetry/history?deviceId=...&from=...&to=...
```

#### Forecasting

```bash
# Generate forecast
POST /api/v1/forecast/generate
Body: {"buildingId": "...", "type": "DEMAND", "horizonHours": 24}

# Get latest forecast
GET /api/v1/forecast/latest?buildingId=...

# Generate optimization
POST /api/v1/optimization/generate
Body: {"buildingId": "...", "type": "PEAK_SHAVING", ...}
```

#### Analytics

```bash
# Dashboard overview
GET /api/v1/analytics/dashboards/overview

# Building dashboard
GET /api/v1/analytics/dashboards/building/:buildingId

# List anomalies
GET /api/v1/analytics/anomalies?severity=HIGH&status=NEW

# Generate report
POST /api/v1/analytics/reports/generate
Body: {"buildingId": "...", "type": "ENERGY_CONSUMPTION", ...}
```

---

## Testing

### Running the Test Script

A comprehensive test script is provided to test all endpoints:

```bash
# Make the script executable
chmod +x test_all_endpoints.sh

# Run all tests
./test_all_endpoints.sh

# The script will:
# - Test all services' health endpoints
# - Test authentication flows
# - Test CRUD operations
# - Test edge cases
# - Display color-coded results (green=pass, red=fail)
```

### Manual Testing

```bash
# Set up authentication
export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.accessToken')

# Test Security Service
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/users

# Test IoT Service
curl -H "Authorization: Bearer $TOKEN" http://localhost:8083/api/v1/iot/devices

# Test Forecast Service
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8082/api/v1/forecast/latest?buildingId=building-001"

# Test Analytics Service
curl -H "Authorization: Bearer $TOKEN" http://localhost:8084/api/v1/analytics/dashboards/overview
```

### Using curl with HTTPie

```bash
# Install HTTPie (optional, for prettier output)
pip install httpie

# Login
http POST localhost:8080/api/v1/auth/login username=admin password=admin123

# List devices
http GET localhost:8083/api/v1/iot/devices "Authorization: Bearer $TOKEN"
```

---

## Deployment

### Production Checklist

- [ ] Change default JWT secret (`JWT_SECRET`)
- [ ] Change default encryption key (`ENCRYPTION_KEY`)
- [ ] Change default admin password
- [ ] Set `GIN_MODE=release`
- [ ] Set appropriate `LOG_LEVEL` (info or warn)
- [ ] Configure HTTPS/TLS termination
- [ ] Set up database backups
- [ ] Configure MQTT authentication
- [ ] Set up monitoring and alerting
- [ ] Configure rate limiting at API gateway

### Docker Compose Production

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  security-service:
    environment:
      - GIN_MODE=release
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```

### Health Checks

All services expose a health endpoint:

```bash
# Check all services
for port in 8080 8082 8083 8084; do
  echo "Checking port $port..."
  curl -s http://localhost:$port/health | jq
done
```

---

## Troubleshooting

### Common Issues

#### Service won't start

```bash
# Check logs
docker-compose logs <service-name>

# Common causes:
# - MongoDB not ready (wait for health check)
# - Port already in use
# - Missing environment variables
```

#### Token validation fails

```bash
# Ensure Security Service is running
curl http://localhost:8080/health

# Check token expiry
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq '.exp'
```

#### MQTT connection issues

```bash
# Check MQTT broker
docker-compose logs mqtt-broker

# Test MQTT connection
mosquitto_sub -h localhost -p 1883 -t '#' -v
```

#### MongoDB connection issues

```bash
# Check MongoDB
docker-compose logs mongodb

# Test connection
docker exec -it mongodb mongosh --eval "db.adminCommand('ping')"
```

### Debug Mode

Enable debug logging:

```bash
# In docker-compose.override.yml
environment:
  - LOG_LEVEL=debug
  - GIN_MODE=debug
```

---

## Contributing

### Code Structure

```
Software-Engineering-Project/
├── security-service/          # Authentication & authorization
│   ├── cmd/                   # Entry point
│   ├── internal/              # Internal packages
│   │   ├── handlers/          # HTTP handlers
│   │   ├── service/           # Business logic
│   │   ├── repository/        # Data access
│   │   ├── models/            # Domain models
│   │   ├── middleware/        # HTTP middleware
│   │   └── integrations/      # External clients
│   └── Dockerfile
├── forecast-service/          # Forecasting & optimization
├── iot-control-service/       # Device management & control
├── analytics-service/         # Analytics & reporting
├── mqtt/                      # MQTT broker config
├── docker-compose.yml         # Docker Compose config
├── test_all_endpoints.sh      # Endpoint test script
├── README.md                  # This file
├── README_security.md         # Security Service docs
├── README_forecast.md         # Forecast Service docs
├── README_iot.md              # IoT Service docs
└── README_analytics.md        # Analytics Service docs
```

### Development Workflow

```bash
# Run single service locally
cd <service-directory>
go run cmd/main.go

# Run tests
go test ./...

# Build Docker image
docker build -t <service-name> .
```

### Adding a New Service

1. Create service directory with standard structure
2. Implement handlers, services, repositories
3. Add Security Service client for auth
4. Add to `docker-compose.yml`
5. Update this README with service documentation

---

## License

This project is part of the Software Engineering Project coursework.

---

## Support

For issues and feature requests, please create an issue in the project repository.
