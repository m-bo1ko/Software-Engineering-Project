// Package models defines the data structures used throughout the application
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ForecastStatus represents the status of a forecast
type ForecastStatus string

const (
	ForecastStatusPending    ForecastStatus = "PENDING"
	ForecastStatusProcessing ForecastStatus = "PROCESSING"
	ForecastStatusCompleted  ForecastStatus = "COMPLETED"
	ForecastStatusFailed     ForecastStatus = "FAILED"
)

// ForecastType represents the type of forecast
type ForecastType string

const (
	ForecastTypeDemand      ForecastType = "DEMAND"
	ForecastTypeConsumption ForecastType = "CONSUMPTION"
	ForecastTypeLoad        ForecastType = "LOAD"
)

// Forecast represents an energy demand forecast
type Forecast struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	BuildingID      string               `bson:"building_id" json:"buildingId"`
	DeviceID        string               `bson:"device_id,omitempty" json:"deviceId,omitempty"`
	Type            ForecastType         `bson:"type" json:"type"`
	Status          ForecastStatus       `bson:"status" json:"status"`
	HorizonHours    int                  `bson:"horizon_hours" json:"horizonHours"`
	StartTime       time.Time            `bson:"start_time" json:"startTime"`
	EndTime         time.Time            `bson:"end_time" json:"endTime"`
	Predictions     []ForecastPrediction `bson:"predictions" json:"predictions"`
	Accuracy        *ForecastAccuracy    `bson:"accuracy,omitempty" json:"accuracy,omitempty"`
	ModelUsed       string               `bson:"model_used" json:"modelUsed"`
	InputParameters ForecastInputParams  `bson:"input_parameters" json:"inputParameters"`
	Metadata        map[string]string    `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt       time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updatedAt"`
	CreatedBy       string               `bson:"created_by" json:"createdBy"`
	ErrorMessage    string               `bson:"error_message,omitempty" json:"errorMessage,omitempty"`
}

// ForecastPrediction represents a single prediction data point
type ForecastPrediction struct {
	Timestamp       time.Time `bson:"timestamp" json:"timestamp"`
	PredictedValue  float64   `bson:"predicted_value" json:"predictedValue"`
	LowerBound      float64   `bson:"lower_bound" json:"lowerBound"`
	UpperBound      float64   `bson:"upper_bound" json:"upperBound"`
	ConfidenceLevel float64   `bson:"confidence_level" json:"confidenceLevel"`
	Unit            string    `bson:"unit" json:"unit"` // kWh, kW, etc.
}

// ForecastAccuracy represents forecast accuracy metrics
type ForecastAccuracy struct {
	MAE   float64 `bson:"mae" json:"mae"`     // Mean Absolute Error
	RMSE  float64 `bson:"rmse" json:"rmse"`   // Root Mean Square Error
	MAPE  float64 `bson:"mape" json:"mape"`   // Mean Absolute Percentage Error
	Score float64 `bson:"score" json:"score"` // Overall accuracy score (0-100)
}

// ForecastInputParams represents input parameters used for forecast generation
type ForecastInputParams struct {
	HistoricalDays    int       `bson:"historical_days" json:"historicalDays"`
	IncludeWeather    bool      `bson:"include_weather" json:"includeWeather"`
	IncludeTariffs    bool      `bson:"include_tariffs" json:"includeTariffs"`
	SeasonalFactors   bool      `bson:"seasonal_factors" json:"seasonalFactors"`
	WeatherData       *Weather  `bson:"weather_data,omitempty" json:"weatherData,omitempty"`
	TariffData        *Tariff   `bson:"tariff_data,omitempty" json:"tariffData,omitempty"`
}

// Weather represents weather data used in forecasting
type Weather struct {
	Temperature     float64 `bson:"temperature" json:"temperature"`
	Humidity        float64 `bson:"humidity" json:"humidity"`
	CloudCover      float64 `bson:"cloud_cover" json:"cloudCover"`
	WindSpeed       float64 `bson:"wind_speed" json:"windSpeed"`
	Condition       string  `bson:"condition" json:"condition"`
	ForecastedHigh  float64 `bson:"forecasted_high" json:"forecastedHigh"`
	ForecastedLow   float64 `bson:"forecasted_low" json:"forecastedLow"`
}

// Tariff represents tariff data used in forecasting
type Tariff struct {
	Region        string       `bson:"region" json:"region"`
	CurrentRate   float64      `bson:"current_rate" json:"currentRate"`
	PeakRate      float64      `bson:"peak_rate" json:"peakRate"`
	OffPeakRate   float64      `bson:"off_peak_rate" json:"offPeakRate"`
	Currency      string       `bson:"currency" json:"currency"`
	TimeOfUseRates []TariffRate `bson:"time_of_use_rates,omitempty" json:"timeOfUseRates,omitempty"`
}

// TariffRate represents a time-of-use tariff rate
type TariffRate struct {
	Name       string  `bson:"name" json:"name"`
	RatePerKWh float64 `bson:"rate_per_kwh" json:"ratePerKWh"`
	StartHour  int     `bson:"start_hour" json:"startHour"`
	EndHour    int     `bson:"end_hour" json:"endHour"`
}

// ForecastGenerateRequest represents the request to generate a forecast
type ForecastGenerateRequest struct {
	BuildingID     string            `json:"buildingId" binding:"required"`
	DeviceID       string            `json:"deviceId"`
	Type           ForecastType      `json:"type" binding:"required"`
	HorizonHours   int               `json:"horizonHours"`
	IncludeWeather bool              `json:"includeWeather"`
	IncludeTariffs bool              `json:"includeTariffs"`
	HistoricalDays int               `json:"historicalDays"`
	Metadata       map[string]string `json:"metadata"`
}

// ForecastResponse represents the forecast data returned in API responses
type ForecastResponse struct {
	ID              string               `json:"id"`
	BuildingID      string               `json:"buildingId"`
	DeviceID        string               `json:"deviceId,omitempty"`
	Type            ForecastType         `json:"type"`
	Status          ForecastStatus       `json:"status"`
	HorizonHours    int                  `json:"horizonHours"`
	StartTime       time.Time            `json:"startTime"`
	EndTime         time.Time            `json:"endTime"`
	Predictions     []ForecastPrediction `json:"predictions"`
	Accuracy        *ForecastAccuracy    `json:"accuracy,omitempty"`
	ModelUsed       string               `json:"modelUsed"`
	CreatedAt       time.Time            `json:"createdAt"`
	ErrorMessage    string               `json:"errorMessage,omitempty"`
}

// ToResponse converts a Forecast to ForecastResponse
func (f *Forecast) ToResponse() *ForecastResponse {
	return &ForecastResponse{
		ID:           f.ID.Hex(),
		BuildingID:   f.BuildingID,
		DeviceID:     f.DeviceID,
		Type:         f.Type,
		Status:       f.Status,
		HorizonHours: f.HorizonHours,
		StartTime:    f.StartTime,
		EndTime:      f.EndTime,
		Predictions:  f.Predictions,
		Accuracy:     f.Accuracy,
		ModelUsed:    f.ModelUsed,
		CreatedAt:    f.CreatedAt,
		ErrorMessage: f.ErrorMessage,
	}
}

// DevicePrediction represents predicted consumption for a specific device
type DevicePrediction struct {
	DeviceID           string               `json:"deviceId"`
	DeviceName         string               `json:"deviceName"`
	DeviceType         string               `json:"deviceType"`
	CurrentConsumption float64              `json:"currentConsumption"`
	PredictedValues    []ForecastPrediction `json:"predictedValues"`
	Trend              string               `json:"trend"` // INCREASING, DECREASING, STABLE
	TrendPercentage    float64              `json:"trendPercentage"`
}

// DeviceOptimization represents optimization recommendations for a device
type DeviceOptimization struct {
	DeviceID          string                  `json:"deviceId"`
	DeviceName        string                  `json:"deviceName"`
	CurrentState      string                  `json:"currentState"`
	OptimalState      string                  `json:"optimalState"`
	PotentialSavings  float64                 `json:"potentialSavings"`
	SavingsUnit       string                  `json:"savingsUnit"`
	Recommendations   []string                `json:"recommendations"`
	ScheduledActions  []ScheduledAction       `json:"scheduledActions,omitempty"`
	Priority          string                  `json:"priority"` // HIGH, MEDIUM, LOW
}

// ScheduledAction represents a scheduled optimization action
type ScheduledAction struct {
	Time        time.Time `json:"time"`
	Action      string    `json:"action"`
	TargetState string    `json:"targetState"`
	Reason      string    `json:"reason"`
}
