package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"forecast-service/internal/config"
	"forecast-service/internal/models"
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

// GetDeviceState retrieves the current state of a device
func (c *IoTClient) GetDeviceState(ctx context.Context, deviceID string, authToken string) (*models.DeviceState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/iot/state/%s", c.baseURL, deviceID), nil)
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
		return nil, fmt.Errorf("failed to get device state: status %d", resp.StatusCode)
	}

	var apiResp struct {
		Success bool               `json:"success"`
		Data    models.DeviceState `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp.Data, nil
}

// GetDevicesByBuilding retrieves all devices for a building
func (c *IoTClient) GetDevicesByBuilding(ctx context.Context, buildingID string, authToken string) ([]models.DeviceState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/iot/devices?buildingId=%s", c.baseURL, buildingID), nil)
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
		return nil, fmt.Errorf("failed to get devices: status %d", resp.StatusCode)
	}

	var apiResp struct {
		Success bool                 `json:"success"`
		Data    []models.DeviceState `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.Data, nil
}

// ApplyOptimizationRequest represents the request to apply optimization
type ApplyOptimizationRequest struct {
	ScenarioID string                      `json:"scenarioId"`
	BuildingID string                      `json:"buildingId"`
	Actions    []models.OptimizationAction `json:"actions"`
	ExecuteNow bool                        `json:"executeNow"`
	DryRun     bool                        `json:"dryRun"`
}

// ApplyOptimizationResponse represents the response from applying optimization
type ApplyOptimizationResponse struct {
	Success        bool     `json:"success"`
	ExecutionID    string   `json:"executionId"`
	ActionsQueued  int      `json:"actionsQueued"`
	ActionsSkipped int      `json:"actionsSkipped"`
	Errors         []string `json:"errors,omitempty"`
	Message        string   `json:"message,omitempty"`
}

// ApplyOptimization sends optimization actions to the IoT service
// Uses /iot/optimization/applySecurity endpoint as per integration contract
func (c *IoTClient) ApplyOptimization(ctx context.Context, scenario *models.OptimizationScenario, executeNow, dryRun bool, authToken string) (*ApplyOptimizationResponse, error) {
	payload := ApplyOptimizationRequest{
		ScenarioID: scenario.ID.Hex(),
		BuildingID: scenario.BuildingID,
		Actions:    scenario.Actions,
		ExecuteNow: executeNow,
		DryRun:     dryRun,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/iot/optimization/applySecurity", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result ApplyOptimizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return &result, fmt.Errorf("optimization apply failed: %s", result.Message)
	}

	return &result, nil
}

// ControlDevice sends a control command to a specific device
func (c *IoTClient) ControlDevice(ctx context.Context, deviceID string, action string, value string, authToken string) error {
	payload := map[string]string{
		"action": action,
		"value":  value,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/iot/control/%s", c.baseURL, deviceID), bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("device control failed with status: %d", resp.StatusCode)
	}

	return nil
}
