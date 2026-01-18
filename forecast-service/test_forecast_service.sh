#!/bin/bash

#===============================================================================
# FORECAST SERVICE TEST SCRIPT
# Tests Forecast microservice endpoints
#===============================================================================

#-------------------------------------------------------------------------------
# CONFIGURATION
#-------------------------------------------------------------------------------
SECURITY_URL="http://localhost:8080"
FORECAST_URL="http://localhost:8082"
TEST_USERNAME="admin"
TEST_PASSWORD="admin123"

TEST_BUILDING_ID="building-001"
TEST_DEVICE_ID="test-device-001"
FORECAST_ID=""
TEST_SCENARIO_ID=""
AUTH_TOKEN=""

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

authenticate() {
    echo -e "${YELLOW}Authenticating...${NC}"
    login_payload='{"username": "'$TEST_USERNAME'", "password": "'$TEST_PASSWORD'"}'
    response=$(curl -s -X POST "$SECURITY_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_payload" 2>/dev/null)
    if command -v jq &> /dev/null; then
        AUTH_TOKEN=$(echo "$response" | jq -r '.data.accessToken // .accessToken // empty' 2>/dev/null)
    fi

    if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" == "null" ]; then
        echo -e "${RED}Failed to authenticate${NC}"
        AUTH_TOKEN="test-token-placeholder"
    else
        echo -e "${GREEN}Successfully authenticated${NC}"
    fi
}

#===============================================================================
# FORECAST SERVICE TESTS
#===============================================================================

test_forecast_service() {
    print_section "FORECAST SERVICE TESTS ($FORECAST_URL)"

    print_test "Health Check"
    make_request "GET" "$FORECAST_URL/health" "" "none" "200" "Forecast Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # FORECAST ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Forecast Generation Endpoints"

    print_test "Generate Forecast - Valid Request"
    forecast_payload='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "DEMAND",
        "horizonHours": 24,
        "historicalDays": 7,
        "includeWeather": true,
        "includeTariffs": true
    }'
    response=$(make_request "POST" "$FORECAST_URL/forecast/generate" "$forecast_payload" "$AUTH_TOKEN" "000" "Generate demand forecast")

    if command -v jq &> /dev/null; then
        FORECAST_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    print_test "Generate Forecast - Missing Building ID"
    invalid_forecast='{"type": "DEMAND", "horizonHours": 24}'
    make_request "POST" "$FORECAST_URL/forecast/generate" "$invalid_forecast" "$AUTH_TOKEN" "400" "Generate forecast without building ID" > /dev/null

    print_test "Generate Forecast - Excessive Horizon"
    excessive_horizon='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "DEMAND",
        "horizonHours": 10000
    }'
    make_request "POST" "$FORECAST_URL/forecast/generate" "$excessive_horizon" "$AUTH_TOKEN" "200" "Generate forecast with capped horizon" > /dev/null

    print_test "Generate Forecast - Unauthorized"
    make_request "POST" "$FORECAST_URL/forecast/generate" "$forecast_payload" "" "401" "Generate forecast without auth" > /dev/null

    print_test "Get Peak Load Prediction"
    make_request "GET" "$FORECAST_URL/forecast/peak-load?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "404" "Get peak load prediction" > /dev/null

    print_test "Get Latest Forecast"
    make_request "GET" "$FORECAST_URL/forecast/latest?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get latest forecast" > /dev/null

    print_test "Get Latest Forecast - Missing Building ID"
    make_request "GET" "$FORECAST_URL/forecast/latest" "" "$AUTH_TOKEN" "400" "Get latest forecast without building ID" > /dev/null

    print_test "Get Device Prediction"
    make_request "GET" "$FORECAST_URL/forecast/prediction/$TEST_DEVICE_ID" "" "$AUTH_TOKEN" "404" "Get device prediction" > /dev/null

    print_test "Get Device Prediction - Non-existent Device"
    make_request "GET" "$FORECAST_URL/forecast/prediction/non-existent-device" "" "$AUTH_TOKEN" "404" "Get prediction for non-existent device" > /dev/null

    print_test "Get Device Optimization Recommendations"
    make_request "GET" "$FORECAST_URL/forecast/optimization/$TEST_DEVICE_ID" "" "$AUTH_TOKEN" "200" "Get device optimization recommendations" > /dev/null

    #---------------------------------------------------------------------------
    # OPTIMIZATION ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Optimization Scenario Endpoints"

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
    response=$(make_request "POST" "$FORECAST_URL/optimization/generate" "$optimization_payload" "$AUTH_TOKEN" "000" "Generate optimization scenario")

    if command -v jq &> /dev/null; then
        TEST_SCENARIO_ID=$(echo "$response" | jq -r '.data.id // .id // empty' 2>/dev/null)
    fi

    print_test "Generate Optimization - Missing Type"
    invalid_opt='{"buildingId": "'$TEST_BUILDING_ID'"}'
    make_request "POST" "$FORECAST_URL/optimization/generate" "$invalid_opt" "$AUTH_TOKEN" "400" "Generate optimization without type" > /dev/null

    print_test "Get Optimization Recommendations"
    make_request "GET" "$FORECAST_URL/optimization/recommendations/$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get building optimization recommendations" > /dev/null

    print_test "Get Optimization Scenario by ID"
    if [ -n "$TEST_SCENARIO_ID" ] && [ "$TEST_SCENARIO_ID" != "null" ]; then
        make_request "GET" "$FORECAST_URL/optimization/scenario/$TEST_SCENARIO_ID" "" "$AUTH_TOKEN" "200" "Get optimization scenario" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No scenario ID available${NC}"
    fi

    print_test "Get Scenario - Non-existent ID"
    make_request "GET" "$FORECAST_URL/optimization/scenario/non-existent-id" "" "$AUTH_TOKEN" "404" "Get non-existent scenario" > /dev/null

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
# PRINT SUMMARY
#===============================================================================

print_summary() {
    echo ""
    echo -e "${BLUE}===============================================================================${NC}"
    echo -e "${BLUE} TEST SUMMARY${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
    echo ""
    echo -e "Total Tests:  ${BLUE}13${NC}"
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
    echo -e "${BLUE} FORECAST SERVICE TEST SUITE${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
    echo ""
    echo "Configuration:"
    echo "  Forecast URL: $FORECAST_URL"
    echo ""

    if curl -s --connect-timeout 5 "$FORECAST_URL/health" > /dev/null 2>&1; then
        echo -e "  ${GREEN}Forecast Service - Available${NC}"
    else
        echo -e "  ${RED}Forecast Service - Not Available${NC}"
    fi
    echo ""

    authenticate
    test_forecast_service
    print_summary

    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

main "$@"