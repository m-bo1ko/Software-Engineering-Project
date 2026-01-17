package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"forecast-service/internal/integrations"
	"forecast-service/internal/models"
	"forecast-service/internal/repository"
)

// OptimizationService handles optimization scenario business logic
type OptimizationService struct {
	optimizationRepo   *repository.OptimizationRepository
	forecastRepo       *repository.ForecastRepository
	recommendationRepo *repository.RecommendationRepository
	iotClient          *integrations.IoTClient
	externalClient     *integrations.ExternalClient
	securityClient     *integrations.SecurityClient
}

// NewOptimizationService creates a new optimization service
func NewOptimizationService(
	optimizationRepo *repository.OptimizationRepository,
	forecastRepo *repository.ForecastRepository,
	recommendationRepo *repository.RecommendationRepository,
	iotClient *integrations.IoTClient,
	externalClient *integrations.ExternalClient,
	securityClient *integrations.SecurityClient,
) *OptimizationService {
	return &OptimizationService{
		optimizationRepo:   optimizationRepo,
		forecastRepo:       forecastRepo,
		recommendationRepo: recommendationRepo,
		iotClient:          iotClient,
		externalClient:     externalClient,
		securityClient:     securityClient,
	}
}

// GenerateOptimization generates an optimization scenario
func (s *OptimizationService) GenerateOptimization(ctx context.Context, req *models.OptimizationGenerateRequest, userID, authToken string) (*models.OptimizationScenarioResponse, error) {
	// Set defaults
	if req.ScheduledStart.IsZero() {
		req.ScheduledStart = time.Now().Add(time.Hour)
	}
	if req.ScheduledEnd.IsZero() {
		req.ScheduledEnd = req.ScheduledStart.Add(8 * time.Hour)
	}
	if req.Priority <= 0 {
		req.Priority = 5
	}

	// Generate scenario name if not provided
	name := req.Name
	if name == "" {
		name = fmt.Sprintf("%s Optimization - %s", req.Type, time.Now().Format("2006-01-02 15:04"))
	}

	// Fetch forecast if provided
	var forecast *models.Forecast
	if req.ForecastID != "" {
		var err error
		forecast, err = s.forecastRepo.FindByID(ctx, req.ForecastID)
		if err != nil {
			return nil, fmt.Errorf("failed to get forecast: %w", err)
		}
	}

	// Get device states
	devices, err := s.iotClient.GetDevicesByBuilding(ctx, req.BuildingID, authToken)
	if err != nil {
		// Continue without device states, use simulated data
		devices = s.generateSimulatedDevices(req.BuildingID)
	}

	// Fetch tariff data if requested
	var tariffData *models.Tariff
	if req.UseTariffData {
		tariffData, _ = s.externalClient.GetCurrentTariff(ctx, "default", authToken)
	}

	// Fetch weather data if requested
	var weatherData *models.Weather
	if req.UseWeatherData {
		weatherData, _ = s.externalClient.GetCurrentWeather(ctx, req.BuildingID, authToken)
	}

	// Generate optimization actions based on type
	actions := s.generateOptimizationActions(req.Type, devices, forecast, tariffData, req.Constraints, req.ScheduledStart)

	// Calculate expected savings
	expectedSavings := s.calculateExpectedSavings(actions, tariffData)

	// Generate description
	description := s.generateScenarioDescription(req.Type, actions, expectedSavings)

	scenario := &models.OptimizationScenario{
		BuildingID:      req.BuildingID,
		Name:            name,
		Description:     description,
		Type:            req.Type,
		Status:          models.OptimizationStatusDraft,
		ForecastID:      req.ForecastID,
		ScheduledStart:  req.ScheduledStart,
		ScheduledEnd:    req.ScheduledEnd,
		Actions:         actions,
		ExpectedSavings: expectedSavings,
		Constraints:     req.Constraints,
		Priority:        req.Priority,
		TariffData:      tariffData,
		WeatherData:     weatherData,
		CreatedBy:       userID,
	}

	createdScenario, err := s.optimizationRepo.Create(ctx, scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to create scenario: %w", err)
	}

	return createdScenario.ToResponse(), nil
}

// generateSimulatedDevices creates simulated device states for demo
func (s *OptimizationService) generateSimulatedDevices(buildingID string) []models.DeviceState {
	return []models.DeviceState{
		{DeviceID: "hvac-1", Status: "ONLINE", CurrentPower: 25.5, CurrentState: "ON", Controllable: true},
		{DeviceID: "hvac-2", Status: "ONLINE", CurrentPower: 22.0, CurrentState: "ON", Controllable: true},
		{DeviceID: "lighting-1", Status: "ONLINE", CurrentPower: 5.2, CurrentState: "ON", Controllable: true},
		{DeviceID: "lighting-2", Status: "ONLINE", CurrentPower: 4.8, CurrentState: "ON", Controllable: true},
		{DeviceID: "equipment-1", Status: "ONLINE", CurrentPower: 15.0, CurrentState: "ON", Controllable: false},
	}
}

// generateOptimizationActions generates optimization actions based on type
func (s *OptimizationService) generateOptimizationActions(
	optType models.OptimizationType,
	devices []models.DeviceState,
	forecast *models.Forecast,
	tariff *models.Tariff,
	constraints models.OptimizationConstraints,
	startTime time.Time,
) []models.OptimizationAction {
	var actions []models.OptimizationAction

	for _, device := range devices {
		if !device.Controllable {
			continue
		}

		// Check if device is excluded
		excluded := false
		for _, excludeID := range constraints.ExcludeDevices {
			if excludeID == device.DeviceID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		action := s.createActionForDevice(optType, device, tariff, constraints, startTime)
		if action != nil {
			actions = append(actions, *action)
		}
	}

	return actions
}

// createActionForDevice creates an optimization action for a specific device
func (s *OptimizationService) createActionForDevice(
	optType models.OptimizationType,
	device models.DeviceState,
	tariff *models.Tariff,
	constraints models.OptimizationConstraints,
	startTime time.Time,
) *models.OptimizationAction {
	actionID := uuid.New().String()[:8]

	switch optType {
	case models.OptimizationTypeCostReduction:
		if device.CurrentPower > 10 {
			reduction := device.CurrentPower * 0.15
			return &models.OptimizationAction{
				ID:             actionID,
				DeviceID:       device.DeviceID,
				DeviceName:     "Device " + device.DeviceID,
				DeviceType:     s.inferDeviceType(device.DeviceID),
				ActionType:     "REDUCE_POWER",
				CurrentValue:   fmt.Sprintf("%.1f kW", device.CurrentPower),
				TargetValue:    fmt.Sprintf("%.1f kW", device.CurrentPower-reduction),
				ScheduledTime:  startTime,
				Duration:       60,
				Status:         "PENDING",
				ExpectedImpact: reduction,
			}
		}

	case models.OptimizationTypePeakShaving:
		if device.CurrentPower > 15 {
			return &models.OptimizationAction{
				ID:             actionID,
				DeviceID:       device.DeviceID,
				DeviceName:     "Device " + device.DeviceID,
				DeviceType:     s.inferDeviceType(device.DeviceID),
				ActionType:     "REDUCE_POWER",
				CurrentValue:   fmt.Sprintf("%.1f kW", device.CurrentPower),
				TargetValue:    fmt.Sprintf("%.1f kW", device.CurrentPower*0.7),
				ScheduledTime:  startTime,
				Duration:       30,
				Status:         "PENDING",
				ExpectedImpact: device.CurrentPower * 0.3,
			}
		}

	case models.OptimizationTypeEfficiency:
		if s.isHVACDevice(device.DeviceID) && !constraints.PreserveComfort {
			return &models.OptimizationAction{
				ID:             actionID,
				DeviceID:       device.DeviceID,
				DeviceName:     "Device " + device.DeviceID,
				DeviceType:     "HVAC",
				ActionType:     "SET_TEMP",
				CurrentValue:   "22°C",
				TargetValue:    "24°C",
				ScheduledTime:  startTime,
				Duration:       120,
				Status:         "PENDING",
				ExpectedImpact: device.CurrentPower * 0.2,
			}
		}

	case models.OptimizationTypeDemandResponse:
		return &models.OptimizationAction{
			ID:             actionID,
			DeviceID:       device.DeviceID,
			DeviceName:     "Device " + device.DeviceID,
			DeviceType:     s.inferDeviceType(device.DeviceID),
			ActionType:     "CURTAIL",
			CurrentValue:   fmt.Sprintf("%.1f kW", device.CurrentPower),
			TargetValue:    fmt.Sprintf("%.1f kW", device.CurrentPower*0.5),
			ScheduledTime:  startTime,
			Duration:       60,
			Status:         "PENDING",
			ExpectedImpact: device.CurrentPower * 0.5,
		}
	}

	return nil
}

// inferDeviceType infers device type from ID
func (s *OptimizationService) inferDeviceType(deviceID string) string {
	if s.isHVACDevice(deviceID) {
		return "HVAC"
	}
	if s.isLightingDevice(deviceID) {
		return "LIGHTING"
	}
	return "EQUIPMENT"
}

func (s *OptimizationService) isHVACDevice(deviceID string) bool {
	return len(deviceID) > 4 && deviceID[:4] == "hvac"
}

func (s *OptimizationService) isLightingDevice(deviceID string) bool {
	return len(deviceID) > 8 && deviceID[:8] == "lighting"
}

// calculateExpectedSavings calculates expected savings from actions
func (s *OptimizationService) calculateExpectedSavings(actions []models.OptimizationAction, tariff *models.Tariff) models.Savings {
	var totalEnergyKWh float64
	for _, action := range actions {
		energySaved := action.ExpectedImpact * (float64(action.Duration) / 60)
		totalEnergyKWh += energySaved
	}

	rate := 0.15 // Default rate
	currency := "USD"
	if tariff != nil {
		rate = tariff.CurrentRate
		currency = tariff.Currency
	}

	costSaved := totalEnergyKWh * rate
	co2Reduction := totalEnergyKWh * 0.4 // Approximate kg CO2 per kWh

	return models.Savings{
		EnergyKWh:        math.Round(totalEnergyKWh*100) / 100,
		CostAmount:       math.Round(costSaved*100) / 100,
		Currency:         currency,
		CO2ReductionKg:   math.Round(co2Reduction*100) / 100,
		PercentReduction: 12.5, // Estimated
	}
}

// generateScenarioDescription generates a description for the scenario
func (s *OptimizationService) generateScenarioDescription(optType models.OptimizationType, actions []models.OptimizationAction, savings models.Savings) string {
	return fmt.Sprintf(
		"%s optimization scenario with %d actions. Expected savings: %.1f kWh (%.2f %s), CO2 reduction: %.1f kg",
		optType,
		len(actions),
		savings.EnergyKWh,
		savings.CostAmount,
		savings.Currency,
		savings.CO2ReductionKg,
	)
}

// GetScenario retrieves an optimization scenario by ID
func (s *OptimizationService) GetScenario(ctx context.Context, scenarioID string) (*models.OptimizationScenarioResponse, error) {
	scenario, err := s.optimizationRepo.FindByID(ctx, scenarioID)
	if err != nil {
		return nil, err
	}
	return scenario.ToResponse(), nil
}

// GetRecommendations retrieves energy-saving recommendations for a building
func (s *OptimizationService) GetRecommendations(ctx context.Context, buildingID, authToken string) (*models.RecommendationsResponse, error) {
	// Try to get existing recommendations
	recs, err := s.recommendationRepo.FindByBuilding(ctx, buildingID)
	if err != nil {
		return nil, err
	}

	// If no recommendations, generate new ones
	if len(recs) == 0 {
		recs = s.generateRecommendations(ctx, buildingID, authToken)
		if len(recs) > 0 {
			s.recommendationRepo.CreateMany(ctx, recs)
		}
	}

	// Build response
	response := &models.RecommendationsResponse{
		BuildingID:            buildingID,
		TotalRecommendations:  len(recs),
		TotalPotentialSavings: models.Savings{Currency: "USD"},
		ByPriority:            models.PrioritySummary{},
		ByCategory:            make(map[string]int),
		Recommendations:       make([]models.RecommendationItem, len(recs)),
		GeneratedAt:           time.Now(),
	}

	for i, rec := range recs {
		response.Recommendations[i] = rec.ToRecommendationItem()
		response.TotalPotentialSavings.EnergyKWh += rec.ExpectedSavings.EnergyKWh
		response.TotalPotentialSavings.CostAmount += rec.ExpectedSavings.CostAmount
		response.TotalPotentialSavings.CO2ReductionKg += rec.ExpectedSavings.CO2ReductionKg

		response.ByCategory[rec.Category]++

		switch rec.Priority {
		case models.RecommendationPriorityCritical:
			response.ByPriority.Critical++
		case models.RecommendationPriorityHigh:
			response.ByPriority.High++
		case models.RecommendationPriorityMedium:
			response.ByPriority.Medium++
		case models.RecommendationPriorityLow:
			response.ByPriority.Low++
		}
	}

	return response, nil
}

// generateRecommendations generates recommendations for a building
func (s *OptimizationService) generateRecommendations(ctx context.Context, buildingID, authToken string) []*models.Recommendation {
	recommendations := []*models.Recommendation{
		{
			BuildingID:  buildingID,
			Type:        models.RecommendationTypeImmediate,
			Priority:    models.RecommendationPriorityHigh,
			Title:       "Optimize HVAC Setpoints",
			Description: "Current HVAC setpoints can be adjusted to reduce energy consumption while maintaining comfort.",
			ActionRequired: "Increase cooling setpoint by 2°C during peak hours",
			ExpectedSavings: models.Savings{
				EnergyKWh:      150.0,
				CostAmount:     22.50,
				Currency:       "USD",
				CO2ReductionKg: 60.0,
			},
			ImplementationSteps: []string{
				"Review current HVAC schedules",
				"Adjust cooling setpoint from 22°C to 24°C during peak hours (14:00-18:00)",
				"Monitor comfort levels and adjust if needed",
			},
			AutomationAvailable: true,
			Category:            "HVAC",
			ValidFrom:           time.Now(),
		},
		{
			BuildingID:  buildingID,
			Type:        models.RecommendationTypeScheduled,
			Priority:    models.RecommendationPriorityMedium,
			Title:       "Implement Lighting Schedules",
			Description: "Lighting in common areas can be scheduled to reduce unnecessary usage.",
			ActionRequired: "Configure automatic lighting schedules",
			ExpectedSavings: models.Savings{
				EnergyKWh:      80.0,
				CostAmount:     12.00,
				Currency:       "USD",
				CO2ReductionKg: 32.0,
			},
			ImplementationSteps: []string{
				"Identify common areas with extended lighting hours",
				"Configure occupancy-based or scheduled lighting",
				"Set dimming levels for daylight harvesting",
			},
			AutomationAvailable: true,
			Category:            "LIGHTING",
			ValidFrom:           time.Now(),
		},
		{
			BuildingID:  buildingID,
			Type:        models.RecommendationTypeLongTerm,
			Priority:    models.RecommendationPriorityLow,
			Title:       "Equipment Upgrade Assessment",
			Description: "Some equipment may benefit from efficiency upgrades.",
			ActionRequired: "Schedule equipment efficiency audit",
			ExpectedSavings: models.Savings{
				EnergyKWh:      500.0,
				CostAmount:     75.00,
				Currency:       "USD",
				CO2ReductionKg: 200.0,
			},
			ImplementationSteps: []string{
				"List all major energy-consuming equipment",
				"Assess age and efficiency ratings",
				"Evaluate upgrade options and ROI",
			},
			AutomationAvailable: false,
			Category:            "EQUIPMENT",
			ValidFrom:           time.Now(),
		},
	}

	return recommendations
}

// SendToIoT sends an optimization scenario to the IoT service for execution
func (s *OptimizationService) SendToIoT(ctx context.Context, req *models.SendToIoTRequest, userID, authToken string) (*models.SendToIoTResponse, error) {
	// Get scenario
	scenario, err := s.optimizationRepo.FindByID(ctx, req.ScenarioID)
	if err != nil {
		return nil, err
	}

	// Check status
	if scenario.Status != models.OptimizationStatusApproved && scenario.Status != models.OptimizationStatusDraft {
		return nil, fmt.Errorf("scenario must be approved or draft to send to IoT")
	}

	// Approve if draft
	if scenario.Status == models.OptimizationStatusDraft {
		if err := s.optimizationRepo.ApproveScenario(ctx, req.ScenarioID, userID); err != nil {
			return nil, fmt.Errorf("failed to approve scenario: %w", err)
		}
	}

	// Send to IoT service
	iotResp, err := s.iotClient.ApplyOptimization(ctx, scenario, req.ExecuteNow, req.DryRun, authToken)
	if err != nil {
		s.optimizationRepo.UpdateStatus(ctx, req.ScenarioID, models.OptimizationStatusFailed, err.Error())
		return nil, fmt.Errorf("failed to send to IoT: %w", err)
	}

	// Update scenario status
	if iotResp.Success && !req.DryRun {
		s.optimizationRepo.UpdateStatus(ctx, req.ScenarioID, models.OptimizationStatusExecuting, "")
		s.optimizationRepo.AddExecutionLog(ctx, req.ScenarioID, models.ExecutionLogEntry{
			Level:   "INFO",
			Message: fmt.Sprintf("Sent to IoT service. Execution ID: %s", iotResp.ExecutionID),
		})
	}

	return &models.SendToIoTResponse{
		Success:       iotResp.Success,
		ScenarioID:    req.ScenarioID,
		ActionsQueued: iotResp.ActionsQueued,
		ActionsSkipped: iotResp.ActionsSkipped,
		Errors:        iotResp.Errors,
		ExecutionID:   iotResp.ExecutionID,
	}, nil
}

// GetDeviceOptimization retrieves optimization recommendations for a device
func (s *OptimizationService) GetDeviceOptimization(ctx context.Context, deviceID, authToken string) (*models.DeviceOptimization, error) {
	// Get device state
	deviceState, err := s.iotClient.GetDeviceState(ctx, deviceID, authToken)
	if err != nil {
		// Use simulated data
		deviceState = &models.DeviceState{
			DeviceID:     deviceID,
			Status:       "ONLINE",
			CurrentPower: 20.0,
			CurrentState: "ON",
			Controllable: true,
		}
	}

	// Generate optimization recommendations
	recommendations := []string{}
	scheduledActions := []models.ScheduledAction{}
	potentialSavings := 0.0

	deviceType := s.inferDeviceType(deviceID)

	switch deviceType {
	case "HVAC":
		recommendations = append(recommendations,
			"Consider increasing setpoint by 1-2°C during peak hours",
			"Enable pre-cooling before peak tariff periods",
			"Check filter status - dirty filters reduce efficiency",
		)
		potentialSavings = deviceState.CurrentPower * 0.15 * 8 // 15% reduction over 8 hours
		scheduledActions = append(scheduledActions, models.ScheduledAction{
			Time:        time.Now().Add(2 * time.Hour),
			Action:      "REDUCE_POWER",
			TargetState: "ECO_MODE",
			Reason:      "Peak tariff period approaching",
		})

	case "LIGHTING":
		recommendations = append(recommendations,
			"Enable daylight harvesting during daytime hours",
			"Reduce illumination levels in unoccupied areas",
			"Schedule automatic off during non-business hours",
		)
		potentialSavings = deviceState.CurrentPower * 0.3 * 10

	default:
		recommendations = append(recommendations,
			"Review operating schedule for optimization opportunities",
			"Consider standby mode during low-usage periods",
		)
		potentialSavings = deviceState.CurrentPower * 0.1 * 8
	}

	return &models.DeviceOptimization{
		DeviceID:         deviceID,
		DeviceName:       "Device " + deviceID,
		CurrentState:     deviceState.CurrentState,
		OptimalState:     "ECO_MODE",
		PotentialSavings: math.Round(potentialSavings*100) / 100,
		SavingsUnit:      "kWh/day",
		Recommendations:  recommendations,
		ScheduledActions: scheduledActions,
		Priority:         "MEDIUM",
	}, nil
}
