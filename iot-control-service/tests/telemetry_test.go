package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"iot-control-service/internal/models"
	"iot-control-service/internal/service"
)

// MockTelemetryRepository is a mock implementation for testing
type MockTelemetryRepository struct {
	telemetry []*models.Telemetry
}

func (m *MockTelemetryRepository) Create(ctx context.Context, telemetry *models.Telemetry) (*models.Telemetry, error) {
	m.telemetry = append(m.telemetry, telemetry)
	return telemetry, nil
}

func (m *MockTelemetryRepository) CreateMany(ctx context.Context, telemetry []*models.Telemetry) error {
	m.telemetry = append(m.telemetry, telemetry...)
	return nil
}

// MockDeviceRepository is a mock implementation for testing
type MockDeviceRepository struct {
	devices map[string]*models.Device
}

func (m *MockDeviceRepository) FindByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	if device, exists := m.devices[deviceID]; exists {
		return device, nil
	}
	return nil, errors.New("device not found")
}

func (m *MockDeviceRepository) UpdateLastSeen(ctx context.Context, deviceID string) error {
	return nil
}

// TestTelemetryIngestion tests telemetry ingestion
func TestTelemetryIngestion(t *testing.T) {
	// Setup mocks
	mockTelemetryRepo := &MockTelemetryRepository{}
	mockDeviceRepo := &MockDeviceRepository{
		devices: map[string]*models.Device{
			"device-001": {
				DeviceID: "device-001",
				Type:     "HVAC",
				Status:   models.DeviceStatusOnline,
			},
		},
	}

	// Create service
	telemetryService := service.NewTelemetryService(mockTelemetryRepo, mockDeviceRepo)

	// Test single telemetry ingestion
	req := &models.TelemetryIngestRequest{
		DeviceID:  "device-001",
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"temperature": 22.5,
			"humidity":    45.0,
		},
	}

	ctx := context.Background()
	response, err := telemetryService.IngestTelemetry(ctx, req, "HTTP")
	if err != nil {
		t.Fatalf("Failed to ingest telemetry: %v", err)
	}

	if response.DeviceID != "device-001" {
		t.Errorf("Expected device ID device-001, got %s", response.DeviceID)
	}

	if len(mockTelemetryRepo.telemetry) != 1 {
		t.Errorf("Expected 1 telemetry record, got %d", len(mockTelemetryRepo.telemetry))
	}
}
