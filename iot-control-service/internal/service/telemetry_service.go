package service

import (
	"context"
	"fmt"
	"time"

	"iot-control-service/internal/models"
	"iot-control-service/internal/repository"
)

// TelemetryService handles telemetry business logic
type TelemetryService struct {
	telemetryRepo *repository.TelemetryRepository
	deviceRepo    *repository.DeviceRepository
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(
	telemetryRepo *repository.TelemetryRepository,
	deviceRepo *repository.DeviceRepository,
) *TelemetryService {
	return &TelemetryService{
		telemetryRepo: telemetryRepo,
		deviceRepo:    deviceRepo,
	}
}

// IngestTelemetry ingests a single telemetry message
func (s *TelemetryService) IngestTelemetry(ctx context.Context, req *models.TelemetryIngestRequest, source string) (*models.TelemetryResponse, error) {
	// Validate device exists
	_, err := s.deviceRepo.FindByDeviceID(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Create telemetry record
	telemetry := &models.Telemetry{
		DeviceID:  req.DeviceID,
		Timestamp: req.Timestamp,
		Metrics:   req.Metrics,
		Source:    source,
	}

	if telemetry.Timestamp.IsZero() {
		telemetry.Timestamp = time.Now()
	}

	createdTelemetry, err := s.telemetryRepo.Create(ctx, telemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry: %w", err)
	}

	// Update device last seen
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.deviceRepo.UpdateLastSeen(bgCtx, req.DeviceID)
	}()

	return createdTelemetry.ToResponse(), nil
}

// IngestBulkTelemetry ingests multiple telemetry messages
func (s *TelemetryService) IngestBulkTelemetry(ctx context.Context, req *models.BulkTelemetryIngestRequest, source string) ([]*models.TelemetryResponse, error) {
	if len(req.Telemetry) == 0 {
		return []*models.TelemetryResponse{}, nil
	}

	telemetryList := make([]*models.Telemetry, 0, len(req.Telemetry))
	deviceIDs := make(map[string]bool)

	now := time.Now()
	for _, t := range req.Telemetry {
		// Validate device exists
		if _, exists := deviceIDs[t.DeviceID]; !exists {
			_, err := s.deviceRepo.FindByDeviceID(ctx, t.DeviceID)
			if err != nil {
				return nil, fmt.Errorf("device %s not found: %w", t.DeviceID, err)
			}
			deviceIDs[t.DeviceID] = true
		}

		telemetry := &models.Telemetry{
			DeviceID:  t.DeviceID,
			Timestamp: t.Timestamp,
			Metrics:   t.Metrics,
			Source:    source,
		}

		if telemetry.Timestamp.IsZero() {
			telemetry.Timestamp = now
		}

		telemetryList = append(telemetryList, telemetry)
	}

	// Bulk insert
	err := s.telemetryRepo.CreateMany(ctx, telemetryList)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry: %w", err)
	}

	// Update device last seen for all devices
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for deviceID := range deviceIDs {
			s.deviceRepo.UpdateLastSeen(bgCtx, deviceID)
		}
	}()

	responses := make([]*models.TelemetryResponse, len(telemetryList))
	for i, t := range telemetryList {
		responses[i] = t.ToResponse()
	}

	return responses, nil
}

// GetTelemetryHistory retrieves telemetry history for a device
func (s *TelemetryService) GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int) ([]*models.TelemetryResponse, int64, error) {
	telemetry, total, err := s.telemetryRepo.FindByDeviceID(ctx, deviceID, from, to, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.TelemetryResponse, len(telemetry))
	for i, t := range telemetry {
		responses[i] = t.ToResponse()
	}

	return responses, total, nil
}

// GetLatestTelemetry retrieves the latest telemetry for a device
func (s *TelemetryService) GetLatestTelemetry(ctx context.Context, deviceID string) (*models.TelemetryResponse, error) {
	telemetry, err := s.telemetryRepo.FindLatestByDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return telemetry.ToResponse(), nil
}
