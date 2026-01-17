package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"analytics-service/internal/config"
	"analytics-service/internal/models"
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

// GetLatestForecast retrieves the latest forecast for a building
func (c *ForecastClient) GetLatestForecast(ctx context.Context, buildingID string, authToken string) (map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/forecast/latest?buildingId=%s", c.baseURL, url.QueryEscape(buildingID))

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
		return nil, fmt.Errorf("forecast service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("forecast service error: %s", apiResp.Error.Message)
	}

	if dataMap, ok := apiResp.Data.(map[string]interface{}); ok {
		return dataMap, nil
	}

	return nil, fmt.Errorf("invalid response format")
}
