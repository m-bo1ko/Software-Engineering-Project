package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// MockReportRepository is a mock implementation for testing
type MockReportRepository struct {
	reports map[string]*models.Report
}

func (m *MockReportRepository) Create(ctx context.Context, report *models.Report) (*models.Report, error) {
	if m.reports == nil {
		m.reports = make(map[string]*models.Report)
	}
	m.reports[report.ReportID] = report
	return report, nil
}

func (m *MockReportRepository) FindByReportID(ctx context.Context, reportID string) (*models.Report, error) {
	if report, exists := m.reports[reportID]; exists {
		return report, nil
	}
	return nil, errors.New("report not found")
}

// MockIoTClient is a mock implementation for testing
type MockIoTClient struct{}

func (m *MockIoTClient) GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"deviceId":  deviceID,
			"timestamp": time.Now().Format(time.RFC3339),
			"metrics": map[string]interface{}{
				"consumption": 100.0,
			},
		},
	}, nil
}

func (m *MockIoTClient) GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"deviceId": "device-001",
			"status":   "ONLINE",
		},
	}, nil
}

// MockForecastClient is a mock implementation for testing
type MockForecastClient struct{}

func (m *MockForecastClient) GetLatestForecast(ctx context.Context, buildingID string, authToken string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"buildingId": buildingID,
		"forecast":   "data",
	}, nil
}

// TestReportGeneration tests report generation
func TestReportGeneration(t *testing.T) {
	// Setup mocks
	mockReportRepo := &MockReportRepository{}
	mockIoTClient := &MockIoTClient{}
	mockForecastClient := &MockForecastClient{}

	// Create service
	reportService := service.NewReportService(mockReportRepo, mockIoTClient, mockForecastClient)

	// Test report generation
	req := &models.GenerateReportRequest{
		Type:       "ENERGY_CONSUMPTION",
		BuildingID: "building-001",
		From:       time.Now().AddDate(0, 0, -30),
		To:         time.Now(),
	}

	ctx := context.Background()
	response, err := reportService.GenerateReport(ctx, req, "user-001", "token")
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	if response.ReportID == "" {
		t.Error("Expected report ID to be generated")
	}

	if response.Status != string(models.ReportStatusGenerating) {
		t.Errorf("Expected status GENERATING, got %s", response.Status)
	}
}
