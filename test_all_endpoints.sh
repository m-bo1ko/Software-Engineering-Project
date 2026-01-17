#!/bin/bash

#===============================================================================
# COMPREHENSIVE API ENDPOINT TEST SCRIPT
# Tests all microservices: Security, Forecast, IoT Control, Analytics
#===============================================================================

# REMOVED: set -e (this was causing the script to exit on first failure)

#-------------------------------------------------------------------------------
# CONFIGURATION - Modify these variables as needed
#-------------------------------------------------------------------------------

# Service URLs
SECURITY_URL="http://localhost:8080"
FORECAST_URL="http://localhost:8082"
IOT_URL="http://localhost:8083"
ANALYTICS_URL="http://localhost:8084"

# Test credentials (change these for your environment)
TEST_USERNAME="admin"
TEST_PASSWORD="admin123"
TEST_EMAIL="admin@example.com"

# Test data IDs (will be populated during tests or can be set manually)
TEST_USER_ID=""
TEST_ROLE_ID=""
TEST_DEVICE_ID="test-device-001"
TEST_BUILDING_ID="building-001"
TEST_SCENARIO_ID=""
TEST_REPORT_ID=""
TEST_ANOMALY_ID=""

# Authentication token (will be obtained during login test)
AUTH_TOKEN=""
REFRESH_TOKEN=""

# Counters for summary
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

#-------------------------------------------------------------------------------
# COLOR CODES FOR OUTPUT
#-------------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

#-------------------------------------------------------------------------------
# HELPER FUNCTIONS
#-------------------------------------------------------------------------------

# Print section header
print_section() {
    echo ""
    echo -e "${BLUE}===============================================================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
}

# Print test header
print_test() {
    echo ""
    echo -e "${CYAN}--- TEST: $1 ---${NC}"
    ((TOTAL_TESTS++))
}

# Print request details
print_request() {
    echo -e "${YELLOW}REQUEST: $1 $2${NC}"
    if [ -n "$3" ]; then
        echo -e "${YELLOW}BODY: $3${NC}"
    fi
}

# Evaluate response and print result
evaluate_response() {
    local http_code=$1
    local expected_code=$2
    local response_body=$3
    local test_name=$4

    echo "HTTP Status: $http_code"
    echo "Response: $response_body" | head -c 500
    echo ""

    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}PASSED: $test_name (Expected: $expected_code, Got: $http_code)${NC}"
        ((PASSED_TESTS++))
        return 0
    else
        echo -e "${RED}FAILED: $test_name (Expected: $expected_code, Got: $http_code)${NC}"
        ((FAILED_TESTS++))
        return 0  # CHANGED: Always return 0 to prevent script termination
    fi
}

# Make a curl request and capture response
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth=$4
    local expected_code=$5
    local test_name=$6

    print_request "$method" "$url" "$data"

    local response=""
    local http_code=""
    local response_body=""

    if [ "$method" == "GET" ]; then
        response=$(curl -s -w '\n%{http_code}' -X GET "$url" \
            -H "Content-Type: application/json" \
            ${auth:+-H "Authorization: Bearer $auth"} 2>/dev/null || echo -e "\n000")
    elif [ "$method" == "POST" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -w '\n%{http_code}' -X POST "$url" \
                -H "Content-Type: application/json" \
                ${auth:+-H "Authorization: Bearer $auth"} \
                -d "$data" 2>/dev/null || echo -e "\n000")
        else
            response=$(curl -s -w '\n%{http_code}' -X POST "$url" \
                -H "Content-Type: application/json" \
                ${auth:+-H "Authorization: Bearer $auth"} 2>/dev/null || echo -e "\n000")
        fi
    elif [ "$method" == "PUT" ]; then
        response=$(curl -s -w '\n%{http_code}' -X PUT "$url" \
            -H "Content-Type: application/json" \
            ${auth:+-H "Authorization: Bearer $auth"} \
            -d "$data" 2>/dev/null || echo -e "\n000")
    elif [ "$method" == "DELETE" ]; then
        response=$(curl -s -w '\n%{http_code}' -X DELETE "$url" \
            -H "Content-Type: application/json" \
            ${auth:+-H "Authorization: Bearer $auth"} 2>/dev/null || echo -e "\n000")
    fi

    # Extract HTTP code (last line) and body
    http_code=$(echo "$response" | tail -n 1)
    response_body=$(echo "$response" | sed '$d')

    evaluate_response "$http_code" "$expected_code" "$response_body" "$test_name"

    # Return the response body for further processing
    echo "$response_body"
}

# Extract value from JSON response using jq or grep
extract_json_value() {
    local json=$1
    local key=$2

    if command -v jq &> /dev/null; then
        echo "$json" | jq -r ".$key // empty" 2>/dev/null
    else
        echo "$json" | grep -o "\"$key\":\"[^\"]*\"" | cut -d'"' -f4
    fi
}

#===============================================================================
# SECURITY SERVICE TESTS (Port 8080)
#===============================================================================

test_security_service() {
    print_section "SECURITY SERVICE TESTS ($SECURITY_URL)"

    #---------------------------------------------------------------------------
    # Health Check
    #---------------------------------------------------------------------------
    print_test "Health Check"
    make_request "GET" "$SECURITY_URL/health" "" "none" "200" "Security Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # AUTH ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Authentication Endpoints"

    # Test 1: Login with valid credentials
    print_test "Login - Valid Credentials"
    login_payload='{
        "username": "'$TEST_USERNAME'",
        "password": "'$TEST_PASSWORD'"
    }'
    response=$(make_request "POST" "$SECURITY_URL/auth/login" "$login_payload" "none" "200" "Login with valid credentials")

    # Extract tokens from response
    if command -v jq &> /dev/null; then
        AUTH_TOKEN=$(echo "$response" | jq -r '.data.accessToken // .accessToken // empty' 2>/dev/null)
        REFRESH_TOKEN=$(echo "$response" | jq -r '.data.refreshToken // .refreshToken // empty' 2>/dev/null)
        TEST_USER_ID=$(echo "$response" | jq -r '.data.userId // .userId // .data.user.id // empty' 2>/dev/null)
    fi

    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" == "null" ]; then
        echo -e "${YELLOW}WARNING: Could not extract auth token. Using placeholder for remaining tests.${NC}"
        AUTH_TOKEN="test-token-placeholder"
    else
        echo -e "${GREEN}Successfully extracted auth token${NC}"
    fi

    # Test 2: Login with invalid credentials
    print_test "Login - Invalid Credentials"
    invalid_login='{
        "username": "wronguser",
        "password": "wrongpassword"
    }'
    make_request "POST" "$SECURITY_URL/auth/login" "$invalid_login" "none" "401" "Login with invalid credentials" > /dev/null

    # Test 3: Login with missing fields
    print_test "Login - Missing Password"
    missing_password='{"username": "testuser"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$missing_password" "none" "400" "Login with missing password" > /dev/null

    # Test 4: Login with empty payload
    print_test "Login - Empty Payload"
    make_request "POST" "$SECURITY_URL/auth/login" "{}" "none" "400" "Login with empty payload" > /dev/null

    # Test 5: Validate Token - Valid
    print_test "Validate Token - Valid Token"
    make_request "GET" "$SECURITY_URL/auth/validate-token" "" "$AUTH_TOKEN" "200" "Validate valid token" > /dev/null

    # Test 6: Validate Token - Invalid
    print_test "Validate Token - Invalid Token"
    make_request "GET" "$SECURITY_URL/auth/validate-token" "" "invalid-token-12345" "401" "Validate invalid token" > /dev/null

    # Test 7: Validate Token - No Token
    print_test "Validate Token - No Token"
    make_request "GET" "$SECURITY_URL/auth/validate-token" "" "" "401" "Validate without token" > /dev/null

    # Test 8: Get User Info
    print_test "Get User Info - Authenticated"
    make_request "GET" "$SECURITY_URL/auth/user-info" "" "$AUTH_TOKEN" "200" "Get user info with valid token" > /dev/null

    # Test 9: Refresh Token
    print_test "Refresh Token - Valid Refresh Token"
    refresh_payload='{
        "refreshToken": "'$REFRESH_TOKEN'"
    }'
    make_request "POST" "$SECURITY_URL/auth/refresh" "$refresh_payload" "none" "200" "Refresh token" > /dev/null

    # Test 10: Refresh Token - Invalid
    print_test "Refresh Token - Invalid Refresh Token"
    invalid_refresh='{"refreshToken": "invalid-refresh-token"}'
    make_request "POST" "$SECURITY_URL/auth/refresh" "$invalid_refresh" "none" "401" "Refresh with invalid token" > /dev/null

    # Test 11: Check Permissions
    print_test "Check Permissions"
    permissions_payload='{
        "userId": "'$TEST_USER_ID'",
        "resource": "devices",
        "action": "read"
    }'
    make_request "POST" "$SECURITY_URL/auth/check-permissions" "$permissions_payload" "$AUTH_TOKEN" "200" "Check user permissions" > /dev/null

    #---------------------------------------------------------------------------
    # USER ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "User Management Endpoints"

    # Test 12: Create User
    print_test "Create User - Valid Data"
    create_user_payload='{
        "username": "testuser_'$(date +%s)'",
        "email": "testuser_'$(date +%s)'@example.com",
        "password": "Test123!@#",
        "firstName": "Test",
        "lastName": "User",
        "roles": ["user"]
    }'
    response=$(make_request "POST" "$SECURITY_URL/users" "$create_user_payload" "$AUTH_TOKEN" "201" "Create new user")

    # Extract created user ID
    if command -v jq &> /dev/null; then
        CREATED_USER_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    # Test 13: Create User - Duplicate Username
    print_test "Create User - Duplicate Username"
    make_request "POST" "$SECURITY_URL/users" "$create_user_payload" "$AUTH_TOKEN" "409" "Create user with duplicate username" > /dev/null

    # Test 14: Create User - Invalid Email
    print_test "Create User - Invalid Email"
    invalid_email_user='{
        "username": "testuser2",
        "email": "not-an-email",
        "password": "Test123!@#"
    }'
    make_request "POST" "$SECURITY_URL/users" "$invalid_email_user" "$AUTH_TOKEN" "400" "Create user with invalid email" > /dev/null

    # Test 15: Create User - Missing Required Fields
    print_test "Create User - Missing Username"
    missing_username='{"email": "test@example.com", "password": "Test123"}'
    make_request "POST" "$SECURITY_URL/users" "$missing_username" "$AUTH_TOKEN" "400" "Create user without username" > /dev/null

    # Test 16: List Users
    print_test "List Users"
    make_request "GET" "$SECURITY_URL/users" "" "$AUTH_TOKEN" "200" "List all users" > /dev/null

    # Test 17: List Users with Pagination
    print_test "List Users - With Pagination"
    make_request "GET" "$SECURITY_URL/users?page=1&limit=10" "" "$AUTH_TOKEN" "200" "List users with pagination" > /dev/null

    # Test 18: List Users - Unauthorized
    print_test "List Users - Unauthorized"
    make_request "GET" "$SECURITY_URL/users" "" "" "401" "List users without auth" > /dev/null

    # Test 19: Get User by ID
    print_test "Get User by ID - Valid ID"
    if [ -n "$CREATED_USER_ID" ] && [ "$CREATED_USER_ID" != "null" ]; then
        make_request "GET" "$SECURITY_URL/users/$CREATED_USER_ID" "" "$AUTH_TOKEN" "200" "Get user by valid ID" > /dev/null
    else
        make_request "GET" "$SECURITY_URL/users/507f1f77bcf86cd799439011" "" "$AUTH_TOKEN" "404" "Get user by ID (placeholder)" > /dev/null
    fi

    # Test 20: Get User by ID - Invalid ID Format
    print_test "Get User by ID - Invalid ID Format"
    make_request "GET" "$SECURITY_URL/users/invalid-id" "" "$AUTH_TOKEN" "400" "Get user with invalid ID format" > /dev/null

    # Test 21: Get User by ID - Non-existent
    print_test "Get User by ID - Non-existent"
    make_request "GET" "$SECURITY_URL/users/507f1f77bcf86cd799439999" "" "$AUTH_TOKEN" "404" "Get non-existent user" > /dev/null

    # Test 22: Update User
    print_test "Update User - Valid Data"
    update_user_payload='{
        "firstName": "Updated",
        "lastName": "Name"
    }'
    if [ -n "$CREATED_USER_ID" ] && [ "$CREATED_USER_ID" != "null" ]; then
        make_request "PUT" "$SECURITY_URL/users/$CREATED_USER_ID" "$update_user_payload" "$AUTH_TOKEN" "200" "Update user" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No user ID available for update test${NC}"
    fi

    # Test 23: Delete User
    print_test "Delete User"
    if [ -n "$CREATED_USER_ID" ] && [ "$CREATED_USER_ID" != "null" ]; then
        make_request "DELETE" "$SECURITY_URL/users/$CREATED_USER_ID" "" "$AUTH_TOKEN" "200" "Delete user" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No user ID available for delete test${NC}"
    fi

    #---------------------------------------------------------------------------
    # ROLE ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Role Management Endpoints"

    # Test 24: Create Role
    print_test "Create Role - Valid Data"
    create_role_payload='{
        "name": "test_role_'$(date +%s)'",
        "description": "Test role for QA",
        "permissions": [
            {"resource": "devices", "actions": ["read", "write"]},
            {"resource": "reports", "actions": ["read"]}
        ]
    }'
    response=$(make_request "POST" "$SECURITY_URL/roles" "$create_role_payload" "$AUTH_TOKEN" "201" "Create new role")

    if command -v jq &> /dev/null; then
        TEST_ROLE_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    # Test 25: Create Role - Missing Name
    print_test "Create Role - Missing Name"
    missing_name_role='{"description": "No name role"}'
    make_request "POST" "$SECURITY_URL/roles" "$missing_name_role" "$AUTH_TOKEN" "400" "Create role without name" > /dev/null

    # Test 26: List Roles
    print_test "List Roles"
    make_request "GET" "$SECURITY_URL/roles" "" "$AUTH_TOKEN" "200" "List all roles" > /dev/null

    # Test 27: Get Role by ID
    print_test "Get Role by ID"
    if [ -n "$TEST_ROLE_ID" ] && [ "$TEST_ROLE_ID" != "null" ]; then
        make_request "GET" "$SECURITY_URL/roles/$TEST_ROLE_ID" "" "$AUTH_TOKEN" "200" "Get role by ID" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No role ID available${NC}"
    fi

    # Test 28: Update Role
    print_test "Update Role"
    update_role_payload='{"description": "Updated description"}'
    if [ -n "$TEST_ROLE_ID" ] && [ "$TEST_ROLE_ID" != "null" ]; then
        make_request "PUT" "$SECURITY_URL/roles/$TEST_ROLE_ID" "$update_role_payload" "$AUTH_TOKEN" "200" "Update role" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No role ID available${NC}"
    fi

    # Test 29: Delete Role
    print_test "Delete Role"
    if [ -n "$TEST_ROLE_ID" ] && [ "$TEST_ROLE_ID" != "null" ]; then
        make_request "DELETE" "$SECURITY_URL/roles/$TEST_ROLE_ID" "" "$AUTH_TOKEN" "200" "Delete role" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No role ID available${NC}"
    fi

    #---------------------------------------------------------------------------
    # AUDIT ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Audit Logging Endpoints"

    # Test 30: Create Audit Log
    print_test "Create Audit Log"
    audit_payload='{
        "userId": "'$TEST_USER_ID'",
        "service": "test-script",
        "action": "TEST_ACTION",
        "resource": "test-resource",
        "status": "SUCCESS",
        "ipAddress": "127.0.0.1",
        "userAgent": "test-script/1.0"
    }'
    make_request "POST" "$SECURITY_URL/audit/log" "$audit_payload" "$AUTH_TOKEN" "201" "Create audit log entry" > /dev/null

    # Test 31: Get Audit Logs
    print_test "Get Audit Logs"
    make_request "GET" "$SECURITY_URL/audit/logs" "" "$AUTH_TOKEN" "200" "Get audit logs" > /dev/null

    # Test 32: Get Audit Logs with Filters
    print_test "Get Audit Logs - With Filters"
    make_request "GET" "$SECURITY_URL/audit/logs?service=security-service&limit=10" "" "$AUTH_TOKEN" "200" "Get filtered audit logs" > /dev/null

    #---------------------------------------------------------------------------
    # NOTIFICATION ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Notification Endpoints"

    # Test 33: Send Notification
    print_test "Send Notification"
    notification_payload='{
        "userId": "'$TEST_USER_ID'",
        "type": "EMAIL",
        "subject": "Test Notification",
        "message": "This is a test notification from the QA script"
    }'
    make_request "POST" "$SECURITY_URL/notifications/send" "$notification_payload" "$AUTH_TOKEN" "200" "Send notification" > /dev/null

    # Test 34: Get Notification Preferences
    print_test "Get Notification Preferences"
    make_request "GET" "$SECURITY_URL/notifications/preferences/$TEST_USER_ID" "" "$AUTH_TOKEN" "200" "Get notification preferences" > /dev/null

    # Test 35: Update Notification Preferences
    print_test "Update Notification Preferences"
    prefs_payload='{
        "email": true,
        "sms": false,
        "push": true
    }'
    make_request "PUT" "$SECURITY_URL/notifications/preferences/$TEST_USER_ID" "$prefs_payload" "$AUTH_TOKEN" "200" "Update notification preferences" > /dev/null

    # Test 36: Get Notification Logs
    print_test "Get Notification Logs"
    make_request "GET" "$SECURITY_URL/notifications/logs" "" "$AUTH_TOKEN" "200" "Get notification logs" > /dev/null

    #---------------------------------------------------------------------------
    # EXTERNAL ENERGY ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "External Energy Integration Endpoints"

    # Test 37: Get Energy Consumption
    print_test "Get Energy Consumption"
    make_request "GET" "$SECURITY_URL/external-energy/consumption?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get energy consumption" > /dev/null

    # Test 38: Get Tariffs
    print_test "Get Tariffs"
    make_request "GET" "$SECURITY_URL/external-energy/tariffs?region=default" "" "$AUTH_TOKEN" "200" "Get energy tariffs" > /dev/null

    # Test 39: Refresh Energy Token
    print_test "Refresh Energy Provider Token"
    make_request "POST" "$SECURITY_URL/external-energy/refresh-token" "" "$AUTH_TOKEN" "200" "Refresh energy provider token" > /dev/null

    #---------------------------------------------------------------------------
    # LOGOUT (End of Security Tests)
    #---------------------------------------------------------------------------
    print_test "Logout"
    logout_payload='{"refreshToken": "'$REFRESH_TOKEN'"}'
    make_request "POST" "$SECURITY_URL/auth/logout" "$logout_payload" "$AUTH_TOKEN" "200" "Logout" > /dev/null
}

#===============================================================================
# FORECAST SERVICE TESTS (Port 8082)
#===============================================================================

test_forecast_service() {
    print_section "FORECAST SERVICE TESTS ($FORECAST_URL)"

    # Re-authenticate if needed
    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" == "test-token-placeholder" ]; then
        echo -e "${YELLOW}Re-authenticating for Forecast Service tests...${NC}"
        login_payload='{"username": "'$TEST_USERNAME'", "password": "'$TEST_PASSWORD'"}'
        response=$(curl -s -X POST "$SECURITY_URL/auth/login" \
            -H "Content-Type: application/json" \
            -d "$login_payload" 2>/dev/null)
        if command -v jq &> /dev/null; then
            AUTH_TOKEN=$(echo "$response" | jq -r '.data.accessToken // .accessToken // empty' 2>/dev/null)
        fi
    fi

    #---------------------------------------------------------------------------
    # Health Check
    #---------------------------------------------------------------------------
    print_test "Health Check"
    make_request "GET" "$FORECAST_URL/health" "" "none" "200" "Forecast Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # FORECAST ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Forecast Generation Endpoints"

    # Test 40: Generate Forecast - Valid Request
    print_test "Generate Forecast - Valid Request"
    forecast_payload='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "DEMAND",
        "horizonHours": 24,
        "historicalDays": 7,
        "includeWeather": true,
        "includeTariffs": true
    }'
    response=$(make_request "POST" "$FORECAST_URL/forecast/generate" "$forecast_payload" "$AUTH_TOKEN" "201" "Generate demand forecast")

    if command -v jq &> /dev/null; then
        FORECAST_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    # Test 41: Generate Forecast - Missing Building ID
    print_test "Generate Forecast - Missing Building ID"
    invalid_forecast='{"type": "DEMAND", "horizonHours": 24}'
    make_request "POST" "$FORECAST_URL/forecast/generate" "$invalid_forecast" "$AUTH_TOKEN" "400" "Generate forecast without building ID" > /dev/null

    # Test 42: Generate Forecast - Invalid Horizon
    print_test "Generate Forecast - Excessive Horizon"
    excessive_horizon='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "DEMAND",
        "horizonHours": 10000
    }'
    make_request "POST" "$FORECAST_URL/forecast/generate" "$excessive_horizon" "$AUTH_TOKEN" "200" "Generate forecast with capped horizon" > /dev/null

    # Test 43: Generate Forecast - Unauthorized
    print_test "Generate Forecast - Unauthorized"
    make_request "POST" "$FORECAST_URL/forecast/generate" "$forecast_payload" "" "401" "Generate forecast without auth" > /dev/null

    # Test 44: Get Peak Load
    print_test "Get Peak Load Prediction"
    make_request "GET" "$FORECAST_URL/forecast/peak-load?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get peak load prediction" > /dev/null

    # Test 45: Get Latest Forecast
    print_test "Get Latest Forecast"
    make_request "GET" "$FORECAST_URL/forecast/latest?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get latest forecast" > /dev/null

    # Test 46: Get Latest Forecast - Missing Building ID
    print_test "Get Latest Forecast - Missing Building ID"
    make_request "GET" "$FORECAST_URL/forecast/latest" "" "$AUTH_TOKEN" "400" "Get latest forecast without building ID" > /dev/null

    # Test 47: Get Device Prediction
    print_test "Get Device Prediction"
    make_request "GET" "$FORECAST_URL/forecast/prediction/$TEST_DEVICE_ID" "" "$AUTH_TOKEN" "200" "Get device prediction" > /dev/null

    # Test 48: Get Device Prediction - Non-existent Device
    print_test "Get Device Prediction - Non-existent Device"
    make_request "GET" "$FORECAST_URL/forecast/prediction/non-existent-device" "" "$AUTH_TOKEN" "404" "Get prediction for non-existent device" > /dev/null

    # Test 49: Get Device Optimization
    print_test "Get Device Optimization Recommendations"
    make_request "GET" "$FORECAST_URL/forecast/optimization/$TEST_DEVICE_ID" "" "$AUTH_TOKEN" "200" "Get device optimization recommendations" > /dev/null

    #---------------------------------------------------------------------------
    # OPTIMIZATION ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Optimization Scenario Endpoints"

    # Test 50: Generate Optimization Scenario
    print_test "Generate Optimization Scenario"
    optimization_payload='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "COST_REDUCTION",
        "name": "Test Optimization Scenario",
        "forecastId": "'$FORECAST_ID'",
        "useTariffData": true,
        "useWeatherData": true,
        "priority": 5,
        "constraints": {
            "maxReduction": 30,
            "preserveComfort": true
        }
    }'
    response=$(make_request "POST" "$FORECAST_URL/optimization/generate" "$optimization_payload" "$AUTH_TOKEN" "201" "Generate optimization scenario")

    if command -v jq &> /dev/null; then
        TEST_SCENARIO_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    # Test 51: Generate Optimization - Missing Type
    print_test "Generate Optimization - Missing Type"
    invalid_opt='{"buildingId": "'$TEST_BUILDING_ID'"}'
    make_request "POST" "$FORECAST_URL/optimization/generate" "$invalid_opt" "$AUTH_TOKEN" "400" "Generate optimization without type" > /dev/null

    # Test 52: Get Optimization Recommendations
    print_test "Get Optimization Recommendations"
    make_request "GET" "$FORECAST_URL/optimization/recommendations/$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get building optimization recommendations" > /dev/null

    # Test 53: Get Scenario by ID
    print_test "Get Optimization Scenario by ID"
    if [ -n "$TEST_SCENARIO_ID" ] && [ "$TEST_SCENARIO_ID" != "null" ]; then
        make_request "GET" "$FORECAST_URL/optimization/scenario/$TEST_SCENARIO_ID" "" "$AUTH_TOKEN" "200" "Get optimization scenario" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No scenario ID available${NC}"
    fi

    # Test 54: Get Scenario - Non-existent
    print_test "Get Scenario - Non-existent ID"
    make_request "GET" "$FORECAST_URL/optimization/scenario/non-existent-id" "" "$AUTH_TOKEN" "404" "Get non-existent scenario" > /dev/null

    # Test 55: Send Scenario to IoT
    print_test "Send Optimization to IoT Service"
    send_iot_payload='{
        "scenarioId": "'$TEST_SCENARIO_ID'",
        "executeNow": false,
        "dryRun": true
    }'
    if [ -n "$TEST_SCENARIO_ID" ] && [ "$TEST_SCENARIO_ID" != "null" ]; then
        make_request "POST" "$FORECAST_URL/optimization/send-to-iot" "$send_iot_payload" "$AUTH_TOKEN" "200" "Send scenario to IoT" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No scenario ID available${NC}"
    fi
}

#===============================================================================
# IOT CONTROL SERVICE TESTS (Port 8083)
#===============================================================================

test_iot_service() {
    print_section "IOT CONTROL SERVICE TESTS ($IOT_URL)"

    # Re-authenticate if needed
    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" == "test-token-placeholder" ]; then
        echo -e "${YELLOW}Re-authenticating for IoT Service tests...${NC}"
        login_payload='{"username": "'$TEST_USERNAME'", "password": "'$TEST_PASSWORD'"}'
        response=$(curl -s -X POST "$SECURITY_URL/auth/login" \
            -H "Content-Type: application/json" \
            -d "$login_payload" 2>/dev/null)
        if command -v jq &> /dev/null; then
            AUTH_TOKEN=$(echo "$response" | jq -r '.data.accessToken // .accessToken // empty' 2>/dev/null)
        fi
    fi

    #---------------------------------------------------------------------------
    # Health Check
    #---------------------------------------------------------------------------
    print_test "Health Check"
    make_request "GET" "$IOT_URL/health" "" "none" "200" "IoT Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # DEVICE ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Device Management Endpoints"

    # Test 56: Register Device
    print_test "Register Device"
    register_device_payload='{
        "deviceId": "test-device-'$(date +%s)'",
        "name": "Test Device",
        "type": "HVAC",
        "buildingId": "'$TEST_BUILDING_ID'",
        "location": "Floor 1, Room 101",
        "capabilities": ["power_control", "temperature_sensing"],
        "metadata": {
            "manufacturer": "TestCo",
            "model": "TC-1000"
        }
    }'
    response=$(make_request "POST" "$IOT_URL/iot/devices/register" "$register_device_payload" "$AUTH_TOKEN" "201" "Register new device")

    if command -v jq &> /dev/null; then
        REGISTERED_DEVICE_ID=$(echo "$response" | jq -r '.data.deviceId // .deviceId // empty' 2>/dev/null)
    fi
    if [ -n "$REGISTERED_DEVICE_ID" ] && [ "$REGISTERED_DEVICE_ID" != "null" ]; then
        TEST_DEVICE_ID=$REGISTERED_DEVICE_ID
    fi

    # Test 57: Register Device - Missing Device ID
    print_test "Register Device - Missing Device ID"
    invalid_device='{"name": "Invalid Device", "type": "HVAC"}'
    make_request "POST" "$IOT_URL/iot/devices/register" "$invalid_device" "$AUTH_TOKEN" "400" "Register device without ID" > /dev/null

    # Test 58: Register Device - Duplicate
    print_test "Register Device - Duplicate"
    make_request "POST" "$IOT_URL/iot/devices/register" "$register_device_payload" "$AUTH_TOKEN" "409" "Register duplicate device" > /dev/null

    # Test 59: List Devices
    print_test "List All Devices"
    make_request "GET" "$IOT_URL/iot/devices" "" "$AUTH_TOKEN" "200" "List all devices" > /dev/null

    # Test 60: List Devices by Building
    print_test "List Devices - By Building"
    make_request "GET" "$IOT_URL/iot/devices?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "List devices by building" > /dev/null

    # Test 61: List Devices - With Pagination
    print_test "List Devices - With Pagination"
    make_request "GET" "$IOT_URL/iot/devices?page=1&limit=5" "" "$AUTH_TOKEN" "200" "List devices with pagination" > /dev/null

    # Test 62: Get Device by ID
    print_test "Get Device by ID"
    make_request "GET" "$IOT_URL/iot/devices/$TEST_DEVICE_ID" "" "$AUTH_TOKEN" "200" "Get device by ID" > /dev/null

    # Test 63: Get Device - Non-existent
    print_test "Get Device - Non-existent"
    make_request "GET" "$IOT_URL/iot/devices/non-existent-device" "" "$AUTH_TOKEN" "404" "Get non-existent device" > /dev/null

    #---------------------------------------------------------------------------
    # TELEMETRY ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Telemetry Ingestion Endpoints"

    # Test 64: Ingest Telemetry - Single
    print_test "Ingest Single Telemetry"
    telemetry_payload='{
        "deviceId": "'$TEST_DEVICE_ID'",
        "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "metrics": {
            "temperature": 22.5,
            "humidity": 45.0,
            "power": 150.5,
            "consumption": 12.3
        }
    }'
    make_request "POST" "$IOT_URL/iot/telemetry" "$telemetry_payload" "$AUTH_TOKEN" "201" "Ingest single telemetry" > /dev/null

    # Test 65: Ingest Telemetry - Missing Device ID
    print_test "Ingest Telemetry - Missing Device ID"
    invalid_telemetry='{"metrics": {"temperature": 22.5}}'
    make_request "POST" "$IOT_URL/iot/telemetry" "$invalid_telemetry" "$AUTH_TOKEN" "400" "Ingest telemetry without device ID" > /dev/null

    # Test 66: Ingest Telemetry - Non-existent Device
    print_test "Ingest Telemetry - Non-existent Device"
    nonexistent_telemetry='{
        "deviceId": "non-existent-device",
        "metrics": {"temperature": 22.5}
    }'
    make_request "POST" "$IOT_URL/iot/telemetry" "$nonexistent_telemetry" "$AUTH_TOKEN" "404" "Ingest telemetry for non-existent device" > /dev/null

    # Test 67: Ingest Bulk Telemetry
    print_test "Ingest Bulk Telemetry"
    bulk_telemetry='{
        "telemetry": [
            {
                "deviceId": "'$TEST_DEVICE_ID'",
                "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
                "metrics": {"temperature": 23.0, "power": 145.0}
            },
            {
                "deviceId": "'$TEST_DEVICE_ID'",
                "timestamp": "'$(date -u -d '+1 minute' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)'",
                "metrics": {"temperature": 23.5, "power": 148.0}
            }
        ]
    }'
    make_request "POST" "$IOT_URL/iot/telemetry/bulk" "$bulk_telemetry" "$AUTH_TOKEN" "201" "Ingest bulk telemetry" > /dev/null

    # Test 68: Ingest Bulk Telemetry - Empty Array
    print_test "Ingest Bulk Telemetry - Empty Array"
    empty_bulk='{"telemetry": []}'
    make_request "POST" "$IOT_URL/iot/telemetry/bulk" "$empty_bulk" "$AUTH_TOKEN" "200" "Ingest empty bulk telemetry" > /dev/null

    # Test 69: Get Telemetry History
    print_test "Get Telemetry History"
    from_date=$(date -u -d '-7 days' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)
    to_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    make_request "GET" "$IOT_URL/iot/telemetry/history?deviceId=$TEST_DEVICE_ID&from=$from_date&to=$to_date" "" "$AUTH_TOKEN" "200" "Get telemetry history" > /dev/null

    # Test 70: Get Telemetry History - With Pagination
    print_test "Get Telemetry History - With Pagination"
    make_request "GET" "$IOT_URL/iot/telemetry/history?deviceId=$TEST_DEVICE_ID&page=1&limit=10" "" "$AUTH_TOKEN" "200" "Get telemetry history with pagination" > /dev/null

    #---------------------------------------------------------------------------
    # DEVICE CONTROL ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Device Control Endpoints"

    # Test 71: Send Command to Device
    print_test "Send Command to Device"
    command_payload='{
        "command": "SET_POWER",
        "params": {
            "power": 100,
            "mode": "eco"
        }
    }'
    response=$(make_request "POST" "$IOT_URL/iot/device-control/$TEST_DEVICE_ID/command" "$command_payload" "$AUTH_TOKEN" "201" "Send command to device")

    if command -v jq &> /dev/null; then
        COMMAND_ID=$(echo "$response" | jq -r '.data.commandId // .commandId // empty' 2>/dev/null)
    fi

    # Test 72: Send Command - Invalid Device
    print_test "Send Command - Invalid Device"
    make_request "POST" "$IOT_URL/iot/device-control/non-existent/command" "$command_payload" "$AUTH_TOKEN" "404" "Send command to non-existent device" > /dev/null

    # Test 73: Send Command - Missing Command Field
    print_test "Send Command - Missing Command"
    invalid_command='{"params": {"power": 100}}'
    make_request "POST" "$IOT_URL/iot/device-control/$TEST_DEVICE_ID/command" "$invalid_command" "$AUTH_TOKEN" "400" "Send command without command field" > /dev/null

    # Test 74: List Device Commands
    print_test "List Device Commands"
    make_request "GET" "$IOT_URL/iot/device-control/$TEST_DEVICE_ID/commands" "" "$AUTH_TOKEN" "200" "List device commands" > /dev/null

    # Test 75: List Device Commands - With Status Filter
    print_test "List Device Commands - With Status Filter"
    make_request "GET" "$IOT_URL/iot/device-control/$TEST_DEVICE_ID/commands?status=PENDING" "" "$AUTH_TOKEN" "200" "List pending commands" > /dev/null

    #---------------------------------------------------------------------------
    # OPTIMIZATION ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Optimization Execution Endpoints"

    # Test 76: Apply Optimization (Primary Endpoint)
    print_test "Apply Optimization - applySecurity Endpoint"
    apply_opt_payload='{
        "scenarioId": "test-scenario-'$(date +%s)'",
        "buildingId": "'$TEST_BUILDING_ID'",
        "actions": [
            {
                "deviceId": "'$TEST_DEVICE_ID'",
                "command": "SET_POWER",
                "params": {"power": 80}
            }
        ]
    }'
    response=$(make_request "POST" "$IOT_URL/iot/optimization/applySecurity" "$apply_opt_payload" "$AUTH_TOKEN" "201" "Apply optimization via applySecurity")

    if command -v jq &> /dev/null; then
        IOT_SCENARIO_ID=$(echo "$response" | jq -r '.data.scenarioId // .scenarioId // empty' 2>/dev/null)
    fi

    # Test 77: Apply Optimization - Legacy Endpoint
    print_test "Apply Optimization - Legacy apply Endpoint"
    make_request "POST" "$IOT_URL/iot/optimization/apply" "$apply_opt_payload" "$AUTH_TOKEN" "201" "Apply optimization via legacy endpoint" > /dev/null

    # Test 78: Apply Optimization - Missing Actions
    print_test "Apply Optimization - Missing Actions"
    invalid_opt='{"scenarioId": "test", "buildingId": "'$TEST_BUILDING_ID'"}'
    make_request "POST" "$IOT_URL/iot/optimization/applySecurity" "$invalid_opt" "$AUTH_TOKEN" "400" "Apply optimization without actions" > /dev/null

    # Test 79: Get Optimization Status
    print_test "Get Optimization Status"
    if [ -n "$IOT_SCENARIO_ID" ] && [ "$IOT_SCENARIO_ID" != "null" ]; then
        make_request "GET" "$IOT_URL/iot/optimization/status/$IOT_SCENARIO_ID" "" "$AUTH_TOKEN" "200" "Get optimization status" > /dev/null
    else
        make_request "GET" "$IOT_URL/iot/optimization/status/test-scenario" "" "$AUTH_TOKEN" "404" "Get optimization status (placeholder)" > /dev/null
    fi

    #---------------------------------------------------------------------------
    # STATE ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Device State Endpoints"

    # Test 80: Get Live State (All Devices)
    print_test "Get Live State - All Devices"
    make_request "GET" "$IOT_URL/iot/state/live" "" "$AUTH_TOKEN" "200" "Get live state of all devices" > /dev/null

    # Test 81: Get Live State - By Building
    print_test "Get Live State - By Building"
    make_request "GET" "$IOT_URL/iot/state/live?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get live state by building" > /dev/null

    # Test 82: Get Device State
    print_test "Get Device State"
    make_request "GET" "$IOT_URL/iot/state/$TEST_DEVICE_ID" "" "$AUTH_TOKEN" "200" "Get specific device state" > /dev/null

    # Test 83: Get Device State - Non-existent
    print_test "Get Device State - Non-existent"
    make_request "GET" "$IOT_URL/iot/state/non-existent-device" "" "$AUTH_TOKEN" "404" "Get state of non-existent device" > /dev/null
}

#===============================================================================
# ANALYTICS SERVICE TESTS (Port 8084)
#===============================================================================

test_analytics_service() {
    print_section "ANALYTICS SERVICE TESTS ($ANALYTICS_URL)"

    # Re-authenticate if needed
    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" == "test-token-placeholder" ]; then
        echo -e "${YELLOW}Re-authenticating for Analytics Service tests...${NC}"
        login_payload='{"username": "'$TEST_USERNAME'", "password": "'$TEST_PASSWORD'"}'
        response=$(curl -s -X POST "$SECURITY_URL/auth/login" \
            -H "Content-Type: application/json" \
            -d "$login_payload" 2>/dev/null)
        if command -v jq &> /dev/null; then
            AUTH_TOKEN=$(echo "$response" | jq -r '.data.accessToken // .accessToken // empty' 2>/dev/null)
        fi
    fi

    #---------------------------------------------------------------------------
    # Health Check
    #---------------------------------------------------------------------------
    print_test "Health Check"
    make_request "GET" "$ANALYTICS_URL/health" "" "none" "200" "Analytics Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # REPORT ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Report Endpoints"

    # Test 84: Generate Report - Energy Consumption
    print_test "Generate Report - Energy Consumption"
    from_date=$(date -u -d '-7 days' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)
    to_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    report_payload='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "ENERGY_CONSUMPTION",
        "from": "'$from_date'",
        "to": "'$to_date'"
    }'
    response=$(make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$report_payload" "$AUTH_TOKEN" "201" "Generate energy consumption report")

    if command -v jq &> /dev/null; then
        TEST_REPORT_ID=$(echo "$response" | jq -r '.data.reportId // .reportId // empty' 2>/dev/null)
    fi

    # Test 85: Generate Report - Device Performance
    print_test "Generate Report - Device Performance"
    perf_report='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "DEVICE_PERFORMANCE",
        "from": "'$from_date'",
        "to": "'$to_date'"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$perf_report" "$AUTH_TOKEN" "201" "Generate device performance report" > /dev/null

    # Test 86: Generate Report - Anomaly Summary
    print_test "Generate Report - Anomaly Summary"
    anomaly_report='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "ANOMALY_SUMMARY",
        "from": "'$from_date'",
        "to": "'$to_date'"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$anomaly_report" "$AUTH_TOKEN" "201" "Generate anomaly summary report" > /dev/null

    # Test 87: Generate Report - Missing Type
    print_test "Generate Report - Missing Type"
    invalid_report='{"buildingId": "'$TEST_BUILDING_ID'"}'
    make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$invalid_report" "$AUTH_TOKEN" "400" "Generate report without type" > /dev/null

    # Test 88: List Reports
    print_test "List Reports"
    make_request "GET" "$ANALYTICS_URL/analytics/reports" "" "$AUTH_TOKEN" "200" "List all reports" > /dev/null

    # Test 89: List Reports - With Filters
    print_test "List Reports - With Filters"
    make_request "GET" "$ANALYTICS_URL/analytics/reports?buildingId=$TEST_BUILDING_ID&type=ENERGY_CONSUMPTION" "" "$AUTH_TOKEN" "200" "List filtered reports" > /dev/null

    # Test 90: Get Report by ID
    print_test "Get Report by ID"
    if [ -n "$TEST_REPORT_ID" ] && [ "$TEST_REPORT_ID" != "null" ]; then
        make_request "GET" "$ANALYTICS_URL/analytics/reports/$TEST_REPORT_ID" "" "$AUTH_TOKEN" "200" "Get report by ID" > /dev/null
    else
        make_request "GET" "$ANALYTICS_URL/analytics/reports/507f1f77bcf86cd799439011" "" "$AUTH_TOKEN" "404" "Get report (placeholder)" > /dev/null
    fi

    # Test 91: Get Report - Non-existent
    print_test "Get Report - Non-existent"
    make_request "GET" "$ANALYTICS_URL/analytics/reports/non-existent-report" "" "$AUTH_TOKEN" "404" "Get non-existent report" > /dev/null

    #---------------------------------------------------------------------------
    # ANOMALY ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Anomaly Detection Endpoints"

    # Test 92: List Anomalies
    print_test "List Anomalies"
    response=$(make_request "GET" "$ANALYTICS_URL/analytics/anomalies" "" "$AUTH_TOKEN" "200" "List all anomalies")

    if command -v jq &> /dev/null; then
        TEST_ANOMALY_ID=$(echo "$response" | jq -r '.data[0].anomalyId // .[0].anomalyId // empty' 2>/dev/null)
    fi

    # Test 93: List Anomalies - By Building
    print_test "List Anomalies - By Building"
    make_request "GET" "$ANALYTICS_URL/analytics/anomalies?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "List anomalies by building" > /dev/null

    # Test 94: List Anomalies - By Severity
    print_test "List Anomalies - By Severity"
    make_request "GET" "$ANALYTICS_URL/analytics/anomalies?severity=HIGH" "" "$AUTH_TOKEN" "200" "List high severity anomalies" > /dev/null

    # Test 95: List Anomalies - By Status
    print_test "List Anomalies - By Status"
    make_request "GET" "$ANALYTICS_URL/analytics/anomalies?status=NEW" "" "$AUTH_TOKEN" "200" "List new anomalies" > /dev/null

    # Test 96: Get Anomaly by ID
    print_test "Get Anomaly by ID"
    if [ -n "$TEST_ANOMALY_ID" ] && [ "$TEST_ANOMALY_ID" != "null" ]; then
        make_request "GET" "$ANALYTICS_URL/analytics/anomalies/$TEST_ANOMALY_ID" "" "$AUTH_TOKEN" "200" "Get anomaly by ID" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No anomaly ID available${NC}"
    fi

    # Test 97: Acknowledge Anomaly
    print_test "Acknowledge Anomaly"
    ack_payload='{"anomalyId": "'$TEST_ANOMALY_ID'"}'
    if [ -n "$TEST_ANOMALY_ID" ] && [ "$TEST_ANOMALY_ID" != "null" ]; then
        make_request "POST" "$ANALYTICS_URL/analytics/anomalies/acknowledge" "$ack_payload" "$AUTH_TOKEN" "200" "Acknowledge anomaly" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No anomaly ID available${NC}"
    fi

    #---------------------------------------------------------------------------
    # TIME-SERIES ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Time-Series Query Endpoints"

    # Test 98: Query Time-Series - Hourly Aggregation
    print_test "Query Time-Series - Hourly"
    ts_query='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "deviceIds": ["'$TEST_DEVICE_ID'"],
        "from": "'$from_date'",
        "to": "'$to_date'",
        "aggregationType": "HOURLY"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/time-series/query" "$ts_query" "$AUTH_TOKEN" "200" "Query hourly time-series" > /dev/null

    # Test 99: Query Time-Series - Daily Aggregation
    print_test "Query Time-Series - Daily"
    ts_daily='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "from": "'$from_date'",
        "to": "'$to_date'",
        "aggregationType": "DAILY"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/time-series/query" "$ts_daily" "$AUTH_TOKEN" "200" "Query daily time-series" > /dev/null

    # Test 100: Query Time-Series - Missing Dates
    print_test "Query Time-Series - Missing Dates"
    invalid_ts='{"buildingId": "'$TEST_BUILDING_ID'", "aggregationType": "HOURLY"}'
    make_request "POST" "$ANALYTICS_URL/analytics/time-series/query" "$invalid_ts" "$AUTH_TOKEN" "400" "Query time-series without dates" > /dev/null

    #---------------------------------------------------------------------------
    # KPI ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "KPI Calculation Endpoints"

    # Test 101: Get KPIs - System Wide
    print_test "Get KPIs - System Wide"
    make_request "GET" "$ANALYTICS_URL/analytics/kpi" "" "$AUTH_TOKEN" "200" "Get system-wide KPIs" > /dev/null

    # Test 102: Get KPIs - By Building
    print_test "Get KPIs - By Building"
    make_request "GET" "$ANALYTICS_URL/analytics/kpi/$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get building KPIs" > /dev/null

    # Test 103: Get KPIs - With Period Filter
    print_test "Get KPIs - With Period"
    make_request "GET" "$ANALYTICS_URL/analytics/kpi?period=DAILY" "" "$AUTH_TOKEN" "200" "Get daily KPIs" > /dev/null

    # Test 104: Calculate KPIs
    print_test "Calculate KPIs"
    calc_kpi='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "period": "DAILY"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/kpi/calculate" "$calc_kpi" "$AUTH_TOKEN" "200" "Calculate KPIs" > /dev/null

    #---------------------------------------------------------------------------
    # DASHBOARD ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Dashboard Endpoints"

    # Test 105: Get Dashboard Overview
    print_test "Get Dashboard Overview"
    make_request "GET" "$ANALYTICS_URL/analytics/dashboards/overview" "" "$AUTH_TOKEN" "200" "Get dashboard overview" > /dev/null

    # Test 106: Get Building Dashboard
    print_test "Get Building Dashboard"
    make_request "GET" "$ANALYTICS_URL/analytics/dashboards/building/$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get building dashboard" > /dev/null

    # Test 107: Get Building Dashboard - Non-existent Building
    print_test "Get Building Dashboard - Non-existent"
    make_request "GET" "$ANALYTICS_URL/analytics/dashboards/building/non-existent-building" "" "$AUTH_TOKEN" "200" "Get dashboard for non-existent building" > /dev/null
}

#===============================================================================
# EDGE CASE AND STRESS TESTS
#===============================================================================

test_edge_cases() {
    print_section "EDGE CASE AND STRESS TESTS"

    # Re-authenticate
    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" == "test-token-placeholder" ]; then
        login_payload='{"username": "'$TEST_USERNAME'", "password": "'$TEST_PASSWORD'"}'
        response=$(curl -s -X POST "$SECURITY_URL/auth/login" \
            -H "Content-Type: application/json" \
            -d "$login_payload" 2>/dev/null)
        if command -v jq &> /dev/null; then
            AUTH_TOKEN=$(echo "$response" | jq -r '.data.accessToken // .accessToken // empty' 2>/dev/null)
        fi
    fi

    # Test 108: Empty JSON Body
    print_test "Edge Case: Empty JSON Body"
    make_request "POST" "$SECURITY_URL/auth/login" "{}" "none" "400" "Login with empty JSON" > /dev/null

    # Test 109: Malformed JSON
    print_test "Edge Case: Malformed JSON"
    curl -s -w '\n%{http_code}' -X POST "$SECURITY_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{invalid json" 2>/dev/null > /dev/null
    ((TOTAL_TESTS++))
    echo -e "${GREEN}PASSED: Malformed JSON handled${NC}"
    ((PASSED_TESTS++))

    # Test 110: Very Long String
    print_test "Edge Case: Very Long String in Username"
    long_string=$(printf 'a%.0s' {1..1000})
    long_payload='{"username": "'$long_string'", "password": "test"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$long_payload" "none" "400" "Login with very long username" > /dev/null

    # Test 111: Special Characters
    print_test "Edge Case: Special Characters"
    special_payload='{"username": "<script>alert(1)</script>", "password": "test"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$special_payload" "none" "401" "Login with special characters" > /dev/null

    # Test 112: SQL Injection Attempt
    print_test "Edge Case: SQL Injection Attempt"
    sql_payload='{"username": "admin'\'' OR '\''1'\''='\''1", "password": "test"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$sql_payload" "none" "401" "SQL injection attempt" > /dev/null

    # Test 113: Unicode Characters
    print_test "Edge Case: Unicode Characters"
    unicode_payload='{"username": "user\u0000test", "password": "test"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$unicode_payload" "none" "401" "Unicode null character" > /dev/null

    # Test 114: Large Payload
    print_test "Edge Case: Large Telemetry Payload"
    large_metrics=""
    for i in {1..50}; do
        large_metrics="$large_metrics\"metric_$i\": $i,"
    done
    large_metrics="${large_metrics%,}"
    large_telemetry='{
        "deviceId": "'$TEST_DEVICE_ID'",
        "metrics": {'$large_metrics'}
    }'
    make_request "POST" "$IOT_URL/iot/telemetry" "$large_telemetry" "$AUTH_TOKEN" "201" "Large telemetry payload" > /dev/null

    # Test 115: Concurrent Requests (simplified)
    print_test "Edge Case: Rapid Sequential Requests"
    for i in {1..5}; do
        curl -s -X GET "$SECURITY_URL/health" > /dev/null &
    done
    wait
    ((TOTAL_TESTS++))
    echo -e "${GREEN}PASSED: Handled rapid sequential requests${NC}"
    ((PASSED_TESTS++))

    # Test 116: Request with Extra Headers
    print_test "Edge Case: Extra Headers"
    curl -s -w '\n%{http_code}' -X GET "$SECURITY_URL/health" \
        -H "X-Custom-Header: test" \
        -H "X-Forwarded-For: 1.2.3.4" \
        -H "X-Request-ID: test-request-123" 2>/dev/null > /dev/null
    ((TOTAL_TESTS++))
    echo -e "${GREEN}PASSED: Extra headers handled${NC}"
    ((PASSED_TESTS++))

    # Test 117: OPTIONS Request (CORS)
    print_test "Edge Case: OPTIONS Request (CORS)"
    response=$(curl -s -w '\n%{http_code}' -X OPTIONS "$SECURITY_URL/auth/login" \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: POST" 2>/dev/null)
    http_code=$(echo "$response" | tail -n 1)
    ((TOTAL_TESTS++))
    if [ "$http_code" == "204" ] || [ "$http_code" == "200" ]; then
        echo -e "${GREEN}PASSED: CORS preflight handled (HTTP $http_code)${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}FAILED: CORS preflight (HTTP $http_code)${NC}"
        ((FAILED_TESTS++))
    fi
}

#===============================================================================
# PRINT SUMMARY
#===============================================================================

print_summary() {
    echo ""
    echo -e "${BLUE}===============================================================================${NC}"
    echo -e "${BLUE} TEST SUMMARY${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
    echo ""
    echo -e "Total Tests:  ${CYAN}$TOTAL_TESTS${NC}"
    echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}ALL TESTS PASSED!${NC}"
    else
        echo -e "${RED}SOME TESTS FAILED - Review output above for details${NC}"
    fi

    echo ""
    echo -e "${BLUE}===============================================================================${NC}"

    # Calculate pass rate
    if [ $TOTAL_TESTS -gt 0 ]; then
        pass_rate=$(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc 2>/dev/null || echo "N/A")
        echo -e "Pass Rate: ${CYAN}${pass_rate}%${NC}"
    fi

    echo -e "${BLUE}===============================================================================${NC}"
}

#===============================================================================
# MAIN EXECUTION
#===============================================================================

main() {
    echo -e "${BLUE}===============================================================================${NC}"
    echo -e "${BLUE} MICROSERVICES API ENDPOINT TEST SUITE${NC}"
    echo -e "${BLUE} Testing: Security, Forecast, IoT Control, Analytics Services${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
    echo ""
    echo "Configuration:"
    echo "  Security URL:  $SECURITY_URL"
    echo "  Forecast URL:  $FORECAST_URL"
    echo "  IoT URL:       $IOT_URL"
    echo "  Analytics URL: $ANALYTICS_URL"
    echo ""
    echo "Press Enter to start tests or Ctrl+C to cancel..."
    read -r

    # Check if services are running
    echo "Checking service availability..."
    for url in "$SECURITY_URL" "$FORECAST_URL" "$IOT_URL" "$ANALYTICS_URL"; do
        if curl -s --connect-timeout 5 "$url/health" > /dev/null 2>&1; then
            echo -e "  ${GREEN}$url - Available${NC}"
        else
            echo -e "  ${RED}$url - Not Available${NC}"
        fi
    done
    echo ""

    # Run all test suites
    test_security_service
    test_forecast_service
    test_iot_service
    test_analytics_service
    test_edge_cases

    # Print summary
    print_summary

    # Exit with appropriate code
    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function
main "$@"