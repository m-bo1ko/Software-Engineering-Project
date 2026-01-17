# Security Service

## Overview

The **Security Service** is the centralized authentication and authorization microservice for the EMSIB (Enterprise Microservices Infrastructure for Buildings) platform. It handles user management, role-based access control (RBAC), JWT token management, audit logging, notifications, and external energy provider integrations.

**Port:** `8080`

---

## Table of Contents

- [Architecture](#architecture)
- [Data Models](#data-models)
- [API Endpoints](#api-endpoints)
- [Authentication & Authorization](#authentication--authorization)
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
│                      Security Service                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Handlers  │  │  Middleware │  │   Router    │              │
│  │             │  │             │  │             │              │
│  │ - Auth      │  │ - Auth      │  │ - v1 routes │              │
│  │ - User      │  │ - CORS      │  │ - Legacy    │              │
│  │ - Role      │  │ - Security  │  │   routes    │              │
│  │ - Audit     │  │ - Logging   │  │             │              │
│  │ - Notify    │  │ - Recovery  │  │             │              │
│  │ - Energy    │  │             │  │             │              │
│  └──────┬──────┘  └─────────────┘  └─────────────┘              │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────┐        │
│  │                    Services                          │        │
│  │  Auth | User | Role | Audit | Notification          │        │
│  └──────┬──────────────────────────────────────────────┘        │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────┐        │
│  │                  Repositories                        │        │
│  │  User | Role | Auth | Audit | Notification          │        │
│  └──────┬──────────────────────────────────────────────┘        │
│         │                                                        │
│  ┌──────▼──────────────────────────────────────────────┐        │
│  │                    MongoDB                           │        │
│  └──────────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────┘
```

### Internal Layers

| Layer | Description |
|-------|-------------|
| **Handlers** | HTTP request handlers that validate input and call services |
| **Middleware** | Authentication, CORS, security headers, logging, recovery |
| **Services** | Business logic layer containing domain operations |
| **Repositories** | Data access layer for MongoDB operations |
| **Models** | Domain entities and DTOs |
| **Integrations** | Clients for external services (Storage, Notifications, Energy) |

---

## Data Models

### User

```json
{
  "id": "507f1f77bcf86cd799439011",
  "username": "john.doe",
  "email": "john.doe@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "roles": ["admin", "building_manager"],
  "isActive": true,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "lastLoginAt": "2024-01-20T08:15:00Z"
}
```

### Role

```json
{
  "id": "507f1f77bcf86cd799439012",
  "name": "building_manager",
  "description": "Manager with access to building and energy data",
  "permissions": [
    {
      "resource": "buildings",
      "actions": ["read", "write"]
    },
    {
      "resource": "devices",
      "actions": ["read"]
    }
  ],
  "isSystem": false,
  "createdAt": "2024-01-10T00:00:00Z",
  "updatedAt": "2024-01-10T00:00:00Z"
}
```

### Audit Log

```json
{
  "id": "507f1f77bcf86cd799439013",
  "userId": "user123",
  "username": "john.doe",
  "service": "security-service",
  "action": "LOGIN",
  "resource": "auth",
  "resourceId": "",
  "details": {"method": "password"},
  "ipAddress": "192.168.1.100",
  "userAgent": "Mozilla/5.0...",
  "status": "SUCCESS",
  "timestamp": "2024-01-20T08:15:00Z",
  "requestPath": "/api/v1/auth/login",
  "method": "POST"
}
```

---

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/login` | User login | No |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | No |
| `POST` | `/api/v1/auth/logout` | User logout | Yes |
| `GET` | `/api/v1/auth/validate-token` | Validate JWT token (internal) | No |
| `POST` | `/api/v1/auth/check-permissions` | Check user permissions (internal) | No |
| `GET` | `/api/v1/auth/user-info` | Get current user info | Yes |

#### POST /api/v1/auth/login

**Request:**
```json
{
  "username": "john.doe",
  "password": "securePassword123"
}
```

**Response (200 OK):**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
  "tokenType": "Bearer",
  "expiresIn": 900,
  "roles": ["admin"],
  "userId": "507f1f77bcf86cd799439011"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

#### GET /api/v1/auth/validate-token

Used internally by other microservices to validate tokens.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/validate-token \
  -H "Authorization: Bearer <access_token>"
```

**Response (200 OK):**
```json
{
  "valid": true,
  "userId": "507f1f77bcf86cd799439011",
  "roles": ["admin", "building_manager"]
}
```

#### POST /api/v1/auth/check-permissions

**Request:**
```json
{
  "userId": "507f1f77bcf86cd799439011",
  "resource": "buildings",
  "action": "write"
}
```

**Response (200 OK):**
```json
{
  "allowed": true,
  "reason": ""
}
```

---

### User Management

| Method | Endpoint | Description | Auth Required | Admin Only |
|--------|----------|-------------|---------------|------------|
| `GET` | `/api/v1/users` | List all users | Yes | Yes |
| `POST` | `/api/v1/users` | Create user | Yes | Yes |
| `GET` | `/api/v1/users/:id` | Get user by ID | Yes | No* |
| `PUT` | `/api/v1/users/:id` | Update user | Yes | No* |
| `DELETE` | `/api/v1/users/:id` | Delete user | Yes | Yes |

*Users can view/update their own profile

#### POST /api/v1/users

**Request:**
```json
{
  "username": "jane.smith",
  "email": "jane.smith@example.com",
  "password": "securePassword123",
  "firstName": "Jane",
  "lastName": "Smith",
  "roles": ["user", "energy_analyst"]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439014",
    "username": "jane.smith",
    "email": "jane.smith@example.com",
    "firstName": "Jane",
    "lastName": "Smith",
    "roles": ["user", "energy_analyst"],
    "isActive": true,
    "createdAt": "2024-01-20T10:00:00Z",
    "updatedAt": "2024-01-20T10:00:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "username": "jane.smith",
    "email": "jane.smith@example.com",
    "password": "securePassword123",
    "firstName": "Jane",
    "lastName": "Smith",
    "roles": ["user"]
  }'
```

---

### Role Management

| Method | Endpoint | Description | Auth Required | Admin Only |
|--------|----------|-------------|---------------|------------|
| `GET` | `/api/v1/roles` | List all roles | Yes | No |
| `POST` | `/api/v1/roles` | Create role | Yes | Yes |
| `PUT` | `/api/v1/roles/:roleName` | Update role | Yes | Yes |
| `DELETE` | `/api/v1/roles/:roleName` | Delete role | Yes | Yes |

#### POST /api/v1/roles

**Request:**
```json
{
  "name": "hvac_operator",
  "description": "Operator with HVAC control permissions",
  "permissions": [
    {
      "resource": "devices",
      "actions": ["read", "write"]
    },
    {
      "resource": "telemetry",
      "actions": ["read"]
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439015",
    "name": "hvac_operator",
    "description": "Operator with HVAC control permissions",
    "permissions": [...],
    "isSystem": false,
    "createdAt": "2024-01-20T10:30:00Z",
    "updatedAt": "2024-01-20T10:30:00Z"
  }
}
```

---

### Audit Logging

| Method | Endpoint | Description | Auth Required | Admin Only |
|--------|----------|-------------|---------------|------------|
| `POST` | `/api/v1/audit/log` | Create audit log (internal) | No | No |
| `GET` | `/api/v1/audit/logs` | List audit logs | Yes | Yes |
| `GET` | `/api/v1/audit/logs/:id` | Get audit log by ID | Yes | Yes |

#### POST /api/v1/audit/log

Used by other microservices to log actions.

**Request:**
```json
{
  "userId": "user123",
  "username": "john.doe",
  "service": "iot-control-service",
  "action": "SEND_COMMAND",
  "resource": "device",
  "resourceId": "device-001",
  "details": {"command": "SET_TEMPERATURE", "value": 22},
  "ipAddress": "192.168.1.100",
  "userAgent": "IoT-Control-Service/1.0",
  "status": "SUCCESS",
  "requestPath": "/api/v1/iot/device-control/device-001/command",
  "method": "POST"
}
```

#### GET /api/v1/audit/logs

**Query Parameters:**
- `from` - Start timestamp (ISO 8601)
- `to` - End timestamp (ISO 8601)
- `userId` - Filter by user ID
- `service` - Filter by service name
- `action` - Filter by action type
- `resource` - Filter by resource type
- `status` - Filter by status (SUCCESS, FAILURE)
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20)

**Example:**
```bash
curl -X GET "http://localhost:8080/api/v1/audit/logs?service=iot-control-service&status=SUCCESS&limit=50" \
  -H "Authorization: Bearer <admin_token>"
```

---

### Notifications

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/notifications/send` | Send notification | Yes |
| `POST` | `/api/v1/notifications/preferences` | Update preferences | Yes |
| `GET` | `/api/v1/notifications/logs` | Get notification logs | Yes |

---

### External Energy Integration

| Method | Endpoint | Description | Auth Required | Admin Only |
|--------|----------|-------------|---------------|------------|
| `GET` | `/api/v1/external-energy/consumption` | Get energy consumption | Yes | No |
| `GET` | `/api/v1/external-energy/tariffs` | Get current tariffs | Yes | No |
| `POST` | `/api/v1/external-energy/refresh-token` | Refresh provider token | Yes | Yes |

---

## Authentication & Authorization

### JWT Token Structure

**Access Token Claims:**
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

### Default Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| `admin` | Full system access | `*:*` (all resources, all actions) |
| `user` | Basic access | `read` on most resources |
| `building_manager` | Building management | `read/write` on buildings, devices |
| `energy_analyst` | Read-only energy data | `read` on energy, reports, analytics |

### Password Requirements
- Minimum 8 characters
- Hashed using bcrypt

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `GIN_MODE` | Gin framework mode | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `security_service` |
| `MONGODB_TIMEOUT` | Connection timeout (seconds) | `10` |
| `JWT_SECRET` | JWT signing secret | `default-secret-change-me` |
| `JWT_ACCESS_TOKEN_EXPIRY` | Access token expiry | `15m` |
| `JWT_REFRESH_TOKEN_EXPIRY` | Refresh token expiry | `168h` (7 days) |
| `ENCRYPTION_KEY` | Data encryption key (32 bytes) | `32-byte-encryption-key-here!!!!` |
| `NOTIFICATION_EMAIL_URL` | Email notification endpoint | `http://localhost:8081/external/notifications/email` |
| `NOTIFICATION_SMS_URL` | SMS notification endpoint | `http://localhost:8081/external/notifications/sms` |
| `NOTIFICATION_PUSH_URL` | Push notification endpoint | `http://localhost:8081/external/notifications/push` |
| `ENERGY_PROVIDER_BASE_URL` | Energy provider API URL | `https://api.energy-provider.com` |
| `ENERGY_PROVIDER_API_KEY` | Energy provider API key | - |
| `STORAGE_SERVICE_URL` | Storage service URL | `http://localhost:8086/storage` |
| `LOG_LEVEL` | Logging level | `debug` |
| `LOG_FORMAT` | Log format | `json` |

### Example .env File

```env
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=release

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=security_service
MONGODB_TIMEOUT=10

JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

ENCRYPTION_KEY=32-byte-encryption-key-here!!!!

LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Database Schema

### MongoDB Collections

| Collection | Description | Indexes |
|------------|-------------|---------|
| `users` | User accounts | `username` (unique), `email` (unique) |
| `roles` | Role definitions | `name` (unique) |
| `refresh_tokens` | Active refresh tokens | `token`, `user_id`, `expires_at` |
| `auth_credentials` | External service credentials | `service_name` |
| `audit_logs` | Audit log entries | `timestamp`, `user_id`, `service`, `action` |
| `notifications` | Notification history | `user_id`, `created_at` |
| `notification_preferences` | User notification settings | `user_id` |

---

## Inter-Service Communication

### Inbound (Other services calling Security Service)

| Caller | Endpoint | Purpose |
|--------|----------|---------|
| All services | `GET /api/v1/auth/validate-token` | Validate JWT tokens |
| All services | `POST /api/v1/auth/check-permissions` | Check user permissions |
| All services | `POST /api/v1/audit/log` | Log audit events |

### Outbound (Security Service calling external services)

| Target | Purpose |
|--------|---------|
| Storage Service | Persist credentials and audit logs |
| Notification Service | Send email/SMS/push notifications |
| Energy Provider API | Fetch consumption and tariff data |

---

## Running the Service

### Prerequisites

- Go 1.21+
- MongoDB 7.0+
- Docker (optional)

### Local Development

```bash
# Navigate to service directory
cd security-service

# Install dependencies
go mod download

# Set environment variables
export MONGODB_URI=mongodb://localhost:27017
export JWT_SECRET=your-secret-key-here

# Run the service
go run cmd/main.go
```

### Using Docker

```bash
# Build the image
docker build -t security-service ./security-service

# Run the container
docker run -p 8080:8080 \
  -e MONGODB_URI=mongodb://host.docker.internal:27017 \
  -e JWT_SECRET=your-secret-key \
  security-service
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d security-service

# View logs
docker-compose logs -f security-service
```

---

## Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "security-service"
}
```

---

## Example Usage

### Complete Authentication Flow

```bash
# 1. Login
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.refreshToken')

# 2. Access protected resource
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Refresh token when expired
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}"

# 4. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Create User with Role

```bash
# Create a new user with building_manager role
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "username": "manager1",
    "email": "manager1@company.com",
    "password": "securePass123!",
    "firstName": "Building",
    "lastName": "Manager",
    "roles": ["building_manager"]
  }'
```

---

## Known Limitations

1. **Single Database:** All data stored in single MongoDB instance; no sharding support
2. **No MFA:** Multi-factor authentication not implemented
3. **Token Revocation:** Access tokens cannot be revoked before expiry (use short expiry times)
4. **Password Reset:** No self-service password reset flow
5. **Rate Limiting:** No built-in rate limiting (implement at API gateway level)
6. **Session Management:** No active session listing/management

---

## Developer Notes

- Always use HTTPS in production
- Rotate JWT secrets periodically
- Monitor audit logs for suspicious activity
- Set appropriate token expiry times based on security requirements
- System roles (`isSystem: true`) cannot be deleted
- Password hashing uses bcrypt with default cost factor
