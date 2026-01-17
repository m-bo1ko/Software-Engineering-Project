package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"iot-control-service/internal/config"
	"iot-control-service/internal/models"
)

// AnalyticsClient handles communication with the Analytics service
type AnalyticsClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewAnalyticsClient creates a new analytics client
func NewAnalyticsClient(cfg *config.Config) *AnalyticsClient {
	return &AnalyticsClient{
		httpClient: &http.Client{
			Timeout: cfg.Analytics.Timeout,
		},
		baseURL: cfg.Analytics.URL,
	}
}

// GetAnomalies retrieves anomaly detection results
func (c *AnalyticsClient) GetAnomalies(ctx context.Context, deviceID string, authToken string) (interface{}, error) {
	url := fmt.Sprintf("%s/analytics/anomalies?deviceId=%s", c.baseURL, deviceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("analytics service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("analytics service error: %s", apiResp.Error.Message)
	}

	return apiResp.Data, nil
}
