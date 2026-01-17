package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"forecast-service/internal/config"
	"forecast-service/internal/models"
)

// ExternalClient handles communication with external APIs (weather, tariffs, ML, storage)
type ExternalClient struct {
	httpClient *http.Client
	weatherURL string
	tariffURL  string
	mlURL      string
	storageURL string
}

// NewExternalClient creates a new external client
func NewExternalClient(cfg *config.Config) *ExternalClient {
	return &ExternalClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		weatherURL: cfg.External.WeatherURL,
		tariffURL:  cfg.External.TariffURL,
		mlURL:      cfg.External.MLURL,
		storageURL: cfg.External.StorageURL,
	}
}

// GetCurrentWeather retrieves current weather for a building location
func (c *ExternalClient) GetCurrentWeather(ctx context.Context, buildingID string, authToken string) (*models.Weather, error) {
	reqURL := fmt.Sprintf("%s/current?buildingId=%s", c.weatherURL, url.QueryEscape(buildingID))

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
		return nil, fmt.Errorf("weather API returned status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Success bool           `json:"success"`
		Data    models.Weather `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp.Data, nil
}

// GetWeatherForecast retrieves weather forecast for a building location
func (c *ExternalClient) GetWeatherForecast(ctx context.Context, buildingID string, hours int, authToken string) ([]WeatherForecastPoint, error) {
	reqURL := fmt.Sprintf("%s/forecast?buildingId=%s&hours=%d", c.weatherURL, url.QueryEscape(buildingID), hours)

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
		return nil, fmt.Errorf("weather forecast API returned status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    []WeatherForecastPoint `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.Data, nil
}

// WeatherForecastPoint represents a point in weather forecast
type WeatherForecastPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	CloudCover  float64   `json:"cloudCover"`
	WindSpeed   float64   `json:"windSpeed"`
	Condition   string    `json:"condition"`
}

// GetCurrentTariff retrieves current tariff for a region
func (c *ExternalClient) GetCurrentTariff(ctx context.Context, region string, authToken string) (*models.Tariff, error) {
	reqURL := fmt.Sprintf("%s/current?region=%s", c.tariffURL, url.QueryEscape(region))

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
		return nil, fmt.Errorf("tariff API returned status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Success bool          `json:"success"`
		Data    models.Tariff `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp.Data, nil
}

// GetHistoricalConsumption retrieves historical consumption data
func (c *ExternalClient) GetHistoricalConsumption(ctx context.Context, buildingID, deviceID string, from, to time.Time, resolution string, authToken string) (*models.HistoricalConsumption, error) {
	reqURL := fmt.Sprintf("%s/consumption/history?buildingId=%s&from=%s&to=%s&resolution=%s",
		c.storageURL,
		url.QueryEscape(buildingID),
		url.QueryEscape(from.Format(time.RFC3339)),
		url.QueryEscape(to.Format(time.RFC3339)),
		url.QueryEscape(resolution),
	)

	if deviceID != "" {
		reqURL += "&deviceId=" + url.QueryEscape(deviceID)
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
		return nil, fmt.Errorf("storage API returned status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Success bool                         `json:"success"`
		Data    models.HistoricalConsumption `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp.Data, nil
}

// MLPredictionRequest represents a request to the ML model
type MLPredictionRequest struct {
	BuildingID       string                          `json:"buildingId"`
	DeviceID         string                          `json:"deviceId,omitempty"`
	HistoricalData   []models.ConsumptionDataPoint   `json:"historicalData"`
	WeatherForecast  []WeatherForecastPoint          `json:"weatherForecast,omitempty"`
	TariffData       *models.Tariff                  `json:"tariffData,omitempty"`
	HorizonHours     int                             `json:"horizonHours"`
	ModelType        string                          `json:"modelType"` // LSTM, ARIMA, PROPHET, etc.
}

// MLPredictionResponse represents a response from the ML model
type MLPredictionResponse struct {
	Success     bool                       `json:"success"`
	Predictions []models.ForecastPrediction `json:"predictions"`
	ModelUsed   string                     `json:"modelUsed"`
	Accuracy    *models.ForecastAccuracy   `json:"accuracy,omitempty"`
	Error       string                     `json:"error,omitempty"`
}

// GetMLPrediction requests a prediction from the ML model service
func (c *ExternalClient) GetMLPrediction(ctx context.Context, request *MLPredictionRequest, authToken string) (*MLPredictionResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.mlURL, bytes.NewBuffer(jsonData))
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

	var result MLPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("ML prediction failed: %s", result.Error)
	}

	return &result, nil
}

// HealthCheck checks if external services are available
func (c *ExternalClient) HealthCheck(ctx context.Context) map[string]bool {
	results := make(map[string]bool)

	// Check weather service
	results["weather"] = c.checkHealth(ctx, c.weatherURL+"/health")

	// Check tariff service
	results["tariff"] = c.checkHealth(ctx, c.tariffURL+"/health")

	// Check ML service
	results["ml"] = c.checkHealth(ctx, c.mlURL+"/health")

	// Check storage service
	results["storage"] = c.checkHealth(ctx, c.storageURL+"/health")

	return results
}

func (c *ExternalClient) checkHealth(ctx context.Context, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
