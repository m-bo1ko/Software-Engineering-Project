package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"analytics-service/internal/config"
	"analytics-service/internal/models"
)

// IoTClient handles communication with the IoT & Control service
type IoTClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewIoTClient creates a new IoT client
func NewIoTClient(cfg *config.Config) *IoTClient {
	return &IoTClient{
		httpClient: &http.Client{
			Timeout: cfg.IoT.Timeout,
		},
		baseURL: cfg.IoT.URL,
	}
}

// GetTelemetryHistory retrieves historical telemetry data
func (c *IoTClient) GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/iot/telemetry/history?deviceId=%s&from=%s&to=%s&page=%d&limit=%d",
		c.baseURL,
		url.QueryEscape(deviceID),
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IoT service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("IoT service error: %s", apiResp.Error.Message)
	}

	// Extract telemetry data
	dataMap, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	telemetryData, ok := dataMap["telemetry"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("telemetry data not found in response")
	}

	result := make([]map[string]interface{}, len(telemetryData))
	for i, item := range telemetryData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			result[i] = itemMap
		}
	}

	return result, nil
}

// GetDevices retrieves device list
func (c *IoTClient) GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/iot/devices", c.baseURL)
	if buildingID != "" {
		reqURL += "?buildingId=" + url.QueryEscape(buildingID)
	}

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IoT service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("IoT service error: %s", apiResp.Error.Message)
	}

	// Extract devices data
	dataMap, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	devicesData, ok := dataMap["devices"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("devices data not found in response")
	}

	result := make([]map[string]interface{}, len(devicesData))
	for i, item := range devicesData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			result[i] = itemMap
		}
	}

	return result, nil
}

// GetDeviceState retrieves device state
func (c *IoTClient) GetDeviceState(ctx context.Context, deviceID string, authToken string) (map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/iot/state/%s", c.baseURL, url.QueryEscape(deviceID))

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IoT service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("IoT service error: %s", apiResp.Error.Message)
	}

	if dataMap, ok := apiResp.Data.(map[string]interface{}); ok {
		return dataMap, nil
	}

	return nil, fmt.Errorf("invalid response format")
}
