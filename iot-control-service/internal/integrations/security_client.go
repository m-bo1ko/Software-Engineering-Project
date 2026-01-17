// Package integrations handles external service integrations
package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"iot-control-service/internal/config"
	"iot-control-service/internal/models"
)

// SecurityClient handles communication with the Security & External Integration service
type SecurityClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewSecurityClient creates a new security client
func NewSecurityClient(cfg *config.Config) *SecurityClient {
	return &SecurityClient{
		httpClient: &http.Client{
			Timeout: cfg.Security.Timeout,
		},
		baseURL: cfg.Security.URL,
	}
}

// ValidateToken validates a JWT token with the security service
func (c *SecurityClient) ValidateToken(ctx context.Context, token string) (*models.TokenValidationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/validate-token", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result models.TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// LogAuditEvent sends an audit log entry to the security service
func (c *SecurityClient) LogAuditEvent(ctx context.Context, req *models.AuditLogRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/audit/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("audit log failed with status: %d", resp.StatusCode)
	}

	return nil
}

// AuditLog is a convenience method to log audit events
func (c *SecurityClient) AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{}) {
	req := &models.AuditLogRequest{
		UserID:      userID,
		Username:    username,
		Service:     "iot-control-service",
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Status:      status,
		ErrorMsg:    errorMsg,
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		RequestPath: requestPath,
		Method:      method,
	}

	// Run audit log asynchronously to not block the main request
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.LogAuditEvent(bgCtx, req)
	}()
}
