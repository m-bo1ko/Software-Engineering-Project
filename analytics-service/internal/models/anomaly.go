package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnomalySeverity represents the severity of an anomaly
type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "LOW"
	AnomalySeverityMedium   AnomalySeverity = "MEDIUM"
	AnomalySeverityHigh     AnomalySeverity = "HIGH"
	AnomalySeverityCritical AnomalySeverity = "CRITICAL"
)

// AnomalyStatus represents the status of an anomaly
type AnomalyStatus string

const (
	AnomalyStatusNew         AnomalyStatus = "NEW"
	AnomalyStatusAcknowledged AnomalyStatus = "ACKNOWLEDGED"
	AnomalyStatusResolved    AnomalyStatus = "RESOLVED"
	AnomalyStatusFalsePositive AnomalyStatus = "FALSE_POSITIVE"
)

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID          primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	AnomalyID   string                      `bson:"anomaly_id" json:"anomalyId"`
	DeviceID    string                      `bson:"device_id" json:"deviceId"`
	BuildingID  string                      `bson:"building_id" json:"buildingId"`
	Type        string                      `bson:"type" json:"type"`
	Severity    AnomalySeverity             `bson:"severity" json:"severity"`
	Status      AnomalyStatus               `bson:"status" json:"status"`
	Details     map[string]interface{}      `bson:"details" json:"details"`
	DetectedAt  time.Time                   `bson:"detected_at" json:"detectedAt"`
	AcknowledgedAt *time.Time               `bson:"acknowledged_at,omitempty" json:"acknowledgedAt,omitempty"`
	AcknowledgedBy string                    `bson:"acknowledged_by,omitempty" json:"acknowledgedBy,omitempty"`
	ResolvedAt  *time.Time                  `bson:"resolved_at,omitempty" json:"resolvedAt,omitempty"`
	CreatedAt   time.Time                   `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time                   `bson:"updated_at" json:"updatedAt"`
}

// AnomalyResponse represents anomaly data in API responses
type AnomalyResponse struct {
	ID            string                 `json:"id"`
	AnomalyID     string                 `json:"anomalyId"`
	DeviceID      string                 `json:"deviceId"`
	BuildingID    string                 `json:"buildingId"`
	Type          string                 `json:"type"`
	Severity      string                 `json:"severity"`
	Status        string                 `json:"status"`
	Details       map[string]interface{} `json:"details"`
	DetectedAt    time.Time              `json:"detectedAt"`
	AcknowledgedAt *time.Time            `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy string                `json:"acknowledgedBy,omitempty"`
	ResolvedAt    *time.Time             `json:"resolvedAt,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
}

// ToResponse converts an Anomaly to AnomalyResponse
func (a *Anomaly) ToResponse() *AnomalyResponse {
	return &AnomalyResponse{
		ID:            a.ID.Hex(),
		AnomalyID:     a.AnomalyID,
		DeviceID:      a.DeviceID,
		BuildingID:    a.BuildingID,
		Type:          a.Type,
		Severity:      string(a.Severity),
		Status:        string(a.Status),
		Details:       a.Details,
		DetectedAt:    a.DetectedAt,
		AcknowledgedAt: a.AcknowledgedAt,
		AcknowledgedBy: a.AcknowledgedBy,
		ResolvedAt:    a.ResolvedAt,
		CreatedAt:     a.CreatedAt,
	}
}

// ListAnomaliesRequest represents query parameters for listing anomalies
type ListAnomaliesRequest struct {
	DeviceID   string `form:"deviceId"`
	BuildingID string `form:"buildingId"`
	Type       string `form:"type"`
	Severity   string `form:"severity"`
	Status     string `form:"status"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}

// AcknowledgeAnomalyRequest represents a request to acknowledge an anomaly
type AcknowledgeAnomalyRequest struct {
	AnomalyID string `json:"anomalyId" binding:"required"`
}
