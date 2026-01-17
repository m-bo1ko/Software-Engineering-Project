package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PeakLoadSeverity represents the severity level of a peak load prediction
type PeakLoadSeverity string

const (
	PeakLoadSeverityLow      PeakLoadSeverity = "LOW"
	PeakLoadSeverityMedium   PeakLoadSeverity = "MEDIUM"
	PeakLoadSeverityHigh     PeakLoadSeverity = "HIGH"
	PeakLoadSeverityCritical PeakLoadSeverity = "CRITICAL"
)

// PeakLoad represents a peak load forecast
type PeakLoad struct {
	ID               primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	BuildingID       string                `bson:"building_id" json:"buildingId"`
	ForecastID       string                `bson:"forecast_id,omitempty" json:"forecastId,omitempty"`
	PredictedPeaks   []PeakPeriod          `bson:"predicted_peaks" json:"predictedPeaks"`
	BaselineLoad     float64               `bson:"baseline_load" json:"baselineLoad"`
	MaxPredictedLoad float64               `bson:"max_predicted_load" json:"maxPredictedLoad"`
	ThresholdPercent float64               `bson:"threshold_percent" json:"thresholdPercent"`
	AnalysisPeriod   AnalysisPeriod        `bson:"analysis_period" json:"analysisPeriod"`
	Contributing     []ContributingFactor  `bson:"contributing_factors" json:"contributingFactors"`
	Recommendations  []string              `bson:"recommendations" json:"recommendations"`
	CreatedAt        time.Time             `bson:"created_at" json:"createdAt"`
	CreatedBy        string                `bson:"created_by" json:"createdBy"`
}

// PeakPeriod represents a predicted period of high energy consumption
type PeakPeriod struct {
	StartTime         time.Time        `bson:"start_time" json:"startTime"`
	EndTime           time.Time        `bson:"end_time" json:"endTime"`
	PeakValue         float64          `bson:"peak_value" json:"peakValue"`
	ExpectedLoad      float64          `bson:"expected_load" json:"expectedLoad"`
	PercentAboveBase  float64          `bson:"percent_above_base" json:"percentAboveBase"`
	Severity          PeakLoadSeverity `bson:"severity" json:"severity"`
	Confidence        float64          `bson:"confidence" json:"confidence"`
	MitigationActions []string         `bson:"mitigation_actions" json:"mitigationActions"`
}

// AnalysisPeriod represents the time period analyzed for peak loads
type AnalysisPeriod struct {
	From time.Time `bson:"from" json:"from"`
	To   time.Time `bson:"to" json:"to"`
}

// ContributingFactor represents a factor contributing to peak load
type ContributingFactor struct {
	Factor      string  `bson:"factor" json:"factor"`           // e.g., "HVAC", "Lighting", "Weather"
	Impact      float64 `bson:"impact" json:"impact"`           // Percentage contribution
	Description string  `bson:"description" json:"description"`
}

// PeakLoadRequest represents the request to generate peak load predictions
type PeakLoadRequest struct {
	BuildingID       string    `json:"buildingId" binding:"required"`
	AnalysisFromDate time.Time `json:"analysisFromDate"`
	AnalysisToDate   time.Time `json:"analysisToDate"`
	ThresholdPercent float64   `json:"thresholdPercent"` // Percentage above baseline to consider peak
	IncludeWeather   bool      `json:"includeWeather"`
}

// PeakLoadResponse represents the peak load data returned in API responses
type PeakLoadResponse struct {
	ID               string               `json:"id"`
	BuildingID       string               `json:"buildingId"`
	ForecastID       string               `json:"forecastId,omitempty"`
	PredictedPeaks   []PeakPeriod         `json:"predictedPeaks"`
	BaselineLoad     float64              `json:"baselineLoad"`
	MaxPredictedLoad float64              `json:"maxPredictedLoad"`
	ThresholdPercent float64              `json:"thresholdPercent"`
	AnalysisPeriod   AnalysisPeriod       `json:"analysisPeriod"`
	Contributing     []ContributingFactor `json:"contributingFactors"`
	Recommendations  []string             `json:"recommendations"`
	CreatedAt        time.Time            `json:"createdAt"`
}

// ToResponse converts a PeakLoad to PeakLoadResponse
func (p *PeakLoad) ToResponse() *PeakLoadResponse {
	return &PeakLoadResponse{
		ID:               p.ID.Hex(),
		BuildingID:       p.BuildingID,
		ForecastID:       p.ForecastID,
		PredictedPeaks:   p.PredictedPeaks,
		BaselineLoad:     p.BaselineLoad,
		MaxPredictedLoad: p.MaxPredictedLoad,
		ThresholdPercent: p.ThresholdPercent,
		AnalysisPeriod:   p.AnalysisPeriod,
		Contributing:     p.Contributing,
		Recommendations:  p.Recommendations,
		CreatedAt:        p.CreatedAt,
	}
}

// PeakLoadSummary provides a summary of peak load analysis
type PeakLoadSummary struct {
	BuildingID          string  `json:"buildingId"`
	TotalPeaksDetected  int     `json:"totalPeaksDetected"`
	CriticalPeaks       int     `json:"criticalPeaks"`
	HighPeaks           int     `json:"highPeaks"`
	MediumPeaks         int     `json:"mediumPeaks"`
	LowPeaks            int     `json:"lowPeaks"`
	AveragePeakDuration float64 `json:"averagePeakDuration"` // in hours
	MaxPeakValue        float64 `json:"maxPeakValue"`
	EstimatedCostImpact float64 `json:"estimatedCostImpact"`
}
