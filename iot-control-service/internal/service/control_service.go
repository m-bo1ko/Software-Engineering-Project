package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"iot-control-service/internal/models"
	"iot-control-service/internal/mqtt"
	"iot-control-service/internal/repository"
)

// ControlService handles device control business logic
type ControlService struct {
	commandRepo *repository.CommandRepository
	deviceRepo  *repository.DeviceRepository
	mqttClient  *mqtt.Client
	config      interface {
		GetCommandTimeout() time.Duration
	}
}

// NewControlService creates a new control service
func NewControlService(
	commandRepo *repository.CommandRepository,
	deviceRepo *repository.DeviceRepository,
	mqttClient *mqtt.Client,
	commandTimeout time.Duration,
) *ControlService {
	return &ControlService{
		commandRepo: commandRepo,
		deviceRepo:  deviceRepo,
		mqttClient:  mqttClient,
		config:      &configWrapper{timeout: commandTimeout},
	}
}

type configWrapper struct {
	timeout time.Duration
}

func (c *configWrapper) GetCommandTimeout() time.Duration {
	return c.timeout
}

// SendCommand sends a command to a device
func (s *ControlService) SendCommand(ctx context.Context, deviceID string, req *models.SendCommandRequest, userID string) (*models.CommandResponse, error) {
	// Validate device exists
	_, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Validate command
	if err := s.validateCommand(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate command ID
	commandID := uuid.New().String()

	// Create command record
	command := &models.DeviceCommand{
		CommandID: commandID,
		DeviceID:  deviceID,
		Command:   req.Command,
		Params:    req.Params,
		Status:    models.CommandStatusPending,
		IssuedBy:  userID,
	}

	createdCommand, err := s.commandRepo.Create(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("failed to create command: %w", err)
	}

	// Publish command to MQTT
	if err := s.mqttClient.PublishCommand(deviceID, createdCommand); err != nil {
		// Update command status to failed
		s.commandRepo.UpdateStatus(ctx, commandID, models.CommandStatusFailed, fmt.Sprintf("MQTT publish failed: %v", err))
		return nil, fmt.Errorf("failed to publish command: %w", err)
	}

	// Update command status to sent
	s.commandRepo.UpdateStatus(ctx, commandID, models.CommandStatusSent, "")

	// Refresh command from DB
	updatedCommand, err := s.commandRepo.FindByCommandID(ctx, commandID)
	if err != nil {
		return createdCommand.ToResponse(), nil
	}

	return updatedCommand.ToResponse(), nil
}

// GetCommand retrieves a command by ID
func (s *ControlService) GetCommand(ctx context.Context, commandID string) (*models.CommandResponse, error) {
	command, err := s.commandRepo.FindByCommandID(ctx, commandID)
	if err != nil {
		return nil, err
	}
	return command.ToResponse(), nil
}

// ListCommands lists commands for a device
func (s *ControlService) ListCommands(ctx context.Context, deviceID string, status string, page, limit int) ([]*models.CommandResponse, int64, error) {
	commands, total, err := s.commandRepo.FindByDeviceID(ctx, deviceID, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.CommandResponse, len(commands))
	for i, cmd := range commands {
		responses[i] = cmd.ToResponse()
	}

	return responses, total, nil
}

// ProcessCommandAck processes a command acknowledgment from a device
func (s *ControlService) ProcessCommandAck(ctx context.Context, ack *models.CommandAck) error {
	_, err := s.commandRepo.FindByCommandID(ctx, ack.CommandID)
	if err != nil {
		return fmt.Errorf("command not found: %w", err)
	}

	status := models.CommandStatusApplied
	if ack.Status == "FAILED" {
		status = models.CommandStatusFailed
	}

	return s.commandRepo.UpdateStatus(ctx, ack.CommandID, status, ack.ErrorMsg)
}

// validateCommand validates a command request
func (s *ControlService) validateCommand(req *models.SendCommandRequest) error {
	if req.Command == "" {
		return fmt.Errorf("command is required")
	}
	return nil
}
