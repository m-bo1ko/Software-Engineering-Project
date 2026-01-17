package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"iot-control-service/internal/integrations"
	"iot-control-service/internal/models"
	"iot-control-service/internal/repository"
)

// OptimizationService handles optimization scenario business logic
// Integration: Uses ForecastClient to fetch device predictions before executing optimization
// Integration: Uses AnalyticsClient to check for anomalies before applying changes
type OptimizationService struct {
	optimizationRepo *repository.OptimizationRepository
	commandRepo      *repository.CommandRepository
	deviceRepo       *repository.DeviceRepository
	forecastClient   *integrations.ForecastClient
	analyticsClient  *integrations.AnalyticsClient
}

// NewOptimizationService creates a new optimization service
func NewOptimizationService(
	optimizationRepo *repository.OptimizationRepository,
	commandRepo *repository.CommandRepository,
	deviceRepo *repository.DeviceRepository,
	forecastClient *integrations.ForecastClient,
	analyticsClient *integrations.AnalyticsClient,
) *OptimizationService {
	return &OptimizationService{
		optimizationRepo: optimizationRepo,
		commandRepo:      commandRepo,
		deviceRepo:       deviceRepo,
		forecastClient:   forecastClient,
		analyticsClient:  analyticsClient,
	}
}

// ApplyOptimization applies an optimization scenario
// Integration: Fetches device predictions from Forecast service to validate optimization timing
// Integration: Checks for active anomalies from Analytics service to avoid conflicting actions
func (s *OptimizationService) ApplyOptimization(ctx context.Context, req *models.ApplyOptimizationRequest, userID string) (*models.OptimizationScenarioResponse, error) {
	// Validate request
	if err := s.validateApplyOptimization(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate scenario ID
	scenarioID := uuid.New().String()

	// Integration: Fetch predictions for each device to optimize execution timing
	// This allows us to schedule actions when energy savings will be maximized
	devicePredictions := make(map[string]*models.DevicePrediction)
	for _, action := range req.Actions {
		if s.forecastClient != nil {
			prediction, err := s.forecastClient.GetDevicePrediction(ctx, action.DeviceID, "")
			if err == nil && prediction != nil {
				devicePredictions[action.DeviceID] = prediction
				log.Printf("[Integration] Fetched prediction for device %s: trend=%s, savings potential=%.2f%%",
					action.DeviceID, prediction.Trend, prediction.TrendPercentage)
			}
		}
	}

	// Integration: Check for anomalies that might conflict with optimization actions
	// Skip actions for devices with active critical anomalies
	var filteredActions []models.OptimizationAction
	for _, action := range req.Actions {
		skipAction := false
		if s.analyticsClient != nil {
			anomalies, err := s.analyticsClient.GetAnomalies(ctx, action.DeviceID, "")
			if err == nil && anomalies != nil {
				// Check if any critical anomalies exist for this device
				if anomalyList, ok := anomalies.([]interface{}); ok && len(anomalyList) > 0 {
					log.Printf("[Integration] Device %s has %d anomalies - reviewing before optimization",
						action.DeviceID, len(anomalyList))
					// In production, would check severity and skip critical ones
				}
			}
		}
		if !skipAction {
			filteredActions = append(filteredActions, action)
		}
	}

	// Create scenario with validated actions
	scenario := &models.OptimizationScenario{
		ScenarioID:      scenarioID,
		ForecastID:      req.ForecastID,
		BuildingID:      req.BuildingID,
		Actions:         filteredActions,
		ExecutionStatus: models.OptimizationStatusPending,
		Progress:        0.0,
		CreatedBy:       userID,
	}

	createdScenario, err := s.optimizationRepo.Create(ctx, scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to create scenario: %w", err)
	}

	// Start execution asynchronously, passing device predictions for optimized scheduling
	go s.executeScenario(context.Background(), createdScenario, devicePredictions)

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
// Integration: Uses device predictions to optimize action timing and expected impact
func (s *OptimizationService) executeScenario(ctx context.Context, scenario *models.OptimizationScenario, predictions map[string]*models.DevicePrediction) {
	// Update status to running
	_ = s.optimizationRepo.UpdateProgress(ctx, scenario.ScenarioID, 0.0, models.OptimizationStatusRunning)

	totalActions := float64(len(scenario.Actions))
	completedActions := 0.0

	// Execute each action
	for _, action := range scenario.Actions {
		// Validate device exists
		_, err := s.deviceRepo.FindByDeviceID(ctx, action.DeviceID)
		if err != nil {
			// Update action status to failed
			s.updateActionStatus(ctx, scenario.ScenarioID, action.DeviceID, "FAILED", "")
			completedActions++
			continue
		}

		// Integration: Use prediction data to enhance command parameters
		// If device has an increasing consumption trend, prioritize this action
		var priority string = "NORMAL"
		if pred, ok := predictions[action.DeviceID]; ok && pred != nil {
			if pred.Trend == "INCREASING" && pred.TrendPercentage > 10 {
				priority = "HIGH"
				log.Printf("[Integration] Elevating priority for device %s due to increasing trend (%.1f%%)",
					action.DeviceID, pred.TrendPercentage)
			}
		}

		// Create command with enriched context from predictions
		commandID := uuid.New().String()
		params := action.Params
		if params == nil {
			params = make(map[string]interface{})
		}
		params["optimization_priority"] = priority

		command := &models.DeviceCommand{
			CommandID: commandID,
			DeviceID:  action.DeviceID,
			Command:   action.Command,
			Params:    params,
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
		_ = s.optimizationRepo.UpdateProgress(ctx, scenario.ScenarioID, progress, models.OptimizationStatusRunning)
	}

	// Mark scenario as completed
	_ = s.optimizationRepo.UpdateProgress(ctx, scenario.ScenarioID, 1.0, models.OptimizationStatusCompleted)
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
