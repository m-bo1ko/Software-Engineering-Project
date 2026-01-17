package tests

import (
	"context"
	"errors"
	"testing"

	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// MockKPIRepository is a mock implementation for testing
type MockKPIRepository struct {
	kpis map[string]*models.KPI
}

func (m *MockKPIRepository) FindLatest(ctx context.Context, buildingID, period string) (*models.KPI, error) {
	key := buildingID + ":" + period
	if kpi, exists := m.kpis[key]; exists {
		return kpi, nil
	}
	return nil, errors.New("KPI not found")
}

func (m *MockKPIRepository) UpdateOrCreate(ctx context.Context, kpi *models.KPI) (*models.KPI, error) {
	if m.kpis == nil {
		m.kpis = make(map[string]*models.KPI)
	}
	key := kpi.BuildingID + ":" + kpi.Period
	m.kpis[key] = kpi
	return kpi, nil
}

// MockAnomalyRepositoryForKPI is a mock implementation for testing
type MockAnomalyRepositoryForKPI struct{}

func (m *MockAnomalyRepositoryForKPI) CountByStatus(ctx context.Context, status string) (int64, error) {
	return 5, nil
}

// MockIoTClientForKPI is a mock implementation for testing
type MockIoTClientForKPI struct{}

func (m *MockIoTClientForKPI) GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"deviceId": "device-001", "status": "ONLINE"},
		{"deviceId": "device-002", "status": "ONLINE"},
		{"deviceId": "device-003", "status": "OFFLINE"},
	}, nil
}

// TestKPICalculation tests KPI calculation
func TestKPICalculation(t *testing.T) {
	// Setup mocks
	mockKPIRepo := &MockKPIRepository{}
	mockAnomalyRepo := &MockAnomalyRepositoryForKPI{}
	mockIoTClient := &MockIoTClientForKPI{}

	// Create service
	kpiService := service.NewKPIService(mockKPIRepo, mockAnomalyRepo, mockIoTClient)

	// Test KPI calculation
	ctx := context.Background()
	response, err := kpiService.CalculateKPIs(ctx, "building-001", "DAILY", "token")
	if err != nil {
		t.Fatalf("Failed to calculate KPIs: %v", err)
	}

	if response.Metrics == nil {
		t.Error("Expected metrics to be calculated")
	}

	if totalDevices, ok := response.Metrics["totalDevices"].(int); !ok || totalDevices != 3 {
		t.Errorf("Expected 3 total devices, got %v", totalDevices)
	}

	if onlineDevices, ok := response.Metrics["onlineDevices"].(int); !ok || onlineDevices != 2 {
		t.Errorf("Expected 2 online devices, got %v", onlineDevices)
	}
}
