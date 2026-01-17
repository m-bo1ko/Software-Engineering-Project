package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Device represents a device for tracking
type Device struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeviceID          string             `bson:"device_id" json:"deviceId"`
	BuildingID        string             `bson:"building_id" json:"buildingId"`
	Name              string             `bson:"name" json:"name"`
	Type              string             `bson:"type" json:"type"` // HVAC, LIGHTING, EQUIPMENT, SENSOR
	Location          string             `bson:"location" json:"location"`
	PowerRating       float64            `bson:"power_rating" json:"powerRating"` // in kW
	OperatingSchedule *OperatingSchedule `bson:"operating_schedule,omitempty" json:"operatingSchedule,omitempty"`
	Metadata          map[string]string  `bson:"metadata,omitempty" json:"metadata,omitempty"`
	LastSyncAt        time.Time          `bson:"last_sync_at" json:"lastSyncAt"`
	CreatedAt         time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updatedAt"`
}

// OperatingSchedule represents the operating schedule for a device
type OperatingSchedule struct {
	TimeZone   string           `bson:"timezone" json:"timezone"`
	WeeklySchedule []DaySchedule `bson:"weekly_schedule" json:"weeklySchedule"`
}

// DaySchedule represents the schedule for a day
type DaySchedule struct {
	DayOfWeek string   `bson:"day_of_week" json:"dayOfWeek"`
	StartTime string   `bson:"start_time" json:"startTime"` // HH:MM
	EndTime   string   `bson:"end_time" json:"endTime"`
	IsActive  bool     `bson:"is_active" json:"isActive"`
}

// DeviceState represents the current state of a device from IoT service
type DeviceState struct {
	DeviceID       string                 `json:"deviceId"`
	Status         string                 `json:"status"` // ONLINE, OFFLINE, ERROR
	CurrentPower   float64                `json:"currentPower"` // in kW
	CurrentState   string                 `json:"currentState"` // ON, OFF, STANDBY
	LastReading    time.Time              `json:"lastReading"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Controllable   bool                   `json:"controllable"`
}

// HistoricalConsumption represents historical consumption data
type HistoricalConsumption struct {
	BuildingID  string                    `json:"buildingId"`
	DeviceID    string                    `json:"deviceId,omitempty"`
	Period      AnalysisPeriod            `json:"period"`
	Resolution  string                    `json:"resolution"` // HOURLY, DAILY, WEEKLY
	DataPoints  []ConsumptionDataPoint    `json:"dataPoints"`
	Summary     ConsumptionSummary        `json:"summary"`
}

// ConsumptionDataPoint represents a single consumption data point
type ConsumptionDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"` // kWh
	Quality     string    `json:"quality"` // ACTUAL, ESTIMATED, INTERPOLATED
}

// ConsumptionSummary provides summary statistics
type ConsumptionSummary struct {
	TotalKWh    float64 `json:"totalKWh"`
	AverageKW   float64 `json:"averageKW"`
	PeakKW      float64 `json:"peakKW"`
	MinKW       float64 `json:"minKW"`
	DataPoints  int     `json:"dataPoints"`
}
