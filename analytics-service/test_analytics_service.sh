#!/bin/bash

# Analytics Component Service - API Test Script
# This script tests all endpoints of the Analytics Component microservice

# Configuration
BASE_URL="${ANALYTICS_SERVICE_URL:-http://localhost:8084}"
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

echo -e "${GREEN}Analytics Component Service - API Test Script${NC}"
echo -e "Base URL: $BASE_URL"
echo -e "Auth Token: ${AUTH_TOKEN:0:20}...\n"

# Health Check
print_section "Health Check"
make_request "GET" "/health" "" "Health Check"

# Reports Endpoints
print_section "Reports API"

# List Reports
make_request "GET" "/analytics/reports?page=1&limit=10" "" "List Reports"

# Generate Energy Consumption Report
ENERGY_REPORT_DATA='{
  "type": "ENERGY_CONSUMPTION",
  "buildingId": "building-001",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z",
  "options": {
    "includeForecast": true
  }
}'
make_request "POST" "/analytics/reports/generate" "$ENERGY_REPORT_DATA" "Generate Energy Consumption Report"

# Wait a moment for report generation
sleep 2

# Generate Device Performance Report
PERF_REPORT_DATA='{
  "type": "DEVICE_PERFORMANCE",
  "buildingId": "building-001",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z"
}'
make_request "POST" "/analytics/reports/generate" "$PERF_REPORT_DATA" "Generate Device Performance Report"

# List Reports with filters
make_request "GET" "/analytics/reports?buildingId=building-001&type=ENERGY_CONSUMPTION&status=COMPLETED" "" "List Reports (Filtered)"

# Get Report by ID (using a placeholder - replace with actual report ID)
# REPORT_ID="your-report-id-here"
# make_request "GET" "/analytics/reports/$REPORT_ID" "" "Get Report by ID"

# Anomalies Endpoints
print_section "Anomalies API"

# List Anomalies
make_request "GET" "/analytics/anomalies?page=1&limit=10" "" "List All Anomalies"

# List Anomalies with filters
make_request "GET" "/analytics/anomalies?status=NEW&severity=HIGH&page=1&limit=10" "" "List Anomalies (Filtered by Status and Severity)"

# List Anomalies by Building
make_request "GET" "/analytics/anomalies?buildingId=building-001&status=NEW" "" "List Anomalies by Building"

# Acknowledge Anomaly (using placeholder - replace with actual anomaly ID)
# ANOMALY_ID="your-anomaly-id-here"
# ACK_DATA="{\"anomalyId\": \"$ANOMALY_ID\"}"
# make_request "POST" "/analytics/anomalies/acknowledge" "$ACK_DATA" "Acknowledge Anomaly"

# Get Anomaly by ID (using placeholder)
# make_request "GET" "/analytics/anomalies/$ANOMALY_ID" "" "Get Anomaly by ID"

# Time-Series Endpoints
print_section "Time-Series API"

# Query Time-Series Data (Hourly)
TIMESERIES_HOURLY='{
  "deviceIds": ["device-001", "device-002"],
  "buildingId": "building-001",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z",
  "aggregationType": "HOURLY",
  "metrics": ["temperature", "consumption"]
}'
make_request "POST" "/analytics/time-series/query" "$TIMESERIES_HOURLY" "Query Time-Series (Hourly Aggregation)"

# Query Time-Series Data (Daily)
TIMESERIES_DAILY='{
  "deviceIds": ["device-001"],
  "buildingId": "building-001",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z",
  "aggregationType": "DAILY"
}'
make_request "POST" "/analytics/time-series/query" "$TIMESERIES_DAILY" "Query Time-Series (Daily Aggregation)"

# Query Time-Series Data (Weekly)
TIMESERIES_WEEKLY='{
  "buildingId": "building-001",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z",
  "aggregationType": "WEEKLY"
}'
make_request "POST" "/analytics/time-series/query" "$TIMESERIES_WEEKLY" "Query Time-Series (Weekly Aggregation)"

# Query Time-Series Data (Monthly)
TIMESERIES_MONTHLY='{
  "buildingId": "building-001",
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-12-31T23:59:59Z",
  "aggregationType": "MONTHLY"
}'
make_request "POST" "/analytics/time-series/query" "$TIMESERIES_MONTHLY" "Query Time-Series (Monthly Aggregation)"

# KPI Endpoints
print_section "KPI API"

# Get System-Wide KPIs
make_request "GET" "/analytics/kpi?period=DAILY" "" "Get System-Wide KPIs (Daily)"

# Get Building KPIs
make_request "GET" "/analytics/kpi/building-001?period=DAILY" "" "Get Building KPIs (Daily)"

# Get Building KPIs (Weekly)
make_request "GET" "/analytics/kpi/building-001?period=WEEKLY" "" "Get Building KPIs (Weekly)"

# Calculate KPIs
make_request "POST" "/analytics/kpi/calculate?buildingId=building-001&period=DAILY" "" "Calculate KPIs for Building"

# Calculate System-Wide KPIs
make_request "POST" "/analytics/kpi/calculate?period=DAILY" "" "Calculate System-Wide KPIs"

# Dashboard Endpoints
print_section "Dashboard API"

# Get Overview Dashboard
make_request "GET" "/analytics/dashboards/overview" "" "Get System Overview Dashboard"

# Get Building Dashboard
make_request "GET" "/analytics/dashboards/building/building-001" "" "Get Building Dashboard"

# API v1 Routes (Testing legacy compatibility)
print_section "API v1 Routes (Legacy Compatibility)"

make_request "GET" "/api/v1/analytics/reports?page=1&limit=5" "" "List Reports (API v1)"
make_request "GET" "/api/v1/analytics/anomalies?status=NEW" "" "List Anomalies (API v1)"
make_request "GET" "/api/v1/analytics/kpi?period=DAILY" "" "Get KPIs (API v1)"
make_request "GET" "/api/v1/analytics/dashboards/overview" "" "Get Overview Dashboard (API v1)"

# Error Cases
print_section "Error Cases Testing"

# Invalid request (missing required field)
INVALID_REPORT='{
  "type": ""
}'
make_request "POST" "/analytics/reports/generate" "$INVALID_REPORT" "Generate Report (Invalid - Missing Type)"

# Invalid time-series query (missing aggregation type)
INVALID_TIMESERIES='{
  "from": "2024-01-01T00:00:00Z",
  "to": "2024-01-31T23:59:59Z"
}'
make_request "POST" "/analytics/time-series/query" "$INVALID_TIMESERIES" "Query Time-Series (Invalid - Missing Aggregation Type)"

# Non-existent report
make_request "GET" "/analytics/reports/non-existent-id" "" "Get Non-Existent Report"

# Unauthorized request (without token)
print_test "Unauthorized Request Test" "GET /analytics/reports (No Token)"
unauth_response=$(curl -s -w "\n%{http_code}" -X "GET" \
    -H "Content-Type: application/json" \
    "$BASE_URL/analytics/reports")
unauth_code=$(echo "$unauth_response" | tail -n1)
if [ "$unauth_code" -eq 401 ]; then
    echo -e "${GREEN}✓ Correctly returned 401 Unauthorized${NC}"
else
    echo -e "${RED}✗ Expected 401, got $unauth_code${NC}"
fi
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Testing Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${YELLOW}Note: Some endpoints require actual IDs from previous requests.${NC}"
echo -e "${YELLOW}Replace placeholders (REPORT_ID, ANOMALY_ID) with actual values to test those endpoints.${NC}\n"
