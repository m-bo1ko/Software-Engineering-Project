package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"security-service/internal/models"
)

// TestAuditLogModel tests the AuditLog model
func TestAuditLogModel(t *testing.T) {
	t.Run("Create audit log", func(t *testing.T) {
		log := models.AuditLog{
			UserID:      "user123",
			Username:    "testuser",
			Service:     "security-service",
			Action:      "LOGIN",
			Resource:    "auth",
			Status:      "SUCCESS",
			IPAddress:   "192.168.1.1",
			UserAgent:   "Mozilla/5.0",
			Timestamp:   time.Now(),
			RequestPath: "/auth/login",
			Method:      "POST",
		}

		assert.Equal(t, "user123", log.UserID)
		assert.Equal(t, "LOGIN", log.Action)
		assert.Equal(t, "SUCCESS", log.Status)
	})

	t.Run("Create audit log with details", func(t *testing.T) {
		log := models.AuditLog{
			UserID:   "user123",
			Service:  "building-service",
			Action:   "CREATE_BUILDING",
			Resource: "building",
			Details: map[string]interface{}{
				"building_name": "Main Office",
				"location":      "New York",
			},
			Status:    "SUCCESS",
			Timestamp: time.Now(),
		}

		assert.NotNil(t, log.Details)
		assert.Equal(t, "Main Office", log.Details["building_name"])
	})
}

// TestAuditLogToResponse tests the ToResponse method
func TestAuditLogToResponse(t *testing.T) {
	log := models.AuditLog{
		UserID:      "user123",
		Username:    "testuser",
		Service:     "security-service",
		Action:      "UPDATE_USER",
		Resource:    "user",
		ResourceID:  "user456",
		Status:      "SUCCESS",
		IPAddress:   "10.0.0.1",
		UserAgent:   "TestAgent/1.0",
		Timestamp:   time.Now(),
		RequestPath: "/users/user456",
		Method:      "PUT",
	}
	log.ID = [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	resp := log.ToResponse()

	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, log.UserID, resp.UserID)
	assert.Equal(t, log.Username, resp.Username)
	assert.Equal(t, log.Service, resp.Service)
	assert.Equal(t, log.Action, resp.Action)
	assert.Equal(t, log.Resource, resp.Resource)
	assert.Equal(t, log.ResourceID, resp.ResourceID)
	assert.Equal(t, log.Status, resp.Status)
}

// TestAuditLogCreateRequest tests audit log creation request
func TestAuditLogCreateRequest(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		req := models.AuditLogCreateRequest{
			UserID:      "user123",
			Username:    "admin",
			Service:     "security-service",
			Action:      "DELETE_USER",
			Resource:    "user",
			ResourceID:  "user789",
			Status:      "SUCCESS",
			IPAddress:   "127.0.0.1",
			UserAgent:   "curl/7.68.0",
			RequestPath: "/users/user789",
			Method:      "DELETE",
		}

		assert.Equal(t, "DELETE_USER", req.Action)
		assert.Equal(t, "SUCCESS", req.Status)
	})

	t.Run("Request with failure", func(t *testing.T) {
		req := models.AuditLogCreateRequest{
			Service:   "security-service",
			Action:    "LOGIN",
			Resource:  "auth",
			Status:    "FAILURE",
			ErrorMsg:  "Invalid credentials",
			IPAddress: "192.168.1.100",
		}

		assert.Equal(t, "FAILURE", req.Status)
		assert.Equal(t, "Invalid credentials", req.ErrorMsg)
	})
}

// TestAuditLogQueryParams tests query parameters for filtering
func TestAuditLogQueryParams(t *testing.T) {
	t.Run("Query with all parameters", func(t *testing.T) {
		now := time.Now()
		params := models.AuditLogQueryParams{
			From:     now.Add(-24 * time.Hour),
			To:       now,
			UserID:   "user123",
			Service:  "security-service",
			Action:   "LOGIN",
			Resource: "auth",
			Status:   "SUCCESS",
			Page:     1,
			Limit:    20,
		}

		assert.True(t, params.From.Before(params.To))
		assert.Equal(t, "user123", params.UserID)
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.Limit)
	})

	t.Run("Query with minimal parameters", func(t *testing.T) {
		params := models.AuditLogQueryParams{
			Page:  1,
			Limit: 10,
		}

		assert.True(t, params.From.IsZero())
		assert.Equal(t, "", params.UserID)
	})
}

// TestPaginatedAuditLogsResponse tests paginated response
func TestPaginatedAuditLogsResponse(t *testing.T) {
	logs := []*models.AuditLogResponse{
		{ID: "1", Action: "LOGIN", Status: "SUCCESS"},
		{ID: "2", Action: "LOGOUT", Status: "SUCCESS"},
	}

	resp := models.PaginatedAuditLogsResponse{
		Logs:       logs,
		Total:      100,
		Page:       1,
		Limit:      20,
		TotalPages: 5,
	}

	assert.Len(t, resp.Logs, 2)
	assert.Equal(t, int64(100), resp.Total)
	assert.Equal(t, 5, resp.TotalPages)
}

// TestAuditLogJSONSerialization tests JSON serialization
func TestAuditLogJSONSerialization(t *testing.T) {
	log := models.AuditLog{
		UserID:    "user123",
		Service:   "security-service",
		Action:    "TEST_ACTION",
		Resource:  "test",
		Status:    "SUCCESS",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	jsonData, err := json.Marshal(log)
	require.NoError(t, err)

	var decoded models.AuditLog
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, log.UserID, decoded.UserID)
	assert.Equal(t, log.Action, decoded.Action)
	assert.NotNil(t, decoded.Details)
}

// TestAuditActionTypes tests common audit action types
func TestAuditActionTypes(t *testing.T) {
	actions := []string{
		"LOGIN",
		"LOGOUT",
		"CREATE_USER",
		"UPDATE_USER",
		"DELETE_USER",
		"CREATE_ROLE",
		"UPDATE_ROLE",
		"DELETE_ROLE",
		"TOKEN_REFRESH",
		"PASSWORD_CHANGE",
		"PERMISSION_CHECK",
	}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			log := models.AuditLog{
				Action: action,
				Status: "SUCCESS",
			}
			assert.Equal(t, action, log.Action)
		})
	}
}

// TestAuditLogStatusValues tests valid status values
func TestAuditLogStatusValues(t *testing.T) {
	validStatuses := []string{"SUCCESS", "FAILURE"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			req := models.AuditLogCreateRequest{
				Service:  "test-service",
				Action:   "TEST",
				Resource: "test",
				Status:   status,
			}
			assert.Equal(t, status, req.Status)
		})
	}
}
