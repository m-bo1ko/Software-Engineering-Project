#!/bin/bash

# IoT & Control Component Service - API Test Script
# This script tests all endpoints of the IoT & Control Component microservice

# Configuration
BASE_URL="${IOT_SERVICE_URL:-http://localhost:8083}"
AUTH_TOKEN="${AUTH_TOKEN:-your-jwt-token-here}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print section headers
print_section() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Helper function to print test results
print_test() {
    echo -e "${YELLOW}Testing: $1${NC}"
    echo -e "${YELLOW}Endpoint: $2${NC}\n"
}

# Helper function to make authenticated requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    print_test "$description" "$method $endpoint"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "Content-Type: application/json" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ Status: $http_code${NC}"
        echo -e "${GREEN}Response:${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ Status: $http_code${NC}"
        echo -e "${RED}Response:${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    fi
    echo ""
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON responses will not be formatted.${NC}"
    echo -e "${YELLOW}Install jq for better output: sudo apt-get install jq${NC}\n"
fi

echo -e "${GREEN}IoT & Control Component Service - API Test Script${NC}"
echo -e "Base URL: $BASE_URL"
echo -e "Auth Token: ${AUTH_TOKEN:0:20}...\n"

# Health Check
print_section "Health Check"
make_request "GET" "/health" "" "Health Check"

# Device Management Endpoints
print_section "Device Management API"

# Register Device
REGISTER_DEVICE='{
  "deviceId": "device-test-001",
  "type": "HVAC",
  "model": "Model-X-Pro",
  "location": {
    "buildingId": "building-001",
    "floor": "3",
    "room": "301"
  },
  "capabilities": ["temperature", "humidity", "consumption"],
  "metadata": {
    "manufacturer": "ACME Corp",
    "firmware": "v2.1.0"
  }
}'
make_request "POST" "/iot/devices/register" "$REGISTER_DEVICE" "Register Device"

# Register Another Device
REGISTER_DEVICE_2='{
  "deviceId": "device-test-002",
  "type": "LIGHTING",
  "model": "Smart-LED-500",
  "location": {
    "buildingId": "building-001",
    "floor": "3",
    "room": "302"
  },
  "capabilities": ["brightness", "consumption"]
}'
make_request "POST" "/iot/devices/register" "$REGISTER_DEVICE_2" "Register Second Device"

# List All Devices
make_request "GET" "/iot/devices?page=1&limit=10" "" "List All Devices"

# List Devices by Building
make_request "GET" "/iot/devices?buildingId=building-001&page=1&limit=10" "" "List Devices by Building"

# List Devices by Type
make_request "GET" "/iot/devices?type=HVAC&page=1&limit=10" "" "List Devices by Type"

# List Devices by Status
make_request "GET" "/iot/devices?status=ONLINE&page=1&limit=10" "" "List Devices by Status"

# Get Device by ID
make_request "GET" "/iot/devices/device-test-001" "" "Get Device by ID"

# Telemetry Ingestion Endpoints
print_section "Telemetry Ingestion API"

# Ingest Single Telemetry
TELEMETRY_SINGLE='{
  "deviceId": "device-test-001",
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": {
    "temperature": 22.5,
    "humidity": 45.0,
    "consumption": 125.5
  }
}'
make_request "POST" "/iot/telemetry" "$TELEMETRY_SINGLE" "Ingest Single Telemetry"

# Ingest Bulk Telemetry
TELEMETRY_BULK='{
  "telemetry": [
    {
      "deviceId": "device-test-001",
      "timestamp": "2024-01-15T10:31:00Z",
      "metrics": {
        "temperature": 22.7,
        "humidity": 45.2,
        "consumption": 126.0
      }
    },
    {
      "deviceId": "device-test-002",
      "timestamp": "2024-01-15T10:31:00Z",
      "metrics": {
        "brightness": 80,
        "consumption": 15.5
      }
    },
    {
      "deviceId": "device-test-001",
      "timestamp": "2024-01-15T10:32:00Z",
      "metrics": {
        "temperature": 22.8,
        "humidity": 45.3,
        "consumption": 127.2
      }
    }
  ]
}'
make_request "POST" "/iot/telemetry/bulk" "$TELEMETRY_BULK" "Ingest Bulk Telemetry"

# Get Telemetry History
FROM_DATE=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v-7d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "2024-01-08T00:00:00Z")
TO_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "2024-01-15T23:59:59Z")
make_request "GET" "/iot/telemetry/history?deviceId=device-test-001&from=$FROM_DATE&to=$TO_DATE&page=1&limit=20" "" "Get Telemetry History"

# Device Control Endpoints
print_section "Device Control API"

# Send Command to Device
SEND_COMMAND='{
  "command": "SET_TEMPERATURE",
  "params": {
    "temperature": 20.0,
    "mode": "AUTO"
  }
}'
make_request "POST" "/iot/device-control/device-test-001/command" "$SEND_COMMAND" "Send Command to Device"

# Send Another Command
SEND_COMMAND_2='{
  "command": "SET_BRIGHTNESS",
  "params": {
    "brightness": 75,
    "duration": 3600
  }
}'
make_request "POST" "/iot/device-control/device-test-002/command" "$SEND_COMMAND_2" "Send Command to Device (Brightness)"

# List Commands for Device
make_request "GET" "/iot/device-control/device-test-001/commands?page=1&limit=10" "" "List Commands for Device"

# List Commands by Status
make_request "GET" "/iot/device-control/device-test-001/commands?status=SENT&page=1&limit=10" "" "List Commands by Status"

# Optimization Scenario Endpoints
print_section "Optimization Scenario API"

# Apply Optimization Scenario
APPLY_OPTIMIZATION='{
  "scenarioId": "opt-scenario-001",
  "forecastId": "forecast-001",
  "buildingId": "building-001",
  "actions": [
    {
      "deviceId": "device-test-001",
      "command": "SET_TEMPERATURE",
      "params": {
        "temperature": 21.0
      },
      "priority": 1
    },
    {
      "deviceId": "device-test-002",
      "command": "SET_BRIGHTNESS",
      "params": {
        "brightness": 70
      },
      "priority": 2
    }
  ]
}'
make_request "POST" "/iot/optimization/apply" "$APPLY_OPTIMIZATION" "Apply Optimization Scenario"

# Wait for scenario processing
sleep 2

# Get Optimization Status
make_request "GET" "/iot/optimization/status/opt-scenario-001" "" "Get Optimization Scenario Status"

# Real-Time State Endpoints
print_section "Real-Time State API"

# Get Live State (All Devices)
make_request "GET" "/iot/state/live" "" "Get Live State (All Devices)"

# Get Device State
make_request "GET" "/iot/state/device-test-001" "" "Get Device State"

# Get Device State (Non-existent)
make_request "GET" "/iot/state/non-existent-device" "" "Get Device State (Non-existent)"

# API v1 Routes (Testing legacy compatibility)
print_section "API v1 Routes (Legacy Compatibility)"

make_request "GET" "/api/v1/iot/devices?page=1&limit=5" "" "List Devices (API v1)"
make_request "GET" "/api/v1/iot/state/device-test-001" "" "Get Device State (API v1)"
make_request "POST" "/api/v1/iot/telemetry" "$TELEMETRY_SINGLE" "Ingest Telemetry (API v1)"

# Error Cases
print_section "Error Cases Testing"

# Invalid device registration (missing required fields)
INVALID_DEVICE='{
  "deviceId": "",
  "type": "HVAC"
}'
make_request "POST" "/iot/devices/register" "$INVALID_DEVICE" "Register Device (Invalid - Missing Device ID)"

# Invalid telemetry (missing device ID)
INVALID_TELEMETRY='{
  "metrics": {
    "temperature": 22.5
  }
}'
make_request "POST" "/iot/telemetry" "$INVALID_TELEMETRY" "Ingest Telemetry (Invalid - Missing Device ID)"

# Invalid command (missing command)
INVALID_COMMAND='{
  "params": {
    "temperature": 20.0
  }
}'
make_request "POST" "/iot/device-control/device-test-001/command" "$INVALID_COMMAND" "Send Command (Invalid - Missing Command)"

# Invalid optimization (missing actions)
INVALID_OPTIMIZATION='{
  "scenarioId": "opt-invalid",
  "buildingId": "building-001"
}'
make_request "POST" "/iot/optimization/apply" "$INVALID_OPTIMIZATION" "Apply Optimization (Invalid - Missing Actions)"

# Unauthorized request (without token)
print_test "Unauthorized Request Test" "GET /iot/devices (No Token)"
unauth_response=$(curl -s -w "\n%{http_code}" -X "GET" \
    -H "Content-Type: application/json" \
    "$BASE_URL/iot/devices")
unauth_code=$(echo "$unauth_response" | tail -n1)
if [ "$unauth_code" -eq 401 ]; then
    echo -e "${GREEN}✓ Correctly returned 401 Unauthorized${NC}"
else
    echo -e "${RED}✗ Expected 401, got $unauth_code${NC}"
fi
echo ""

# MQTT Topic Examples
print_section "MQTT Topic Examples"
echo -e "${YELLOW}Note: These are examples of MQTT messages that devices would publish/subscribe to:${NC}\n"

echo -e "${BLUE}Telemetry Topic (Device → Service):${NC}"
echo -e "Topic: mqtt/iot/device-test-001/telemetry"
echo -e "Message:"
echo '{
  "deviceId": "device-test-001",
  "timestamp": "2024-01-15T10:35:00Z",
  "metrics": {
    "temperature": 23.0,
    "humidity": 46.0,
    "consumption": 128.5
  }
}' | jq '.' 2>/dev/null || cat
echo ""

echo -e "${BLUE}Command Topic (Service → Device):${NC}"
echo -e "Topic: mqtt/iot/device-test-001/command"
echo -e "Message:"
echo '{
  "commandId": "cmd-123",
  "deviceId": "device-test-001",
  "command": "SET_TEMPERATURE",
  "params": {
    "temperature": 20.0
  }
}' | jq '.' 2>/dev/null || cat
echo ""

echo -e "${BLUE}Command Acknowledgment Topic (Device → Service):${NC}"
echo -e "Topic: mqtt/iot/device-test-001/ack"
echo -e "Message:"
echo '{
  "commandId": "cmd-123",
  "deviceId": "device-test-001",
  "status": "APPLIED",
  "timestamp": "2024-01-15T10:35:05Z"
}' | jq '.' 2>/dev/null || cat
echo ""

echo -e "${BLUE}Broadcast Topic (Service → All Devices):${NC}"
echo -e "Topic: mqtt/iot/broadcast/announcement"
echo -e "Message:"
echo '{
  "type": "SYSTEM_UPDATE",
  "message": "Scheduled maintenance at 2 AM",
  "timestamp": "2024-01-15T10:35:00Z"
}' | jq '.' 2>/dev/null || cat
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Testing Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Usage Tips:${NC}"
echo -e "1. Set AUTH_TOKEN environment variable: export AUTH_TOKEN='your-token'"
echo -e "2. Set IOT_SERVICE_URL if service is not on localhost:8083"
echo -e "3. Install jq for formatted JSON output: sudo apt-get install jq"
echo -e "4. Some endpoints require actual IDs from previous requests"
echo -e "5. For MQTT testing, use an MQTT client like mosquitto_pub/mosquitto_sub\n"
