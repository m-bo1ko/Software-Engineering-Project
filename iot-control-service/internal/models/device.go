package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceStatus represents the status of a device
type DeviceStatus string

const (
	DeviceStatusOnline      DeviceStatus = "ONLINE"
	DeviceStatusOffline     DeviceStatus = "OFFLINE"
	DeviceStatusError       DeviceStatus = "ERROR"
	DeviceStatusMaintenance DeviceStatus = "MAINTENANCE"
)

// Device represents a device in the system
type Device struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	DeviceID     string                 `bson:"device_id" json:"deviceId"`
	Type         string                 `bson:"type" json:"type"`
	Model        string                 `bson:"model" json:"model"`
	Location     DeviceLocation         `bson:"location" json:"location"`
	Capabilities []string               `bson:"capabilities" json:"capabilities"`
	Status       DeviceStatus           `bson:"status" json:"status"`
	LastSeen     time.Time              `bson:"last_seen" json:"lastSeen"`
	Metadata     map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    time.Time              `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time              `bson:"updated_at" json:"updatedAt"`
	CreatedBy    string                 `bson:"created_by" json:"createdBy"`
}

// DeviceLocation represents device location information
type DeviceLocation struct {
	BuildingID string  `bson:"building_id" json:"buildingId"`
	Floor      string  `bson:"floor,omitempty" json:"floor,omitempty"`
	Room       string  `bson:"room,omitempty" json:"room,omitempty"`
	Latitude   float64 `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Longitude  float64 `bson:"longitude,omitempty" json:"longitude,omitempty"`
}

// UnmarshalJSON allows DeviceLocation to be unmarshaled from either a string or an object
func (dl *DeviceLocation) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var locationString string
	if err := json.Unmarshal(data, &locationString); err == nil {
		// If it's a string, parse it as room description
		dl.Room = locationString
		return nil
	}

	// Otherwise, unmarshal as object
	type Alias DeviceLocation
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(dl),
	}
	return json.Unmarshal(data, aux)
}

// DeviceResponse represents device data in API responses
type DeviceResponse struct {
	ID           string                 `json:"id"`
	DeviceID     string                 `json:"deviceId"`
	Type         string                 `json:"type"`
	Model        string                 `json:"model"`
	Location     DeviceLocation         `json:"location"`
	Capabilities []string               `json:"capabilities"`
	Status       string                 `json:"status"`
	LastSeen     time.Time              `json:"lastSeen"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// ToResponse converts a Device to DeviceResponse
func (d *Device) ToResponse() *DeviceResponse {
	return &DeviceResponse{
		ID:           d.ID.Hex(),
		DeviceID:     d.DeviceID,
		Type:         d.Type,
		Model:        d.Model,
		Location:     d.Location,
		Capabilities: d.Capabilities,
		Status:       string(d.Status),
		LastSeen:     d.LastSeen,
		Metadata:     d.Metadata,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

// RegisterDeviceRequest represents a request to register a device
type RegisterDeviceRequest struct {
	DeviceID     string                 `json:"deviceId" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Model        string                 `json:"model"`
	Name         string                 `json:"name"`
	BuildingID   string                 `json:"buildingId"`
	Location     DeviceLocation         `json:"location"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GetBuildingID returns the building ID from either top-level field or location
func (r *RegisterDeviceRequest) GetBuildingID() string {
	if r.BuildingID != "" {
		return r.BuildingID
	}
	return r.Location.BuildingID
}

// ListDevicesRequest represents query parameters for listing devices
type ListDevicesRequest struct {
	BuildingID string `form:"buildingId"`
	Type       string `form:"type"`
	Status     string `form:"status"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}

// DevicePrediction represents forecast prediction data for a device
type DevicePrediction struct {
	DeviceID           string               `json:"deviceId"`
	DeviceName         string               `json:"deviceName"`
	DeviceType         string               `json:"deviceType"`
	CurrentConsumption float64              `json:"currentConsumption"`
	PredictedValues    []ForecastPrediction `json:"predictedValues"`
	Trend              string               `json:"trend"`
	TrendPercentage    float64              `json:"trendPercentage"`
}

// ForecastPrediction represents a single prediction point
type ForecastPrediction struct {
	Timestamp       time.Time `json:"timestamp"`
	PredictedValue  float64   `json:"predictedValue"`
	LowerBound      float64   `json:"lowerBound"`
	UpperBound      float64   `json:"upperBound"`
	ConfidenceLevel float64   `json:"confidenceLevel"`
	Unit            string    `json:"unit"`
}
