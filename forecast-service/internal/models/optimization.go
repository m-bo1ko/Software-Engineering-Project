package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OptimizationStatus represents the status of an optimization scenario
type OptimizationStatus string

const (
	OptimizationStatusDraft     OptimizationStatus = "DRAFT"
	OptimizationStatusPending   OptimizationStatus = "PENDING"
	OptimizationStatusApproved  OptimizationStatus = "APPROVED"
	OptimizationStatusExecuting OptimizationStatus = "EXECUTING"
	OptimizationStatusCompleted OptimizationStatus = "COMPLETED"
	OptimizationStatusFailed    OptimizationStatus = "FAILED"
	OptimizationStatusCancelled OptimizationStatus = "CANCELLED"
)

// OptimizationType represents the type of optimization
type OptimizationType string

const (
	OptimizationTypeCostReduction     OptimizationType = "COST_REDUCTION"
	OptimizationTypePeakShaving       OptimizationType = "PEAK_SHAVING"
	OptimizationTypeLoadBalancing     OptimizationType = "LOAD_BALANCING"
	OptimizationTypeEfficiency        OptimizationType = "EFFICIENCY"
	OptimizationTypeComfort           OptimizationType = "COMFORT"
	OptimizationTypeDemandResponse    OptimizationType = "DEMAND_RESPONSE"
)

// OptimizationScenario represents an optimization scenario
type OptimizationScenario struct {
	ID                primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	BuildingID        string                  `bson:"building_id" json:"buildingId"`
	Name              string                  `bson:"name" json:"name"`
	Description       string                  `bson:"description" json:"description"`
	Type              OptimizationType        `bson:"type" json:"type"`
	Status            OptimizationStatus      `bson:"status" json:"status"`
	ForecastID        string                  `bson:"forecast_id,omitempty" json:"forecastId,omitempty"`
	ScheduledStart    time.Time               `bson:"scheduled_start" json:"scheduledStart"`
	ScheduledEnd      time.Time               `bson:"scheduled_end" json:"scheduledEnd"`
	Actions           []OptimizationAction    `bson:"actions" json:"actions"`
	ExpectedSavings   Savings                 `bson:"expected_savings" json:"expectedSavings"`
	ActualSavings     *Savings                `bson:"actual_savings,omitempty" json:"actualSavings,omitempty"`
	Constraints       OptimizationConstraints `bson:"constraints" json:"constraints"`
	Priority          int                     `bson:"priority" json:"priority"` // 1-10, higher = more important
	TariffData        *Tariff                 `bson:"tariff_data,omitempty" json:"tariffData,omitempty"`
	WeatherData       *Weather                `bson:"weather_data,omitempty" json:"weatherData,omitempty"`
	CreatedAt         time.Time               `bson:"created_at" json:"createdAt"`
	UpdatedAt         time.Time               `bson:"updated_at" json:"updatedAt"`
	CreatedBy         string                  `bson:"created_by" json:"createdBy"`
	ApprovedBy        string                  `bson:"approved_by,omitempty" json:"approvedBy,omitempty"`
	ApprovedAt        *time.Time              `bson:"approved_at,omitempty" json:"approvedAt,omitempty"`
	ExecutionLog      []ExecutionLogEntry     `bson:"execution_log,omitempty" json:"executionLog,omitempty"`
	ErrorMessage      string                  `bson:"error_message,omitempty" json:"errorMessage,omitempty"`
}

// OptimizationAction represents a single action in an optimization scenario
type OptimizationAction struct {
	ID              string    `bson:"id" json:"id"`
	DeviceID        string    `bson:"device_id" json:"deviceId"`
	DeviceName      string    `bson:"device_name" json:"deviceName"`
	DeviceType      string    `bson:"device_type" json:"deviceType"`
	ActionType      string    `bson:"action_type" json:"actionType"` // SET_TEMP, TURN_OFF, REDUCE_POWER, etc.
	CurrentValue    string    `bson:"current_value" json:"currentValue"`
	TargetValue     string    `bson:"target_value" json:"targetValue"`
	ScheduledTime   time.Time `bson:"scheduled_time" json:"scheduledTime"`
	Duration        int       `bson:"duration" json:"duration"` // in minutes
	Status          string    `bson:"status" json:"status"`     // PENDING, EXECUTED, FAILED
	ExpectedImpact  float64   `bson:"expected_impact" json:"expectedImpact"` // kWh saved
	ActualImpact    *float64  `bson:"actual_impact,omitempty" json:"actualImpact,omitempty"`
	ExecutedAt      *time.Time `bson:"executed_at,omitempty" json:"executedAt,omitempty"`
	ErrorMessage    string    `bson:"error_message,omitempty" json:"errorMessage,omitempty"`
}

// Savings represents energy and cost savings
type Savings struct {
	EnergyKWh       float64 `bson:"energy_kwh" json:"energyKWh"`
	CostAmount      float64 `bson:"cost_amount" json:"costAmount"`
	Currency        string  `bson:"currency" json:"currency"`
	CO2ReductionKg  float64 `bson:"co2_reduction_kg" json:"co2ReductionKg"`
	PercentReduction float64 `bson:"percent_reduction" json:"percentReduction"`
}

// OptimizationConstraints represents constraints for optimization
type OptimizationConstraints struct {
	MinTemperature      *float64 `bson:"min_temperature,omitempty" json:"minTemperature,omitempty"`
	MaxTemperature      *float64 `bson:"max_temperature,omitempty" json:"maxTemperature,omitempty"`
	MinLightLevel       *float64 `bson:"min_light_level,omitempty" json:"minLightLevel,omitempty"`
	OccupancyRequired   bool     `bson:"occupancy_required" json:"occupancyRequired"`
	MaxPeakReduction    *float64 `bson:"max_peak_reduction,omitempty" json:"maxPeakReduction,omitempty"`
	PreserveComfort     bool     `bson:"preserve_comfort" json:"preserveComfort"`
	ExcludeDevices      []string `bson:"exclude_devices,omitempty" json:"excludeDevices,omitempty"`
	TimeWindows         []TimeWindow `bson:"time_windows,omitempty" json:"timeWindows,omitempty"`
}

// TimeWindow represents a time window for optimization
type TimeWindow struct {
	StartTime string `bson:"start_time" json:"startTime"` // HH:MM format
	EndTime   string `bson:"end_time" json:"endTime"`
	DaysOfWeek []string `bson:"days_of_week" json:"daysOfWeek"` // Monday, Tuesday, etc.
}

// ExecutionLogEntry represents a log entry for scenario execution
type ExecutionLogEntry struct {
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Level     string    `bson:"level" json:"level"` // INFO, WARNING, ERROR
	Message   string    `bson:"message" json:"message"`
	ActionID  string    `bson:"action_id,omitempty" json:"actionId,omitempty"`
}

// OptimizationGenerateRequest represents the request to generate an optimization scenario
type OptimizationGenerateRequest struct {
	BuildingID      string                  `json:"buildingId" binding:"required"`
	Name            string                  `json:"name"`
	Type            OptimizationType        `json:"type" binding:"required"`
	ScheduledStart  time.Time               `json:"scheduledStart"`
	ScheduledEnd    time.Time               `json:"scheduledEnd"`
	ForecastID      string                  `json:"forecastId"`
	UseTariffData   bool                    `json:"useTariffData"`
	UseWeatherData  bool                    `json:"useWeatherData"`
	Constraints     OptimizationConstraints `json:"constraints"`
	Priority        int                     `json:"priority"`
}

// OptimizationScenarioResponse represents the optimization scenario in API responses
type OptimizationScenarioResponse struct {
	ID              string                  `json:"id"`
	BuildingID      string                  `json:"buildingId"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Type            OptimizationType        `json:"type"`
	Status          OptimizationStatus      `json:"status"`
	ForecastID      string                  `json:"forecastId,omitempty"`
	ScheduledStart  time.Time               `json:"scheduledStart"`
	ScheduledEnd    time.Time               `json:"scheduledEnd"`
	Actions         []OptimizationAction    `json:"actions"`
	ExpectedSavings Savings                 `json:"expectedSavings"`
	ActualSavings   *Savings                `json:"actualSavings,omitempty"`
	Constraints     OptimizationConstraints `json:"constraints"`
	Priority        int                     `json:"priority"`
	CreatedAt       time.Time               `json:"createdAt"`
	CreatedBy       string                  `json:"createdBy"`
	ApprovedBy      string                  `json:"approvedBy,omitempty"`
	ErrorMessage    string                  `json:"errorMessage,omitempty"`
}

// ToResponse converts an OptimizationScenario to OptimizationScenarioResponse
func (o *OptimizationScenario) ToResponse() *OptimizationScenarioResponse {
	return &OptimizationScenarioResponse{
		ID:              o.ID.Hex(),
		BuildingID:      o.BuildingID,
		Name:            o.Name,
		Description:     o.Description,
		Type:            o.Type,
		Status:          o.Status,
		ForecastID:      o.ForecastID,
		ScheduledStart:  o.ScheduledStart,
		ScheduledEnd:    o.ScheduledEnd,
		Actions:         o.Actions,
		ExpectedSavings: o.ExpectedSavings,
		ActualSavings:   o.ActualSavings,
		Constraints:     o.Constraints,
		Priority:        o.Priority,
		CreatedAt:       o.CreatedAt,
		CreatedBy:       o.CreatedBy,
		ApprovedBy:      o.ApprovedBy,
		ErrorMessage:    o.ErrorMessage,
	}
}

// SendToIoTRequest represents the request to send a scenario to IoT service
type SendToIoTRequest struct {
	ScenarioID  string `json:"scenarioId" binding:"required"`
	ExecuteNow  bool   `json:"executeNow"`
	DryRun      bool   `json:"dryRun"`
}

// SendToIoTResponse represents the response from sending to IoT service
type SendToIoTResponse struct {
	Success       bool     `json:"success"`
	ScenarioID    string   `json:"scenarioId"`
	ActionsQueued int      `json:"actionsQueued"`
	ActionsSkipped int     `json:"actionsSkipped"`
	Errors        []string `json:"errors,omitempty"`
	ExecutionID   string   `json:"executionId,omitempty"`
}
