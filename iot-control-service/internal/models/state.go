package models

import (
	"time"
)

// DeviceState represents the current state of a device
type DeviceState struct {
	DeviceID   string                 `json:"deviceId"`
	Status     string                 `json:"status"`
	LastSeen   time.Time              `json:"lastSeen"`
	Metrics    map[string]interface{} `json:"metrics"`
	LastUpdate time.Time              `json:"lastUpdate"`
}

// LiveStateResponse represents live state data for multiple devices
type LiveStateResponse struct {
	Devices []DeviceState `json:"devices"`
	Count   int           `json:"count"`
	Updated time.Time     `json:"updated"`
}
