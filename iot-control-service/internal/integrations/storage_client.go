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

	"iot-control-service/internal/config"
	"iot-control-service/internal/models"
)

// StorageClient handles communication with the external Storage service
// as per integration contract requirements:
// - /storage/telemetry/save
// - /storage/commands/save
// - /storage/devices/{id}/history
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

// SaveTelemetry saves telemetry data to the storage service
// POST /storage/telemetry/save
func (c *StorageClient) SaveTelemetry(ctx context.Context, telemetry *models.Telemetry, authToken string) error {
	jsonData, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/telemetry/save", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

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

// SaveTelemetryBulk saves multiple telemetry records to the storage service
// POST /storage/telemetry/save (with array payload)
func (c *StorageClient) SaveTelemetryBulk(ctx context.Context, telemetryList []*models.Telemetry, authToken string) error {
	jsonData, err := json.Marshal(telemetryList)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry list: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/telemetry/save", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

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

// SaveCommand saves a device command to the storage service
// POST /storage/commands/save
func (c *StorageClient) SaveCommand(ctx context.Context, command *models.DeviceCommand, authToken string) error {
	jsonData, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/commands/save", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

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

// DeviceHistoryResponse represents the response from device history endpoint
type DeviceHistoryResponse struct {
	DeviceID   string                   `json:"deviceId"`
	History    []map[string]interface{} `json:"history"`
	TotalCount int                      `json:"totalCount"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
}

// GetDeviceHistory retrieves device history from the storage service
// GET /storage/devices/{id}/history
func (c *StorageClient) GetDeviceHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) (*DeviceHistoryResponse, error) {
	reqURL := fmt.Sprintf("%s/devices/%s/history?from=%s&to=%s&page=%d&limit=%d",
		c.baseURL,
		url.PathEscape(deviceID),
		url.QueryEscape(from.Format(time.RFC3339)),
		url.QueryEscape(to.Format(time.RFC3339)),
		page,
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("storage service error: %s", apiResp.Error.Message)
	}

	// Convert data to DeviceHistoryResponse
	jsonData, _ := json.Marshal(apiResp.Data)
	var historyResp DeviceHistoryResponse
	if err := json.Unmarshal(jsonData, &historyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history response: %w", err)
	}

	return &historyResp, nil
}
