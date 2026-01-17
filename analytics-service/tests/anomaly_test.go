package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// MockAnomalyRepository is a mock implementation for testing
type MockAnomalyRepository struct {
	anomalies map[string]*models.Anomaly
}

func (m *MockAnomalyRepository) Create(ctx context.Context, anomaly *models.Anomaly) (*models.Anomaly, error) {
	if m.anomalies == nil {
		m.anomalies = make(map[string]*models.Anomaly)
	}
	m.anomalies[anomaly.AnomalyID] = anomaly
	return anomaly, nil
}

func (m *MockAnomalyRepository) FindByAnomalyID(ctx context.Context, anomalyID string) (*models.Anomaly, error) {
	if anomaly, exists := m.anomalies[anomalyID]; exists {
		return anomaly, nil
	}
	return nil, errors.New("anomaly not found")
}

func (m *MockAnomalyRepository) FindAll(ctx context.Context, deviceID, buildingID, anomalyType, severity, status string, page, limit int) ([]*models.Anomaly, int64, error) {
	results := make([]*models.Anomaly, 0)
	for _, anomaly := range m.anomalies {
		results = append(results, anomaly)
	}
	return results, int64(len(results)), nil
}

func (m *MockAnomalyRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	count := int64(0)
	for _, anomaly := range m.anomalies {
		if string(anomaly.Status) == status {
			count++
		}
	}
	return count, nil
}

// MockIoTClientForAnomaly is a mock implementation for testing
type MockIoTClientForAnomaly struct{}

func (m *MockIoTClientForAnomaly) GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"deviceId":  deviceID,
			"timestamp": time.Now().Format(time.RFC3339),
			"metrics": map[string]interface{}{
				"temperature": 35.0, // Anomaly: above threshold
				"consumption": 500.0,
			},
		},
	}, nil
}

func (m *MockIoTClientForAnomaly) GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

// TestAnomalyDetection tests anomaly detection
func TestAnomalyDetection(t *testing.T) {
	// Setup mocks
	mockAnomalyRepo := &MockAnomalyRepository{}
	mockIoTClient := &MockIoTClientForAnomaly{}

	// Create service
	anomalyService := service.NewAnomalyService(mockAnomalyRepo, mockIoTClient)

	// Test anomaly detection
	ctx := context.Background()
	anomalies, err := anomalyService.DetectAnomalies(ctx, "device-001", "building-001", "token")
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	if len(anomalies) == 0 {
		t.Error("Expected at least one anomaly to be detected")
	}

	// Check anomaly properties
	anomaly := anomalies[0]
	if anomaly.DeviceID != "device-001" {
		t.Errorf("Expected device ID device-001, got %s", anomaly.DeviceID)
	}

	if anomaly.Severity != string(models.AnomalySeverityHigh) {
		t.Errorf("Expected severity HIGH, got %s", anomaly.Severity)
	}
}
