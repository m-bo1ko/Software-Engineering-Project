#!/bin/bash

#===============================================================================
# ANALYTICS SERVICE TEST SCRIPT
# Tests Analytics microservice endpoints
#===============================================================================

#-------------------------------------------------------------------------------
# CONFIGURATION
#-------------------------------------------------------------------------------
SECURITY_URL="http://localhost:8080"
ANALYTICS_URL="http://localhost:8084"
TEST_USERNAME="admin"
TEST_PASSWORD="admin123"

TEST_BUILDING_ID="building-001"
TEST_DEVICE_ID="test-device-001"
TEST_REPORT_ID=""
TEST_ANOMALY_ID=""
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
# ANALYTICS SERVICE TESTS
#===============================================================================

test_analytics_service() {
    print_section "ANALYTICS SERVICE TESTS ($ANALYTICS_URL)"

    print_test "Health Check"
    make_request "GET" "$ANALYTICS_URL/health" "" "none" "200" "Analytics Service Health Check" > /dev/null

    #---------------------------------------------------------------------------
    # REPORT ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Report Endpoints"

    from_date=$(date -u -d '-7 days' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)
    to_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    print_test "Generate Report - Energy Consumption"
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

    print_test "Generate Report - Device Performance"
    perf_report='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "DEVICE_PERFORMANCE",
        "from": "'$from_date'",
        "to": "'$to_date'"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$perf_report" "$AUTH_TOKEN" "201" "Generate device performance report" > /dev/null

    print_test "Generate Report - Anomaly Summary"
    anomaly_report='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "type": "ANOMALY_SUMMARY",
        "from": "'$from_date'",
        "to": "'$to_date'"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$anomaly_report" "$AUTH_TOKEN" "201" "Generate anomaly summary report" > /dev/null

    print_test "Generate Report - Missing Type"
    invalid_report='{"buildingId": "'$TEST_BUILDING_ID'"}'
    make_request "POST" "$ANALYTICS_URL/analytics/reports/generate" "$invalid_report" "$AUTH_TOKEN" "400" "Generate report without type" > /dev/null

    print_test "List Reports"
    make_request "GET" "$ANALYTICS_URL/analytics/reports" "" "$AUTH_TOKEN" "200" "List all reports" > /dev/null

    print_test "List Reports - With Filters"
    make_request "GET" "$ANALYTICS_URL/analytics/reports?buildingId=$TEST_BUILDING_ID&type=ENERGY_CONSUMPTION" "" "$AUTH_TOKEN" "200" "List filtered reports" > /dev/null

    print_test "Get Report by ID"
    if [ -n "$TEST_REPORT_ID" ] && [ "$TEST_REPORT_ID" != "null" ]; then
        make_request "GET" "$ANALYTICS_URL/analytics/reports/$TEST_REPORT_ID" "" "$AUTH_TOKEN" "200" "Get report by ID" > /dev/null
    else
        make_request "GET" "$ANALYTICS_URL/analytics/reports/507f1f77bcf86cd799439011" "" "$AUTH_TOKEN" "404" "Get report (placeholder)" > /dev/null
    fi

    print_test "Get Report - Non-existent"
    make_request "GET" "$ANALYTICS_URL/analytics/reports/non-existent-report" "" "$AUTH_TOKEN" "404" "Get non-existent report" > /dev/null

    #---------------------------------------------------------------------------
    # ANOMALY ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "Anomaly Detection Endpoints"

    print_test "List Anomalies"
    response=$(make_request "GET" "$ANALYTICS_URL/analytics/anomalies" "" "$AUTH_TOKEN" "200" "List all anomalies")

    if command -v jq &> /dev/null; then
        TEST_ANOMALY_ID=$(echo "$response" | jq -r '.data[0].anomalyId // .[0].anomalyId // empty' 2>/dev/null)
    fi

    print_test "List Anomalies - By Building"
    make_request "GET" "$ANALYTICS_URL/analytics/anomalies?buildingId=$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "List anomalies by building" > /dev/null

    print_test "List Anomalies - By Severity"
    make_request "GET" "$ANALYTICS_URL/analytics/anomalies?severity=HIGH" "" "$AUTH_TOKEN" "200" "List high severity anomalies" > /dev/null

    print_test "List Anomalies - By Status"
    make_request "GET" "$ANALYTICS_URL/analytics/anomalies?status=NEW" "" "$AUTH_TOKEN" "200" "List new anomalies" > /dev/null

    print_test "Get Anomaly by ID"
    if [ -n "$TEST_ANOMALY_ID" ] && [ "$TEST_ANOMALY_ID" != "null" ]; then
        make_request "GET" "$ANALYTICS_URL/analytics/anomalies/$TEST_ANOMALY_ID" "" "$AUTH_TOKEN" "200" "Get anomaly by ID" > /dev/null
    else
        echo -e "${YELLOW}SKIPPED: No anomaly ID available${NC}"
    fi

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

    print_test "Query Time-Series - Hourly"
    ts_query='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "deviceIds": ["'$TEST_DEVICE_ID'"],
        "from": "'$from_date'",
        "to": "'$to_date'",
        "aggregationType": "HOURLY"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/time-series/query" "$ts_query" "$AUTH_TOKEN" "200" "Query hourly time-series" > /dev/null

    print_test "Query Time-Series - Daily"
    ts_daily='{
        "buildingId": "'$TEST_BUILDING_ID'",
        "from": "'$from_date'",
        "to": "'$to_date'",
        "aggregationType": "DAILY"
    }'
    make_request "POST" "$ANALYTICS_URL/analytics/time-series/query" "$ts_daily" "$AUTH_TOKEN" "200" "Query daily time-series" > /dev/null

    print_test "Query Time-Series - Missing Dates"
    invalid_ts='{"buildingId": "'$TEST_BUILDING_ID'", "aggregationType": "HOURLY"}'
    make_request "POST" "$ANALYTICS_URL/analytics/time-series/query" "$invalid_ts" "$AUTH_TOKEN" "400" "Query time-series without dates" > /dev/null

    #---------------------------------------------------------------------------
    # KPI ENDPOINTS
    #---------------------------------------------------------------------------
    print_section "KPI Calculation Endpoints"

    print_test "Get KPIs - System Wide"
    make_request "GET" "$ANALYTICS_URL/analytics/kpi" "" "$AUTH_TOKEN" "200" "Get system-wide KPIs" > /dev/null

    print_test "Get KPIs - By Building"
    make_request "GET" "$ANALYTICS_URL/analytics/kpi/$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "404" "Get building KPIs" > /dev/null

    print_test "Get KPIs - With Period Filter"
    make_request "GET" "$ANALYTICS_URL/analytics/kpi?period=DAILY" "" "$AUTH_TOKEN" "200" "Get daily KPIs" > /dev/null

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

    print_test "Get Dashboard Overview"
    make_request "GET" "$ANALYTICS_URL/analytics/dashboards/overview" "" "$AUTH_TOKEN" "200" "Get dashboard overview" > /dev/null

    print_test "Get Building Dashboard"
    make_request "GET" "$ANALYTICS_URL/analytics/dashboards/building/$TEST_BUILDING_ID" "" "$AUTH_TOKEN" "200" "Get building dashboard" > /dev/null

    print_test "Get Building Dashboard - Non-existent"
    make_request "GET" "$ANALYTICS_URL/analytics/dashboards/building/non-existent-building" "" "$AUTH_TOKEN" "200" "Get dashboard for non-existent building" > /dev/null
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
    echo -e "Total Tests:  ${BLUE}21${NC}"
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
    echo -e "${BLUE} ANALYTICS SERVICE TEST SUITE${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
    echo ""
    echo "Configuration:"
    echo "  Analytics URL: $ANALYTICS_URL"
    echo ""

    if curl -s --connect-timeout 5 "$ANALYTICS_URL/health" > /dev/null 2>&1; then
        echo -e "  ${GREEN}Analytics Service - Available${NC}"
    else
        echo -e "  ${RED}Analytics Service - Not Available${NC}"
    fi
    echo ""

    authenticate
    test_analytics_service
    print_summary

    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

main "$@"