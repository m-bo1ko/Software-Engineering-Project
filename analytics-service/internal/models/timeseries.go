package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AggregationType represents the type of time-series aggregation
type AggregationType string

const (
	AggregationTypeHourly AggregationType = "HOURLY"
	AggregationTypeDaily  AggregationType = "DAILY"
	AggregationTypeWeekly AggregationType = "WEEKLY"
	AggregationTypeMonthly AggregationType = "MONTHLY"
)

// TimeSeries represents aggregated time-series data
type TimeSeries struct {
	ID              primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	DeviceID        string                      `bson:"device_id" json:"deviceId"`
	BuildingID      string                      `bson:"building_id" json:"buildingId"`
	Timestamp       time.Time                   `bson:"timestamp" json:"timestamp"`
	AggregationType AggregationType             `bson:"aggregation_type" json:"aggregationType"`
	Metrics         map[string]interface{}      `bson:"metrics" json:"metrics"`
	CreatedAt       time.Time                   `bson:"created_at" json:"createdAt"`
}

// TimeSeriesResponse represents time-series data in API responses
type TimeSeriesResponse struct {
	DeviceID        string                 `json:"deviceId"`
	BuildingID      string                 `json:"buildingId"`
	Timestamp       time.Time              `json:"timestamp"`
	AggregationType string                 `json:"aggregationType"`
	Metrics         map[string]interface{} `json:"metrics"`
}

// ToResponse converts a TimeSeries to TimeSeriesResponse
func (t *TimeSeries) ToResponse() *TimeSeriesResponse {
	return &TimeSeriesResponse{
		DeviceID:        t.DeviceID,
		BuildingID:      t.BuildingID,
		Timestamp:       t.Timestamp,
		AggregationType: string(t.AggregationType),
		Metrics:         t.Metrics,
	}
}

// TimeSeriesQueryRequest represents a request to query time-series data
type TimeSeriesQueryRequest struct {
	DeviceIDs       []string    `json:"deviceIds,omitempty"`
	BuildingID      string      `json:"buildingId,omitempty"`
	From            time.Time   `json:"from" binding:"required"`
	To              time.Time   `json:"to" binding:"required"`
	AggregationType string      `json:"aggregationType" binding:"required,oneof=HOURLY DAILY WEEKLY MONTHLY"`
	Metrics         []string    `json:"metrics,omitempty"`
}
