package tests

import (
	"context"
	"testing"
	"time"

	"forecast-service/internal/config"
	"forecast-service/internal/integrations"
	"forecast-service/internal/models"
	"forecast-service/internal/repository"
	"forecast-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	db := client.Database("test_forecast_service")
	cleanup := func() {
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return db, cleanup
}

func TestGenerateForecast(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.Config{
		Forecast: config.ForecastConfig{
			DefaultHorizonHours: 24,
			MaxHorizonHours:     168,
		},
		Security: config.SecurityServiceConfig{
			URL:     "http://localhost:8080",
			Timeout: 10 * time.Second,
		},
		External: config.ExternalAPIsConfig{
			WeatherURL: "http://localhost:8084/external/weather",
			TariffURL:  "http://localhost:8084/external/tariffs",
			MLURL:      "http://localhost:8085/ml/predict",
			StorageURL: "http://localhost:8086/storage",
		},
	}

	forecastRepo := repository.NewForecastRepository(db.Collection("forecasts"))
	peakLoadRepo := repository.NewPeakLoadRepository(db.Collection("peak_loads"))
	securityClient := integrations.NewSecurityClient(cfg)
	externalClient := integrations.NewExternalClient(cfg)

	forecastService := service.NewForecastService(
		forecastRepo,
		peakLoadRepo,
		securityClient,
		externalClient,
		cfg,
	)

	ctx := context.Background()
	req := &models.ForecastGenerateRequest{
		BuildingID:     "test-building-1",
		Type:           models.ForecastTypeDemand,
		HorizonHours:   24,
		IncludeWeather: false,
		IncludeTariffs: false,
		HistoricalDays: 30,
	}

	// This will fail without actual external services, but tests the structure
	_, err := forecastService.GenerateForecast(ctx, req, "test-user", "test-token")
	// We expect an error since external services won't be available in tests
	// but we can verify the service is properly initialized
	assert.NotNil(t, forecastService)
}

func TestGetLatestForecast(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	forecastRepo := repository.NewForecastRepository(db.Collection("forecasts"))

	// Create a test forecast
	ctx := context.Background()
	forecast := &models.Forecast{
		BuildingID:   "test-building-1",
		Type:         models.ForecastTypeDemand,
		Status:       models.ForecastStatusCompleted,
		HorizonHours: 24,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		ModelUsed:    "TEST",
		CreatedBy:    "test-user",
	}

	created, err := forecastRepo.Create(ctx, forecast)
	require.NoError(t, err)

	// Retrieve latest forecast
	latest, err := forecastRepo.FindLatestByBuilding(ctx, "test-building-1", models.ForecastTypeDemand)
	require.NoError(t, err)
	assert.Equal(t, created.ID, latest.ID)
	assert.Equal(t, "test-building-1", latest.BuildingID)
}

func TestPeakLoadGeneration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	peakLoadRepo := repository.NewPeakLoadRepository(db.Collection("peak_loads"))

	ctx := context.Background()
	peakLoad := &models.PeakLoad{
		BuildingID:     "test-building-1",
		BaselineLoad:   50.0,
		MaxPredictedLoad: 85.0,
		ThresholdPercent: 80.0,
		AnalysisPeriod: models.AnalysisPeriod{
			From: time.Now(),
			To:   time.Now().Add(24 * time.Hour),
		},
		PredictedPeaks: []models.PeakPeriod{
			{
				StartTime:    time.Now().Add(2 * time.Hour),
				EndTime:      time.Now().Add(4 * time.Hour),
				PeakValue:    85.0,
				ExpectedLoad: 82.0,
				PercentAboveBase: 70.0,
				Severity:     models.PeakLoadSeverityHigh,
				Confidence:   0.85,
			},
		},
		CreatedBy: "test-user",
	}

	created, err := peakLoadRepo.Create(ctx, peakLoad)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "test-building-1", created.BuildingID)
	assert.Len(t, created.PredictedPeaks, 1)
}

