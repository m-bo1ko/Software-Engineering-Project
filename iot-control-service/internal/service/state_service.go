package service

import (
	"context"
	"time"

	"iot-control-service/internal/models"
	"iot-control-service/internal/repository"
)

// StateService handles device state business logic
type StateService struct {
	deviceRepo    *repository.DeviceRepository
	telemetryRepo *repository.TelemetryRepository
}

// NewStateService creates a new state service
func NewStateService(
	deviceRepo *repository.DeviceRepository,
	telemetryRepo *repository.TelemetryRepository,
) *StateService {
	return &StateService{
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
	}
}

// GetLiveState retrieves live state for all devices
func (s *StateService) GetLiveState(ctx context.Context) (*models.LiveStateResponse, error) {
	// Get all online devices
	devices, _, err := s.deviceRepo.FindAll(ctx, "", "", "ONLINE", 1, 1000)
	if err != nil {
		return nil, err
	}

	deviceIDs := make([]string, len(devices))
	for i, d := range devices {
		deviceIDs[i] = d.DeviceID
	}

	// Get latest telemetry for all devices
	latestTelemetry, err := s.telemetryRepo.FindLatestMetricsByDevice(ctx, deviceIDs)
	if err != nil {
		return nil, err
	}

	// Build state response
	states := make([]models.DeviceState, 0, len(devices))
	for _, device := range devices {
		state := models.DeviceState{
			DeviceID:   device.DeviceID,
			Status:     string(device.Status),
			LastSeen:   device.LastSeen,
			LastUpdate: device.UpdatedAt,
		}

		if telemetry, exists := latestTelemetry[device.DeviceID]; exists {
			state.Metrics = telemetry.Metrics
		} else {
			state.Metrics = make(map[string]interface{})
		}

		states = append(states, state)
	}

	// Determine last update time
	var lastUpdate time.Time
	if len(states) > 0 {
		lastUpdate = states[0].LastUpdate
	} else {
		lastUpdate = time.Now()
	}

	return &models.LiveStateResponse{
		Devices: states,
		Count:   len(states),
		Updated: lastUpdate,
	}, nil
}

// GetDeviceState retrieves state for a specific device
func (s *StateService) GetDeviceState(ctx context.Context, deviceID string) (*models.DeviceState, error) {
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	telemetry, err := s.telemetryRepo.FindLatestByDevice(ctx, deviceID)
	if err != nil {
		// Device exists but no telemetry yet
		return &models.DeviceState{
			DeviceID:   device.DeviceID,
			Status:     string(device.Status),
			LastSeen:   device.LastSeen,
			Metrics:    make(map[string]interface{}),
			LastUpdate: device.UpdatedAt,
		}, nil
	}

	return &models.DeviceState{
		DeviceID:   device.DeviceID,
		Status:     string(device.Status),
		LastSeen:   device.LastSeen,
		Metrics:    telemetry.Metrics,
		LastUpdate: telemetry.Timestamp,
	}, nil
}
