package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"analytics-service/internal/models"
	"analytics-service/internal/repository"
)

// AnomalyService handles anomaly detection business logic
type AnomalyService struct {
	anomalyRepo *repository.AnomalyRepository
	iotClient   interface {
		GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error)
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	}
}

// NewAnomalyService creates a new anomaly service
func NewAnomalyService(
	anomalyRepo *repository.AnomalyRepository,
	iotClient interface {
		GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error)
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	},
) *AnomalyService {
	return &AnomalyService{
		anomalyRepo: anomalyRepo,
		iotClient:   iotClient,
	}
}

// DetectAnomalies detects anomalies in telemetry data
func (s *AnomalyService) DetectAnomalies(ctx context.Context, deviceID, buildingID string, authToken string) ([]*models.AnomalyResponse, error) {
	// Get recent telemetry
	to := time.Now()
	from := to.Add(-24 * time.Hour) // Last 24 hours

	telemetry, err := s.iotClient.GetTelemetryHistory(ctx, deviceID, from, to, 1, 100, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry: %w", err)
	}

	anomalies := make([]*models.Anomaly, 0)

	// Simple anomaly detection: check for values outside normal range
	for _, t := range telemetry {
		if metrics, ok := t["metrics"].(map[string]interface{}); ok {
			// Check temperature anomalies
			if temp, ok := metrics["temperature"].(float64); ok {
				if temp > 30.0 || temp < 10.0 {
					anomaly := s.createAnomaly(deviceID, buildingID, "TEMPERATURE_OUT_OF_RANGE", models.AnomalySeverityHigh, map[string]interface{}{
						"temperature": temp,
						"threshold":   "10-30°C",
					})
					anomalies = append(anomalies, anomaly)
				}
			}

			// Check consumption spikes
			if consumption, ok := metrics["consumption"].(float64); ok {
				if consumption > 1000.0 { // Threshold example
					anomaly := s.createAnomaly(deviceID, buildingID, "CONSUMPTION_SPIKE", models.AnomalySeverityMedium, map[string]interface{}{
						"consumption": consumption,
						"threshold":   1000.0,
					})
					anomalies = append(anomalies, anomaly)
				}
			}
		}
	}

	// Save anomalies
	responses := make([]*models.AnomalyResponse, 0)
	for _, anomaly := range anomalies {
		created, err := s.anomalyRepo.Create(ctx, anomaly)
		if err != nil {
			continue
		}
		responses = append(responses, created.ToResponse())
	}

	return responses, nil
}

// createAnomaly creates an anomaly record
func (s *AnomalyService) createAnomaly(deviceID, buildingID, anomalyType string, severity models.AnomalySeverity, details map[string]interface{}) *models.Anomaly {
	return &models.Anomaly{
		AnomalyID:  uuid.New().String(),
		DeviceID:   deviceID,
		BuildingID: buildingID,
		Type:       anomalyType,
		Severity:   severity,
		Status:     models.AnomalyStatusNew,
		Details:    details,
		DetectedAt: time.Now(),
	}
}

// GetAnomaly retrieves an anomaly by ID
func (s *AnomalyService) GetAnomaly(ctx context.Context, anomalyID string) (*models.AnomalyResponse, error) {
	anomaly, err := s.anomalyRepo.FindByAnomalyID(ctx, anomalyID)
	if err != nil {
		return nil, err
	}
	return anomaly.ToResponse(), nil
}

// ListAnomalies lists anomalies with filters
func (s *AnomalyService) ListAnomalies(ctx context.Context, deviceID, buildingID, anomalyType, severity, status string, page, limit int) ([]*models.AnomalyResponse, int64, error) {
	anomalies, total, err := s.anomalyRepo.FindAll(ctx, deviceID, buildingID, anomalyType, severity, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.AnomalyResponse, len(anomalies))
	for i, anomaly := range anomalies {
		responses[i] = anomaly.ToResponse()
	}

	return responses, total, nil
}

// AcknowledgeAnomaly acknowledges an anomaly
func (s *AnomalyService) AcknowledgeAnomaly(ctx context.Context, anomalyID, userID string) (*models.AnomalyResponse, error) {
	anomaly, err := s.anomalyRepo.FindByAnomalyID(ctx, anomalyID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         models.AnomalyStatusAcknowledged,
		"acknowledged_at": now,
		"acknowledged_by": userID,
	}

	updated, err := s.anomalyRepo.Update(ctx, anomaly.ID.Hex(), updates)
	if err != nil {
		return nil, err
	}

	return updated.ToResponse(), nil
}
