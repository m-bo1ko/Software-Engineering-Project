package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"iot-control-service/internal/models"
	"iot-control-service/internal/repository"
)

// OptimizationService handles optimization scenario business logic
type OptimizationService struct {
	optimizationRepo *repository.OptimizationRepository
	commandRepo      *repository.CommandRepository
	deviceRepo       *repository.DeviceRepository
}

// NewOptimizationService creates a new optimization service
func NewOptimizationService(
	optimizationRepo *repository.OptimizationRepository,
	commandRepo *repository.CommandRepository,
	deviceRepo *repository.DeviceRepository,
) *OptimizationService {
	return &OptimizationService{
		optimizationRepo: optimizationRepo,
		commandRepo:      commandRepo,
		deviceRepo:       deviceRepo,
	}
}

// ApplyOptimization applies an optimization scenario
func (s *OptimizationService) ApplyOptimization(ctx context.Context, req *models.ApplyOptimizationRequest, userID string) (*models.OptimizationScenarioResponse, error) {
	// Validate request
	if err := s.validateApplyOptimization(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate scenario ID
	scenarioID := uuid.New().String()

	// Create scenario
	scenario := &models.OptimizationScenario{
		ScenarioID:      scenarioID,
		ForecastID:      req.ForecastID,
		BuildingID:      req.BuildingID,
		Actions:         req.Actions,
		ExecutionStatus: models.OptimizationStatusPending,
		Progress:        0.0,
		CreatedBy:       userID,
	}

	createdScenario, err := s.optimizationRepo.Create(ctx, scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to create scenario: %w", err)
	}

	// Start execution asynchronously
	go s.executeScenario(context.Background(), createdScenario)

	return createdScenario.ToResponse(), nil
}

// GetOptimizationStatus retrieves the status of an optimization scenario
func (s *OptimizationService) GetOptimizationStatus(ctx context.Context, scenarioID string) (*models.OptimizationScenarioResponse, error) {
	scenario, err := s.optimizationRepo.FindByScenarioID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	return scenario.ToResponse(), nil
}

// executeScenario executes an optimization scenario
func (s *OptimizationService) executeScenario(ctx context.Context, scenario *models.OptimizationScenario) {
	// Update status to running
	s.optimizationRepo.UpdateProgress(ctx, scenario.ScenarioID, 0.0, models.OptimizationStatusRunning)

	totalActions := float64(len(scenario.Actions))
	completedActions := 0.0

	// Execute each action
	for i, action := range scenario.Actions {
		// Validate device exists
		_, err := s.deviceRepo.FindByDeviceID(ctx, action.DeviceID)
		if err != nil {
			// Update action status to failed
			s.updateActionStatus(ctx, scenario.ScenarioID, action.DeviceID, "FAILED", "")
			completedActions++
			continue
		}

		// Create command
		commandID := uuid.New().String()
		command := &models.DeviceCommand{
			CommandID: commandID,
			DeviceID:  action.DeviceID,
			Command:   action.Command,
			Params:    action.Params,
			Status:    models.CommandStatusPending,
			IssuedBy:  scenario.CreatedBy,
		}

		_, err = s.commandRepo.Create(ctx, command)
		if err != nil {
			s.updateActionStatus(ctx, scenario.ScenarioID, action.DeviceID, "FAILED", "")
			completedActions++
			continue
		}

		// Update action with command ID
		s.updateActionStatus(ctx, scenario.ScenarioID, action.DeviceID, "SENT", commandID)

		// Wait for command to be applied (simplified - in production, use proper async handling)
		time.Sleep(1 * time.Second)

		// Check command status
		cmd, err := s.commandRepo.FindByCommandID(ctx, commandID)
		if err == nil {
			if cmd.Status == models.CommandStatusApplied {
				s.updateActionStatus(ctx, scenario.ScenarioID, action.DeviceID, "APPLIED", commandID)
			} else if cmd.Status == models.CommandStatusFailed {
				s.updateActionStatus(ctx, scenario.ScenarioID, action.DeviceID, "FAILED", commandID)
			}
		}

		completedActions++
		progress := completedActions / totalActions
		s.optimizationRepo.UpdateProgress(ctx, scenario.ScenarioID, progress, models.OptimizationStatusRunning)
	}

	// Mark scenario as completed
	s.optimizationRepo.UpdateProgress(ctx, scenario.ScenarioID, 1.0, models.OptimizationStatusCompleted)
}

// updateActionStatus updates the status of an action in a scenario
func (s *OptimizationService) updateActionStatus(ctx context.Context, scenarioID, deviceID, status, commandID string) {
	s.optimizationRepo.UpdateActionStatus(ctx, scenarioID, deviceID, status, commandID)
}

// validateApplyOptimization validates an apply optimization request
func (s *OptimizationService) validateApplyOptimization(req *models.ApplyOptimizationRequest) error {
	if req.ScenarioID == "" {
		return fmt.Errorf("scenario ID is required")
	}
	if req.BuildingID == "" {
		return fmt.Errorf("building ID is required")
	}
	if len(req.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}
	return nil
}
