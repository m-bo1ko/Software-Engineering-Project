#!/bin/bash

# --- НАСТРОЙКИ ---
# Адреса сервисов
AUTH_URL="http://localhost:8081/auth/login"
FORECAST_SERVICE_URL="http://localhost:8082"

# Данные для входа (Security Service)
USERNAME="admin"
PASSWORD="admin123"

# Тестовые данные для запросов
BUILDING_ID="building-123"
DEVICE_ID="device-123"

# Цвета для вывода в консоль
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== Starting Forecast Service Integration Test ===${NC}"

# --- 1. ПОЛУЧЕНИЕ ТОКЕНА ---
echo "Attempting to get JWT token from Security Service..."

# Проверяем наличие jq
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: 'jq' is not installed. Please install it to parse JSON.${NC}"
    exit 1
fi

TOKEN_RESPONSE=$(curl -s -X POST "$AUTH_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}")

TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.data.accessToken // empty')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}Error: Could not retrieve token.${NC}"
  echo "Response from Auth: $TOKEN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}Token obtained successfully.${NC}"
echo "--------------------------------------------------"

# Функция для вывода заголовка теста
test_header() {
    echo -e "\n${BLUE}TEST:${NC} $1"
}

# --- 2. ПРОВЕРКА HEALTH CHECK ---
test_header "Health Check"
curl -s "$FORECAST_SERVICE_URL/health" | jq .

# --- 3. ГЕНЕРАЦИЯ ПРОГНОЗА ПОТРЕБЛЕНИЯ ---
test_header "Generate Energy Demand Forecast"
curl -s -X POST "$FORECAST_SERVICE_URL/forecast/generate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"buildingId\": \"$BUILDING_ID\",
    \"type\": \"DEMAND\",
    \"horizonHours\": 24,
    \"includeWeather\": true,
    \"includeTariffs\": true,
    \"historicalDays\": 30
  }" | jq .

# --- 4. ПРОГНОЗ ПИКОВОЙ НАГРУЗКИ ---
test_header "Predict Peak Load Periods"
# Генерируем даты в формате ISO8601 для запроса
START_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
END_DATE=$(date -u -v+1d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 day" +"%Y-%m-%dT%H:%M:%SZ")

curl -s -X POST "$FORECAST_SERVICE_URL/forecast/peak-load" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"buildingId\": \"$BUILDING_ID\",
    \"analysisFromDate\": \"$START_DATE\",
    \"analysisToDate\": \"$END_DATE\",
    \"thresholdPercent\": 80.0,
    \"includeWeather\": true
  }" | jq .

# --- 5. ПОЛУЧЕНИЕ ОПТИМИЗАЦИИ ДЛЯ УСТРОЙСТВА ---
test_header "Get Device Optimization Recommendations"
curl -s -X GET "$FORECAST_SERVICE_URL/forecast/optimization/$DEVICE_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .

# --- 6. СОЗДАНИЕ СЦЕНАРИЯ ОПТИМИЗАЦИИ ---
test_header "Generate Optimization Scenario"
SCENARIO_RESPONSE=$(curl -s -X POST "$FORECAST_SERVICE_URL/optimization/generate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"buildingId\": \"$BUILDING_ID\",
    \"type\": \"COST_REDUCTION\",
    \"priority\": 5,
    \"constraints\": { \"preserveComfort\": true }
  }")

echo $SCENARIO_RESPONSE | jq .

# Извлекаем ID сценария для теста отправки в IoT
SCENARIO_ID=$(echo $SCENARIO_RESPONSE | jq -r '.data.id // empty')

# --- 7. ОТПРАВКА В IOT СЕРВИС (если сценарий создан) ---
if [ ! -z "$SCENARIO_ID" ] && [ "$SCENARIO_ID" != "null" ]; then
    test_header "Sending Scenario to IoT Service"
    curl -s -X POST "$FORECAST_SERVICE_URL/optimization/send-to-iot" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"scenarioId\": \"$SCENARIO_ID\",
        \"executeNow\": false,
        \"dryRun\": true
      }" | jq .
else
    echo -e "\n${RED}Skipping 'Send to IoT' test: No Scenario ID received.${NC}"
fi

# --- 8. ПОЛУЧЕНИЕ РЕКОМЕНДАЦИЙ ДЛЯ ЗДАНИЯ ---
test_header "Get Building Recommendations"
curl -s -X GET "$FORECAST_SERVICE_URL/optimization/recommendations/$BUILDING_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n${BLUE}=== Testing Completed ===${NC}"