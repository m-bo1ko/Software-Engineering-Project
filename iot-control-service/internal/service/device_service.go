// Package service contains business logic for the application
package service

import (
	"context"
	"fmt"
	"time"

	"iot-control-service/internal/models"
	"iot-control-service/internal/repository"
)

// DeviceService handles device business logic
type DeviceService struct {
	deviceRepo *repository.DeviceRepository
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceRepo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
	}
}

// RegisterDevice registers a new device
func (s *DeviceService) RegisterDevice(ctx context.Context, req *models.RegisterDeviceRequest, userID string) (*models.DeviceResponse, error) {
	// Validate request
	if err := s.validateRegisterDevice(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if device already exists
	_, err := s.deviceRepo.FindByDeviceID(ctx, req.DeviceID)
	if err == nil {
		return nil, fmt.Errorf("device with ID %s already exists", req.DeviceID)
	}

	// Create device
	device := &models.Device{
		DeviceID:     req.DeviceID,
		Type:         req.Type,
		Model:        req.Model,
		Location:     req.Location,
		Capabilities: req.Capabilities,
		Status:       models.DeviceStatusOffline,
		LastSeen:     time.Time{},
		Metadata:     req.Metadata,
		CreatedBy:    userID,
	}

	createdDevice, err := s.deviceRepo.Create(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return createdDevice.ToResponse(), nil
}

// GetDevice retrieves a device by ID
func (s *DeviceService) GetDevice(ctx context.Context, deviceID string) (*models.DeviceResponse, error) {
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return device.ToResponse(), nil
}

// ListDevices lists devices with filters
func (s *DeviceService) ListDevices(ctx context.Context, buildingID, deviceType, status string, page, limit int) ([]*models.DeviceResponse, int64, error) {
	devices, total, err := s.deviceRepo.FindAll(ctx, buildingID, deviceType, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.DeviceResponse, len(devices))
	for i, device := range devices {
		responses[i] = device.ToResponse()
	}

	return responses, total, nil
}

// UpdateDevice updates a device
func (s *DeviceService) UpdateDevice(ctx context.Context, deviceID string, updates map[string]interface{}) (*models.DeviceResponse, error) {
	// Find device first to get MongoDB ID
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	updatedDevice, err := s.deviceRepo.Update(ctx, device.ID.Hex(), updates)
	if err != nil {
		return nil, err
	}

	return updatedDevice.ToResponse(), nil
}

// DeleteDevice deletes a device
func (s *DeviceService) DeleteDevice(ctx context.Context, deviceID string) error {
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return err
	}

	return s.deviceRepo.Delete(ctx, device.ID.Hex())
}

// UpdateDeviceLastSeen updates the last seen timestamp for a device
func (s *DeviceService) UpdateDeviceLastSeen(ctx context.Context, deviceID string) error {
	return s.deviceRepo.UpdateLastSeen(ctx, deviceID)
}

// validateRegisterDevice validates device registration request
func (s *DeviceService) validateRegisterDevice(req *models.RegisterDeviceRequest) error {
	if req.DeviceID == "" {
		return fmt.Errorf("device ID is required")
	}
	if req.Type == "" {
		return fmt.Errorf("device type is required")
	}
	if req.Location.BuildingID == "" {
		return fmt.Errorf("building ID is required")
	}
	return nil
}
