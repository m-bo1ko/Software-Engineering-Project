package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// KPI represents Key Performance Indicators
type KPI struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	BuildingID   string                 `bson:"building_id,omitempty" json:"buildingId,omitempty"` // Empty for system-wide KPIs
	CalculatedAt time.Time              `bson:"calculated_at" json:"calculatedAt"`
	Metrics      map[string]interface{} `bson:"metrics" json:"metrics"`
	Period       string                 `bson:"period" json:"period"` // "DAILY", "WEEKLY", "MONTHLY"
	CreatedAt    time.Time              `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time              `bson:"updated_at" json:"updatedAt"`
}

// KPIResponse represents KPI data in API responses
type KPIResponse struct {
	ID           string                 `json:"id"`
	BuildingID   string                 `json:"buildingId,omitempty"`
	CalculatedAt time.Time              `json:"calculatedAt"`
	Metrics      map[string]interface{} `json:"metrics"`
	Period       string                 `json:"period"`
	CreatedAt    time.Time              `json:"createdAt"`
}

// ToResponse converts a KPI to KPIResponse
func (k *KPI) ToResponse() *KPIResponse {
	return &KPIResponse{
		ID:           k.ID.Hex(),
		BuildingID:   k.BuildingID,
		CalculatedAt: k.CalculatedAt,
		Metrics:      k.Metrics,
		Period:       k.Period,
		CreatedAt:    k.CreatedAt,
	}
}

// DashboardOverview represents system-wide dashboard metrics
type DashboardOverview struct {
	TotalDevices    int                    `json:"totalDevices"`
	OnlineDevices   int                    `json:"onlineDevices"`
	TotalBuildings  int                    `json:"totalBuildings"`
	ActiveAnomalies int                    `json:"activeAnomalies"`
	KPIs            map[string]interface{} `json:"kpis"`
	RecentAnomalies []AnomalyResponse      `json:"recentAnomalies"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// BuildingDashboard represents building-specific dashboard metrics
type BuildingDashboard struct {
	BuildingID        string                 `json:"buildingId"`
	DeviceCount       int                    `json:"deviceCount"`
	OnlineDeviceCount int                    `json:"onlineDeviceCount"`
	ActiveAnomalies   int                    `json:"activeAnomalies"`
	KPIs              map[string]interface{} `json:"kpis"`
	// Integration: ForecastSummary contains prediction data from Forecast service
	ForecastSummary map[string]interface{} `json:"forecastSummary,omitempty"`
	RecentTelemetry []TimeSeriesResponse   `json:"recentTelemetry"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}
