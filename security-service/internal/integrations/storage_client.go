// Package integrations handles external service integrations
package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"security-service/internal/config"
	"security-service/internal/models"
)

// StorageClient handles communication with the external Storage service
// as per integration contract requirements:
// - /storage/auth/credentials
// - /storage/audit/save
// - /storage/audit/query
type StorageClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewStorageClient creates a new storage client
func NewStorageClient(cfg *config.Config) *StorageClient {
	return &StorageClient{
		httpClient: &http.Client{
			Timeout: cfg.Storage.Timeout,
		},
		baseURL: cfg.Storage.URL,
	}
}

// SaveAuthCredential saves authentication credentials to the storage service
// POST /storage/auth/credentials
func (c *StorageClient) SaveAuthCredential(ctx context.Context, credential *models.AuthCredential) error {
	jsonData, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/credentials", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	return nil
}

// GetAuthCredential retrieves authentication credentials from the storage service
// GET /storage/auth/credentials/{serviceName}
func (c *StorageClient) GetAuthCredential(ctx context.Context, serviceName string) (*models.AuthCredential, error) {
	reqURL := fmt.Sprintf("%s/auth/credentials/%s", c.baseURL, url.PathEscape(serviceName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("credential not found: %s", serviceName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	var credential models.AuthCredential
	if err := json.NewDecoder(resp.Body).Decode(&credential); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &credential, nil
}

// UpdateAuthCredential updates authentication credentials in the storage service
// PUT /storage/auth/credentials/{serviceName}
func (c *StorageClient) UpdateAuthCredential(ctx context.Context, serviceName string, token string, expiresAt time.Time) error {
	payload := map[string]interface{}{
		"encrypted_token":  token,
		"token_expires_at": expiresAt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	reqURL := fmt.Sprintf("%s/auth/credentials/%s", c.baseURL, url.PathEscape(serviceName))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	return nil
}

// SaveAuditLog saves an audit log entry to the storage service
// POST /storage/audit/save
func (c *StorageClient) SaveAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	jsonData, err := json.Marshal(auditLog)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/audit/save", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	return nil
}

// AuditQueryRequest represents query parameters for audit log search
type AuditQueryRequest struct {
	UserID   string    `json:"userId,omitempty"`
	Service  string    `json:"service,omitempty"`
	Action   string    `json:"action,omitempty"`
	Resource string    `json:"resource,omitempty"`
	Status   string    `json:"status,omitempty"`
	From     time.Time `json:"from,omitempty"`
	To       time.Time `json:"to,omitempty"`
	Page     int       `json:"page,omitempty"`
	Limit    int       `json:"limit,omitempty"`
}

// AuditQueryResponse represents the response from audit query
type AuditQueryResponse struct {
	Logs       []models.AuditLog `json:"logs"`
	TotalCount int64             `json:"totalCount"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
}

// QueryAuditLogs queries audit logs from the storage service
// POST /storage/audit/query
func (c *StorageClient) QueryAuditLogs(ctx context.Context, query *AuditQueryRequest) (*AuditQueryResponse, error) {
	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/audit/query", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	var result AuditQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
