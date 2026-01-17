package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Telemetry represents a telemetry data point
type Telemetry struct {
	ID        primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	DeviceID  string                      `bson:"device_id" json:"deviceId"`
	Timestamp time.Time                   `bson:"timestamp" json:"timestamp"`
	Metrics   map[string]interface{}      `bson:"metrics" json:"metrics"`
	Source    string                      `bson:"source" json:"source"` // "HTTP" or "MQTT"
	CreatedAt time.Time                   `bson:"created_at" json:"createdAt"`
}

// TelemetryResponse represents telemetry data in API responses
type TelemetryResponse struct {
	ID        string                 `json:"id"`
	DeviceID  string                 `json:"deviceId"`
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
	Source    string                 `json:"source"`
}

// ToResponse converts a Telemetry to TelemetryResponse
func (t *Telemetry) ToResponse() *TelemetryResponse {
	return &TelemetryResponse{
		ID:        t.ID.Hex(),
		DeviceID:  t.DeviceID,
		Timestamp: t.Timestamp,
		Metrics:   t.Metrics,
		Source:    t.Source,
	}
}

// TelemetryIngestRequest represents a single telemetry ingestion request
type TelemetryIngestRequest struct {
	DeviceID  string                 `json:"deviceId" binding:"required"`
	Timestamp time.Time               `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics" binding:"required"`
}

// BulkTelemetryIngestRequest represents a batch telemetry ingestion request
type BulkTelemetryIngestRequest struct {
	Telemetry []TelemetryIngestRequest `json:"telemetry" binding:"required"`
}

// TelemetryHistoryRequest represents query parameters for telemetry history
type TelemetryHistoryRequest struct {
	DeviceID string    `form:"deviceId" binding:"required"`
	From     time.Time `form:"from"`
	To       time.Time `form:"to"`
	Page     int       `form:"page"`
	Limit    int       `form:"limit"`
}
