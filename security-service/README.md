# Security & External Integration Microservice

A comprehensive security microservice for EMSIB (Enterprise Microservices Infrastructure for Buildings) that provides centralized authentication, authorization, token management, audit logging, user/role management, notification delivery, and external API integration.

## Features

- **Authentication**: JWT-based login, logout, and token refresh
- **Authorization**: Role-based access control (RBAC) with fine-grained permissions
- **User Management**: CRUD operations for user accounts
- **Role Management**: Create and manage roles with specific permissions
- **Audit Logging**: Track all critical actions across microservices
- **Notifications**: Send email, SMS, and push notifications
- **External Energy Integration**: Connect to external energy providers for consumption and tariff data

## Technology Stack

- **Language**: Go (Golang) 1.21+
- **Framework**: Gin
- **Database**: MongoDB
- **Authentication**: JWT (JSON Web Tokens)
- **Containerization**: Docker

## Project Structure

```
security-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── handlers/
│   │   ├── auth_handler.go     # Authentication endpoints
│   │   ├── user_handler.go     # User management endpoints
│   │   ├── role_handler.go     # Role management endpoints
│   │   ├── audit_handler.go    # Audit logging endpoints
│   │   ├── notification_handler.go
│   │   ├── energy_handler.go   # External energy integration
│   │   └── router.go           # Route configuration
│   ├── middleware/
│   │   ├── auth.go             # JWT authentication middleware
│   │   └── common.go           # Common middleware (CORS, logging, etc.)
│   ├── models/                 # Data models
│   ├── repository/             # Database operations
│   ├── service/                # Business logic
│   └── integrations/           # External service clients
├── pkg/
│   └── utils/
│       ├── jwt.go              # JWT utilities
│       ├── password.go         # Password hashing
│       └── encryption.go       # Data encryption
├── tests/                      # Test files
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── .env.example
```

## Quick Start

### Prerequisites

- Go 1.21+
- MongoDB 7.0+
- Docker & Docker Compose (optional)

### Running with Docker

```bash
# Clone the repository
cd security-service

# Copy and configure environment
cp .env.example .env
# Edit .env with your configuration

# Start services
docker-compose up -d

# View logs
docker-compose logs -f security-service
```

### Running Locally

```bash
# Install dependencies
go mod download

# Set environment variables (or create .env file)
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=security_service
export JWT_SECRET=your-secret-key

# Run the application
go run cmd/main.go
```

### Running Tests

```bash
go test ./tests/... -v
```

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/login` | User login | No |
| POST | `/auth/refresh` | Refresh access token | No |
| POST | `/auth/logout` | User logout | Yes |
| GET | `/auth/validate-token` | Validate token (internal) | No |
| POST | `/auth/check-permissions` | Check user permissions | No |
| GET | `/auth/user-info` | Get current user info | Yes |

### User Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/users` | List all users | Yes (Admin) |
| GET | `/users/{id}` | Get user by ID | Yes |
| POST | `/users` | Create new user | Yes (Admin) |
| PUT | `/users/{id}` | Update user | Yes |
| DELETE | `/users/{id}` | Delete user | Yes (Admin) |

### Role Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/roles` | List all roles | Yes |
| POST | `/roles` | Create new role | Yes (Admin) |
| PUT | `/roles/{roleName}` | Update role | Yes (Admin) |
| DELETE | `/roles/{roleName}` | Delete role | Yes (Admin) |

### Notifications

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/notifications/send` | Send notification | Yes |
| POST | `/notifications/preferences` | Update preferences | Yes |
| GET | `/notifications/logs` | Get notification history | Yes |

### Audit Logging

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/audit/log` | Create audit log | No (Internal) |
| GET | `/audit/logs` | Get audit logs | Yes (Admin) |

### External Energy Integration

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/external-energy/consumption` | Get energy consumption | Yes |
| GET | `/external-energy/tariffs` | Get tariff info | Yes |
| POST | `/external-energy/refresh-token` | Refresh external token | Yes (Admin) |

## Example curl Commands

### Authentication

#### Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "tokenType": "Bearer",
    "expiresIn": 900,
    "roles": ["admin"],
    "userId": "507f1f77bcf86cd799439011"
  }
}
```

#### Refresh Token
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

#### Logout
```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

#### Validate Token (Internal)
```bash
curl -X GET http://localhost:8080/auth/validate-token \
  -H "Authorization: Bearer <access_token>"
```

#### Check Permissions
```bash
curl -X POST http://localhost:8080/auth/check-permissions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "507f1f77bcf86cd799439011",
    "resource": "buildings",
    "action": "write"
  }'
```

#### Get User Info
```bash
curl -X GET http://localhost:8080/auth/user-info \
  -H "Authorization: Bearer <access_token>"
```

### User Management

#### List Users (Admin only)
```bash
curl -X GET "http://localhost:8080/users?page=1&limit=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

#### Get User by ID
```bash
curl -X GET http://localhost:8080/users/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <access_token>"
```

#### Create User (Admin only)
```bash
curl -X POST http://localhost:8080/users \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "securePassword123",
    "firstName": "John",
    "lastName": "Doe",
    "roles": ["user", "building_manager"]
  }'
```

#### Update User
```bash
curl -X PUT http://localhost:8080/users/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "Jane",
    "lastName": "Smith",
    "email": "jane.smith@example.com"
  }'
```

#### Delete User (Admin only)
```bash
curl -X DELETE http://localhost:8080/users/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <admin_access_token>"
```

### Role Management

#### List Roles
```bash
curl -X GET http://localhost:8080/roles \
  -H "Authorization: Bearer <access_token>"
```

#### Create Role (Admin only)
```bash
curl -X POST http://localhost:8080/roles \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "custom_analyst",
    "description": "Custom analyst role",
    "permissions": [
      {"resource": "reports", "actions": ["read", "write"]},
      {"resource": "energy", "actions": ["read"]}
    ]
  }'
```

#### Update Role (Admin only)
```bash
curl -X PUT http://localhost:8080/roles/custom_analyst \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "permissions": [
      {"resource": "reports", "actions": ["read", "write", "delete"]},
      {"resource": "energy", "actions": ["read"]}
    ]
  }'
```

#### Delete Role (Admin only)
```bash
curl -X DELETE http://localhost:8080/roles/custom_analyst \
  -H "Authorization: Bearer <admin_access_token>"
```

### Notifications

#### Send Notification
```bash
curl -X POST http://localhost:8080/notifications/send \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "507f1f77bcf86cd799439011",
    "type": "email",
    "subject": "Energy Alert",
    "content": "High energy consumption detected in Building A",
    "recipient": "user@example.com"
  }'
```

#### Update Notification Preferences
```bash
curl -X POST http://localhost:8080/notifications/preferences \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "507f1f77bcf86cd799439011",
    "emailEnabled": true,
    "smsEnabled": false,
    "pushEnabled": true,
    "quietHoursEnabled": true,
    "quietHoursStart": "22:00",
    "quietHoursEnd": "08:00"
  }'
```

#### Get Notification Logs
```bash
curl -X GET "http://localhost:8080/notifications/logs?userId=507f1f77bcf86cd799439011&page=1&limit=20" \
  -H "Authorization: Bearer <access_token>"
```

### Audit Logging

#### Create Audit Log (Internal)
```bash
curl -X POST http://localhost:8080/audit/log \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "507f1f77bcf86cd799439011",
    "username": "admin",
    "service": "building-service",
    "action": "CREATE_BUILDING",
    "resource": "building",
    "resourceId": "building123",
    "status": "SUCCESS",
    "details": {"name": "Main Office"},
    "ipAddress": "192.168.1.100",
    "requestPath": "/buildings",
    "method": "POST"
  }'
```

#### Get Audit Logs (Admin only)
```bash
curl -X GET "http://localhost:8080/audit/logs?from=2024-01-01T00:00:00Z&to=2024-12-31T23:59:59Z&userId=507f1f77bcf86cd799439011&service=security-service&page=1&limit=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

### External Energy Integration

#### Get Energy Consumption
```bash
curl -X GET "http://localhost:8080/external-energy/consumption?buildingId=building123&from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z" \
  -H "Authorization: Bearer <access_token>"
```

#### Get Tariffs
```bash
curl -X GET "http://localhost:8080/external-energy/tariffs?region=northeast" \
  -H "Authorization: Bearer <access_token>"
```

#### Refresh External Token (Admin only)
```bash
curl -X POST http://localhost:8080/external-energy/refresh-token \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "energy_provider"
  }'
```

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

## Configuration

The application can be configured via environment variables or a `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8080` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `security_service` |
| `JWT_SECRET` | JWT signing secret | Required |
| `JWT_ACCESS_TOKEN_EXPIRY` | Access token expiry | `15m` |
| `JWT_REFRESH_TOKEN_EXPIRY` | Refresh token expiry | `168h` |
| `ENCRYPTION_KEY` | Encryption key (32 bytes) | Required |

## Default Roles

The system initializes with these default roles:

| Role | Description |
|------|-------------|
| `admin` | Full access to all resources |
| `user` | Basic user access |
| `building_manager` | Building and energy data access |
| `energy_analyst` | Read access to energy data |

## Default Admin User

On first startup, a default admin user is created:
- **Username**: `admin`
- **Password**: `admin123`

**Important**: Change the default password immediately in production!

## Security Considerations

1. Always use HTTPS in production
2. Change default credentials immediately
3. Use strong JWT secrets (256+ bits)
4. Enable rate limiting at the API gateway level
5. Rotate encryption keys periodically
6. Monitor audit logs for suspicious activity

## License

This project is part of the EMSIB microservices infrastructure.
