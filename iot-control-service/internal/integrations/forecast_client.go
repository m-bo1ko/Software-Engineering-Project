package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"iot-control-service/internal/config"
	"iot-control-service/internal/models"
)

// ForecastClient handles communication with the Forecast & Optimization service
type ForecastClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewForecastClient creates a new forecast client
func NewForecastClient(cfg *config.Config) *ForecastClient {
	return &ForecastClient{
		httpClient: &http.Client{
			Timeout: cfg.Forecast.Timeout,
		},
		baseURL: cfg.Forecast.URL,
	}
}

// GetDevicePrediction retrieves predicted consumption for a device
func (c *ForecastClient) GetDevicePrediction(ctx context.Context, deviceID, authToken string) (*models.DevicePrediction, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/forecast/prediction/"+deviceID, nil)
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
		return nil, fmt.Errorf("forecast service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("forecast service error: %s", apiResp.Error.Message)
	}

	// Unmarshal data
	jsonData, _ := json.Marshal(apiResp.Data)
	var prediction models.DevicePrediction
	if err := json.Unmarshal(jsonData, &prediction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prediction: %w", err)
	}

	return &prediction, nil
}

// GetDeviceOptimization retrieves optimization recommendations for a device
func (c *ForecastClient) GetDeviceOptimization(ctx context.Context, deviceID, authToken string) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/forecast/optimization/"+deviceID, nil)
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
		return nil, fmt.Errorf("forecast service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("forecast service error: %s", apiResp.Error.Message)
	}

	return apiResp.Data, nil
}
