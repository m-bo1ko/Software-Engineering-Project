package models

import "time"

// EnergyConsumptionRequest represents the query parameters for energy consumption
type EnergyConsumptionRequest struct {
	BuildingID string    `form:"buildingId" binding:"required"`
	From       time.Time `form:"from" binding:"required"`
	To         time.Time `form:"to" binding:"required"`
}

// EnergyConsumption represents energy consumption data from external provider
type EnergyConsumption struct {
	BuildingID   string                      `json:"buildingId"`
	Period       EnergyPeriod                `json:"period"`
	TotalKWh     float64                     `json:"totalKWh"`
	PeakKW       float64                     `json:"peakKW"`
	AverageKW    float64                     `json:"averageKW"`
	CostEstimate float64                     `json:"costEstimate"`
	Currency     string                      `json:"currency"`
	Breakdown    []EnergyConsumptionBreakdown `json:"breakdown,omitempty"`
	RetrievedAt  time.Time                   `json:"retrievedAt"`
}

// EnergyPeriod represents a time period for energy data
type EnergyPeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// EnergyConsumptionBreakdown represents hourly/daily energy breakdown
type EnergyConsumptionBreakdown struct {
	Timestamp time.Time `json:"timestamp"`
	KWh       float64   `json:"kWh"`
	KW        float64   `json:"kW"`
	Cost      float64   `json:"cost"`
}

// TariffRequest represents the query parameters for tariff data
type TariffRequest struct {
	Region string `form:"region" binding:"required"`
}

// Tariff represents energy tariff data from external provider
type Tariff struct {
	Region        string        `json:"region"`
	Provider      string        `json:"provider"`
	EffectiveFrom time.Time     `json:"effectiveFrom"`
	EffectiveTo   *time.Time    `json:"effectiveTo,omitempty"`
	Currency      string        `json:"currency"`
	Rates         []TariffRate  `json:"rates"`
	RetrievedAt   time.Time     `json:"retrievedAt"`
}

// TariffRate represents a specific tariff rate
type TariffRate struct {
	Name          string  `json:"name"`           // e.g., "Peak", "Off-Peak", "Standard"
	RatePerKWh    float64 `json:"ratePerKWh"`
	StartHour     int     `json:"startHour"`      // 0-23
	EndHour       int     `json:"endHour"`        // 0-23
	ApplicableDays []string `json:"applicableDays"` // e.g., ["Monday", "Tuesday", ...]
}

// ExternalTokenRefreshRequest represents request to refresh external API tokens
type ExternalTokenRefreshRequest struct {
	Provider string `json:"provider" binding:"required"` // e.g., "energy_provider"
}

// ExternalTokenRefreshResponse represents response after refreshing external API tokens
type ExternalTokenRefreshResponse struct {
	Provider    string    `json:"provider"`
	Success     bool      `json:"success"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Message     string    `json:"message,omitempty"`
}

// ExternalAPIError represents an error from external API calls
type ExternalAPIError struct {
	Provider   string `json:"provider"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retryAfter,omitempty"` // seconds
}
