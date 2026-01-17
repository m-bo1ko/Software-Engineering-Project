package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OptimizationExecutionStatus represents the execution status of an optimization scenario
type OptimizationExecutionStatus string

const (
	OptimizationStatusPending   OptimizationExecutionStatus = "PENDING"
	OptimizationStatusRunning   OptimizationExecutionStatus = "RUNNING"
	OptimizationStatusCompleted OptimizationExecutionStatus = "COMPLETED"
	OptimizationStatusFailed    OptimizationExecutionStatus = "FAILED"
	OptimizationStatusCancelled OptimizationExecutionStatus = "CANCELLED"
)

// OptimizationScenario represents an optimization scenario
type OptimizationScenario struct {
	ID              primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	ScenarioID      string                      `bson:"scenario_id" json:"scenarioId"`
	ForecastID      string                      `bson:"forecast_id,omitempty" json:"forecastId,omitempty"`
	BuildingID      string                      `bson:"building_id" json:"buildingId"`
	Actions         []OptimizationAction        `bson:"actions" json:"actions"`
	ExecutionStatus OptimizationExecutionStatus `bson:"execution_status" json:"executionStatus"`
	Progress        float64                     `bson:"progress" json:"progress"` // 0.0 to 1.0
	StartedAt       *time.Time                  `bson:"started_at,omitempty" json:"startedAt,omitempty"`
	CompletedAt     *time.Time                  `bson:"completed_at,omitempty" json:"completedAt,omitempty"`
	ErrorMsg        string                      `bson:"error_msg,omitempty" json:"errorMsg,omitempty"`
	CreatedBy       string                      `bson:"created_by" json:"createdBy"`
	CreatedAt       time.Time                   `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time                   `bson:"updated_at" json:"updatedAt"`
}

// OptimizationAction represents a single action in an optimization scenario
type OptimizationAction struct {
	DeviceID  string                 `bson:"device_id" json:"deviceId"`
	Command   string                 `bson:"command" json:"command"`
	Params    map[string]interface{} `bson:"params" json:"params"`
	Priority  int                    `bson:"priority" json:"priority"`
	Status    string                 `bson:"status" json:"status"` // "PENDING", "SENT", "APPLIED", "FAILED"
	CommandID string                 `bson:"command_id,omitempty" json:"commandId,omitempty"`
}

// OptimizationScenarioResponse represents optimization scenario data in API responses
type OptimizationScenarioResponse struct {
	ID              string                 `json:"id"`
	ScenarioID      string                 `json:"scenarioId"`
	ForecastID      string                 `json:"forecastId,omitempty"`
	BuildingID      string                 `json:"buildingId"`
	Actions         []OptimizationAction   `json:"actions"`
	ExecutionStatus string                 `json:"executionStatus"`
	Progress        float64                `json:"progress"`
	StartedAt       *time.Time             `json:"startedAt,omitempty"`
	CompletedAt     *time.Time             `json:"completedAt,omitempty"`
	ErrorMsg        string                 `json:"errorMsg,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// ToResponse converts an OptimizationScenario to OptimizationScenarioResponse
func (o *OptimizationScenario) ToResponse() *OptimizationScenarioResponse {
	return &OptimizationScenarioResponse{
		ID:              o.ID.Hex(),
		ScenarioID:      o.ScenarioID,
		ForecastID:      o.ForecastID,
		BuildingID:      o.BuildingID,
		Actions:         o.Actions,
		ExecutionStatus: string(o.ExecutionStatus),
		Progress:        o.Progress,
		StartedAt:       o.StartedAt,
		CompletedAt:     o.CompletedAt,
		ErrorMsg:        o.ErrorMsg,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

// ApplyOptimizationRequest represents a request to apply an optimization scenario
type ApplyOptimizationRequest struct {
	ScenarioID string                 `json:"scenarioId" binding:"required"`
	ForecastID string                 `json:"forecastId,omitempty"`
	BuildingID string                 `json:"buildingId" binding:"required"`
	Actions    []OptimizationAction    `json:"actions" binding:"required"`
}
