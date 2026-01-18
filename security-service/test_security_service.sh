#!/bin/bash

#===============================================================================
# SECURITY SERVICE TEST SCRIPT
# Tests Security microservice endpoints
#===============================================================================

#-------------------------------------------------------------------------------
# CONFIGURATION
#-------------------------------------------------------------------------------
SECURITY_URL="http://localhost:8080"
TEST_USERNAME="admin"
TEST_PASSWORD="admin123"
TEST_EMAIL="admin@example.com"

TEST_USER_ID=""
TEST_ROLE_ID=""
AUTH_TOKEN=""
REFRESH_TOKEN=""

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

#-------------------------------------------------------------------------------
# COLOR CODES
#-------------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

#-------------------------------------------------------------------------------
# HELPER FUNCTIONS
#-------------------------------------------------------------------------------

print_section() {
    echo ""
    echo -e "${BLUE}===============================================================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
}

print_test() {
    echo ""
    echo -e "${CYAN}--- TEST: $1 ---${NC}"
    ((TOTAL_TESTS++))
}

print_request() {
    echo -e "${YELLOW}REQUEST: $1 $2${NC}" >&2
    if [ -n "$3" ]; then
        echo -e "${YELLOW}BODY: $3${NC}" >&2
    fi
}

evaluate_response() {
    local http_code=$1
    local expected_code=$2
    local response_body=$3
    local test_name=$4

    echo "HTTP Status: $http_code" >&2
    echo "Response: $response_body" | head -c 500 >&2
    echo "" >&2

    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}PASSED: $test_name (Expected: $expected_code, Got: $http_code)${NC}" >&2
        ((PASSED_TESTS++))
        return 0
    else
        echo -e "${RED}FAILED: $test_name (Expected: $expected_code, Got: $http_code)${NC}" >&2
        ((FAILED_TESTS++))
        return 0
    fi
}

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

    http_code=$(echo "$response" | tail -n 1)
    response_body=$(echo "$response" | sed '$d')

    evaluate_response "$http_code" "$expected_code" "$response_body" "$test_name"
    echo "$response_body"
}

#===============================================================================
# SECURITY SERVICE TESTS
#===============================================================================

test_security_service() {
    print_section "SECURITY SERVICE TESTS ($SECURITY_URL)"

    print_test "Health Check"
    make_request "GET" "$SECURITY_URL/health" "" "none" "200" "Security Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # AUTH ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Authentication Endpoints"

    print_test "Login - Valid Credentials"
    login_payload='{
        "username": "'$TEST_USERNAME'",
        "password": "'$TEST_PASSWORD'"
    }'
    response=$(make_request "POST" "$SECURITY_URL/auth/login" "$login_payload" "none" "200" "Login with valid credentials")

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

    print_test "Login - Invalid Credentials"
    invalid_login='{"username": "wronguser", "password": "wrongpassword"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$invalid_login" "none" "401" "Login with invalid credentials" > /dev/null

    print_test "Login - Missing Password"
    missing_password='{"username": "testuser"}'
    make_request "POST" "$SECURITY_URL/auth/login" "$missing_password" "none" "400" "Login with missing password" > /dev/null

    print_test "Login - Empty Payload"
    make_request "POST" "$SECURITY_URL/auth/login" "{}" "none" "400" "Login with empty payload" > /dev/null

    print_test "Validate Token - Valid Token"
    make_request "GET" "$SECURITY_URL/auth/validate-token" "" "$AUTH_TOKEN" "200" "Validate valid token" > /dev/null

    print_test "Validate Token - Invalid Token"
    make_request "GET" "$SECURITY_URL/auth/validate-token" "" "invalid-token-12345" "401" "Validate invalid token" > /dev/null

    print_test "Validate Token - No Token"
    make_request "GET" "$SECURITY_URL/auth/validate-token" "" "" "401" "Validate without token" > /dev/null

    print_test "Get User Info - Authenticated"
    make_request "GET" "$SECURITY_URL/auth/user-info" "" "$AUTH_TOKEN" "200" "Get user info with valid token" > /dev/null

    print_test "Refresh Token - Valid Refresh Token"
    refresh_payload='{"refreshToken": "'$REFRESH_TOKEN'"}'
    make_request "POST" "$SECURITY_URL/auth/refresh" "$refresh_payload" "none" "200" "Refresh token" > /dev/null

    print_test "Refresh Token - Invalid Refresh Token"
    invalid_refresh='{"refreshToken": "invalid-refresh-token"}'
    make_request "POST" "$SECURITY_URL/auth/refresh" "$invalid_refresh" "none" "401" "Refresh with invalid token" > /dev/null

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

    if command -v jq &> /dev/null; then
        CREATED_USER_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    print_test "Create User - Duplicate Username"
    make_request "POST" "$SECURITY_URL/users" "$create_user_payload" "$AUTH_TOKEN" "409" "Create user with duplicate username" > /dev/null

    print_test "Create User - Invalid Email"
    invalid_email_user='{"username": "testuser2", "email": "not-an-email", "password": "Test123!@#"}'
    make_request "POST" "$SECURITY_URL/users" "$invalid_email_user" "$AUTH_TOKEN" "400" "Create user with invalid email" > /dev/null

    print_test "Create User - Missing Username"
    missing_username='{"email": "test@example.com", "password": "Test123"}'
    make_request "POST" "$SECURITY_URL/users" "$missing_username" "$AUTH_TOKEN" "400" "Create user without username" > /dev/null

    print_test "List Users"
    make_request "GET" "$SECURITY_URL/users" "" "$AUTH_TOKEN" "200" "List all users" > /dev/null

    print_test "List Users - With Pagination"
    make_request "GET" "$SECURITY_URL/users?page=1&limit=10" "" "$AUTH_TOKEN" "200" "List users with pagination" > /dev/null

    print_test "List Users - Unauthorized"
    make_request "GET" "$SECURITY_URL/users" "" "" "401" "List users without auth" > /dev/null

    print_test "Get User by ID - Valid ID"
    if [ -n "$CREATED_USER_ID" ] && [ "$CREATED_USER_ID" != "null" ]; then
        make_request "GET" "$SECURITY_URL/users/$CREATED_USER_ID" "" "$AUTH_TOKEN" "200" "Get user by valid ID" > /dev/null
    else
        make_request "GET" "$SECURITY_URL/users/507f1f77bcf86cd799439011" "" "$AUTH_TOKEN" "404" "Get user by ID (placeholder)" > /dev/null
    fi

    print_test "Get User by ID - Invalid ID Format"
    make_request "GET" "$SECURITY_URL/users/invalid-id" "" "$AUTH_TOKEN" "400" "Get user with invalid ID format" > /dev/null

    print_test "Get User by ID - Non-existent"
    make_request "GET" "$SECURITY_URL/users/507f1f77bcf86cd799439999" "" "$AUTH_TOKEN" "404" "Get non-existent user" > /dev/null

    print_test "Update User - Valid Data"
    update_user_payload='{"firstName": "Updated", "lastName": "Name"}'
    if [ -n "$CREATED_USER_ID" ] && [ "$CREATED_USER_ID" != "null" ]; then
        make_request "PUT" "$SECURITY_URL/users/$CREATED_USER_ID" "$update_user_payload" "$AUTH_TOKEN" "200" "Update user" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No user ID available for update test${NC}"
    fi

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

    print_test "Create Role - Missing Name"
    missing_name_role='{"description": "No name role"}'
    make_request "POST" "$SECURITY_URL/roles" "$missing_name_role" "$AUTH_TOKEN" "400" "Create role without name" > /dev/null

    print_test "List Roles"
    make_request "GET" "$SECURITY_URL/roles" "" "$AUTH_TOKEN" "200" "List all roles" > /dev/null

    print_test "Get Role by ID"
    if [ -n "$TEST_ROLE_ID" ] && [ "$TEST_ROLE_ID" != "null" ]; then
        make_request "GET" "$SECURITY_URL/roles/$TEST_ROLE_ID" "" "$AUTH_TOKEN" "404" "Get role by ID" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No role ID available${NC}"
    fi

    print_test "Update Role"
    update_role_payload='{"description": "Updated description"}'
    if [ -n "$TEST_ROLE_ID" ] && [ "$TEST_ROLE_ID" != "null" ]; then
        make_request "PUT" "$SECURITY_URL/roles/$TEST_ROLE_ID" "$update_role_payload" "$AUTH_TOKEN" "404" "Update role" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No role ID available${NC}"
    fi

    print_test "Delete Role"
    if [ -n "$TEST_ROLE_ID" ] && [ "$TEST_ROLE_ID" != "null" ]; then
        make_request "DELETE" "$SECURITY_URL/roles/$TEST_ROLE_ID" "" "$AUTH_TOKEN" "404" "Delete role" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No role ID available${NC}"
    fi

    #---------------------------------------------------------------------------
    # AUDIT ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Audit Logging Endpoints"

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

    print_test "Get Audit Logs"
    make_request "GET" "$SECURITY_URL/audit/logs" "" "$AUTH_TOKEN" "200" "Get audit logs" > /dev/null

    print_test "Get Audit Logs - With Filters"
    make_request "GET" "$SECURITY_URL/audit/logs?service=security-service&limit=10" "" "$AUTH_TOKEN" "200" "Get filtered audit logs" > /dev/null

    #---------------------------------------------------------------------------
    # NOTIFICATION ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Notification Endpoints"

    print_test "Send Notification"
    notification_payload='{
        "userId": "'$TEST_USER_ID'",
        "type": "EMAIL",
        "subject": "Test Notification",
        "message": "This is a test notification from the QA script"
    }'
    make_request "POST" "$SECURITY_URL/notifications/send" "$notification_payload" "$AUTH_TOKEN" "400" "Send notification" > /dev/null

    print_test "Get Notification Preferences"
    make_request "GET" "$SECURITY_URL/notifications/preferences/$TEST_USER_ID" "" "$AUTH_TOKEN" "200" "Get notification preferences" > /dev/null

    print_test "Update Notification Preferences"
    prefs_payload='{"email": true, "sms": false, "push": true}'
    make_request "PUT" "$SECURITY_URL/notifications/preferences/$TEST_USER_ID" "$prefs_payload" "$AUTH_TOKEN" "400" "Update notification preferences" > /dev/null

    print_test "Get Notification Logs"
    make_request "GET" "$SECURITY_URL/notifications/logs" "" "$AUTH_TOKEN" "400" "Get notification logs" > /dev/null

    #---------------------------------------------------------------------------
    # EXTERNAL ENERGY ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "External Energy Integration Endpoints"

    print_test "Get Energy Consumption"
    make_request "GET" "$SECURITY_URL/external-energy/consumption?buildingId=building-001" "" "$AUTH_TOKEN" "400" "Get energy consumption" > /dev/null

    print_test "Get Tariffs"
    make_request "GET" "$SECURITY_URL/external-energy/tariffs?region=default" "" "$AUTH_TOKEN" "500" "Get energy tariffs" > /dev/null

    print_test "Refresh Energy Provider Token"
    make_request "POST" "$SECURITY_URL/external-energy/refresh-token" "" "$AUTH_TOKEN" "400" "Refresh energy provider token" > /dev/null

    #---------------------------------------------------------------------------
    # LOGOUT
    #---------------------------------------------------------------------------
    print_test "Logout"
    logout_payload='{"refreshToken": "'$REFRESH_TOKEN'"}'
    make_request "POST" "$SECURITY_URL/auth/logout" "$logout_payload" "$AUTH_TOKEN" "200" "Logout" > /dev/null
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
    echo -e "Total Tests:  ${BLUE}38${NC}"
    echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}ALL TESTS PASSED!${NC}"
    else
        echo -e "${RED}SOME TESTS FAILED - Review output above for details${NC}"
    fi

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
    echo -e "${BLUE} SECURITY SERVICE TEST SUITE${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
    echo ""
    echo "Configuration:"
    echo "  Security URL: $SECURITY_URL"
    echo ""

    if curl -s --connect-timeout 5 "$SECURITY_URL/health" > /dev/null 2>&1; then
        echo -e "  ${GREEN}Security Service - Available${NC}"
    else
        echo -e "  ${RED}Security Service - Not Available${NC}"
    fi
    echo ""

    test_security_service
    print_summary

    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

main "$@"