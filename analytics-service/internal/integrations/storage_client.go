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

	"analytics-service/internal/config"
	"analytics-service/internal/models"
)

// StorageClient handles communication with the external Storage service
// as per integration contract requirements:
// - /storage/analytics/*
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

// SaveReport saves a report to the storage service
// POST /storage/analytics/reports
func (c *StorageClient) SaveReport(ctx context.Context, report *models.Report, authToken string) error {
	jsonData, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analytics/reports", bytes.NewBuffer(jsonData))
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

// SaveAnomaly saves an anomaly to the storage service
// POST /storage/analytics/anomalies
func (c *StorageClient) SaveAnomaly(ctx context.Context, anomaly *models.Anomaly, authToken string) error {
	jsonData, err := json.Marshal(anomaly)
	if err != nil {
		return fmt.Errorf("failed to marshal anomaly: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analytics/anomalies", bytes.NewBuffer(jsonData))
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

// SaveTimeSeries saves time series data to the storage service
// POST /storage/analytics/timeseries
func (c *StorageClient) SaveTimeSeries(ctx context.Context, timeseries *models.TimeSeries, authToken string) error {
	jsonData, err := json.Marshal(timeseries)
	if err != nil {
		return fmt.Errorf("failed to marshal timeseries: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analytics/timeseries", bytes.NewBuffer(jsonData))
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

// SaveKPI saves KPI data to the storage service
// POST /storage/analytics/kpi
func (c *StorageClient) SaveKPI(ctx context.Context, kpi *models.KPI, authToken string) error {
	jsonData, err := json.Marshal(kpi)
	if err != nil {
		return fmt.Errorf("failed to marshal kpi: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analytics/kpi", bytes.NewBuffer(jsonData))
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

// GetAnalyticsData retrieves analytics data from the storage service
// GET /storage/analytics/{type}
func (c *StorageClient) GetAnalyticsData(ctx context.Context, dataType string, buildingID string, from, to time.Time, authToken string) (interface{}, error) {
	reqURL := fmt.Sprintf("%s/analytics/%s?from=%s&to=%s",
		c.baseURL,
		url.PathEscape(dataType),
		url.QueryEscape(from.Format(time.RFC3339)),
		url.QueryEscape(to.Format(time.RFC3339)),
	)

	if buildingID != "" {
		reqURL += "&buildingId=" + url.QueryEscape(buildingID)
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
		return nil, fmt.Errorf("storage service returned status: %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("storage service error: %s", apiResp.Error.Message)
	}

	return apiResp.Data, nil
}
