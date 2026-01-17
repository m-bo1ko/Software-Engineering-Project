  BASE_URL="http://localhost:8081"

  echo "=== Testing Security Service ==="

  # Health check
  echo -e "\n1. Health Check"
  curl -s $BASE_URL/health | jq

  # Login
  echo -e "\n2. Login as admin"
  LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin123"}')
  echo $LOGIN_RESPONSE | jq

  ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.accessToken')
  USER_ID=$(echo $LOGIN_RESPONSE | jq -r '.data.userId')

  echo -e "\n3. Get User Info"
  curl -s -X GET $BASE_URL/auth/user-info \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq

  echo -e "\n4. Validate Token"
  curl -s -X GET $BASE_URL/auth/validate-token \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq

  echo -e "\n5. List Users"
  curl -s -X GET "$BASE_URL/users?page=1&limit=10" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq

  echo -e "\n6. List Roles"
  curl -s -X GET $BASE_URL/roles \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq

  echo -e "\n7. Check Permissions"
  curl -s -X POST $BASE_URL/auth/check-permissions \
    -H "Content-Type: application/json" \
    -d "{\"userId\": \"$USER_ID\", \"resource\": \"users\", \"action\": \"write\"}" | jq

  echo -e "\n8. Create Test User"
  curl -s -X POST $BASE_URL/users \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "username": "testuser123",
      "email": "test123@example.com",
      "password": "TestPass123!",
      "firstName": "Test",
      "lastName": "User",
      "roles": ["user"]
    }' | jq

  echo -e "\n9. Get Audit Logs"
  curl -s -X GET "$BASE_URL/audit/logs?page=1&limit=5" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq

  echo -e "\n=== Tests Complete ==="
