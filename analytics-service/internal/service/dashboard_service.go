package service

import (
	"context"
	"time"

	"analytics-service/internal/models"
	"analytics-service/internal/repository"
)

// DashboardService handles dashboard business logic
type DashboardService struct {
	anomalyRepo *repository.AnomalyRepository
	kpiRepo     *repository.KPIRepository
	iotClient   interface {
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	}
	forecastClient interface {
		GetLatestForecast(ctx context.Context, buildingID string, authToken string) (map[string]interface{}, error)
	}
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	anomalyRepo *repository.AnomalyRepository,
	kpiRepo *repository.KPIRepository,
	iotClient interface {
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	},
	forecastClient interface {
		GetLatestForecast(ctx context.Context, buildingID string, authToken string) (map[string]interface{}, error)
	},
) *DashboardService {
	return &DashboardService{
		anomalyRepo:    anomalyRepo,
		kpiRepo:        kpiRepo,
		iotClient:      iotClient,
		forecastClient: forecastClient,
	}
}

// GetOverviewDashboard retrieves system-wide dashboard overview
func (s *DashboardService) GetOverviewDashboard(ctx context.Context, authToken string) (*models.DashboardOverview, error) {
	// Get all devices
	devices, err := s.iotClient.GetDevices(ctx, "", authToken)
	if err != nil {
		return nil, err
	}

	onlineCount := 0
	for _, device := range devices {
		if status, ok := device["status"].(string); ok && status == "ONLINE" {
			onlineCount++
		}
	}

	// Get active anomalies
	activeAnomalies, _ := s.anomalyRepo.CountByStatus(ctx, "NEW")

	// Get recent anomalies
	recentAnomalies, _, _ := s.anomalyRepo.FindAll(ctx, "", "", "", "", "", 1, 10)
	anomalyResponses := make([]models.AnomalyResponse, len(recentAnomalies))
	for i, a := range recentAnomalies {
		anomalyResponses[i] = *a.ToResponse()
	}

	// Get system-wide KPIs
	kpi, _ := s.kpiRepo.FindLatest(ctx, "", "DAILY")
	kpiMetrics := make(map[string]interface{})
	if kpi != nil {
		kpiMetrics = kpi.Metrics
	}

	return &models.DashboardOverview{
		TotalDevices:    len(devices),
		OnlineDevices:   onlineCount,
		TotalBuildings: 1, // Simplified
		ActiveAnomalies: int(activeAnomalies),
		KPIs:           kpiMetrics,
		RecentAnomalies: anomalyResponses,
		UpdatedAt:      time.Now(),
	}, nil
}

// GetBuildingDashboard retrieves building-specific dashboard
func (s *DashboardService) GetBuildingDashboard(ctx context.Context, buildingID string, authToken string) (*models.BuildingDashboard, error) {
	// Get devices for building
	devices, err := s.iotClient.GetDevices(ctx, buildingID, authToken)
	if err != nil {
		return nil, err
	}

	onlineCount := 0
	for _, device := range devices {
		if status, ok := device["status"].(string); ok && status == "ONLINE" {
			onlineCount++
		}
	}

	// Get active anomalies for building
	activeAnomalies, _ := s.anomalyRepo.CountByBuildingAndStatus(ctx, buildingID, "NEW")

	// Get building KPIs
	kpi, _ := s.kpiRepo.FindLatest(ctx, buildingID, "DAILY")
	kpiMetrics := make(map[string]interface{})
	if kpi != nil {
		kpiMetrics = kpi.Metrics
	}

	// Get forecast
	forecast, _ := s.forecastClient.GetLatestForecast(ctx, buildingID, authToken)

	return &models.BuildingDashboard{
		BuildingID:        buildingID,
		DeviceCount:       len(devices),
		OnlineDeviceCount: onlineCount,
		ActiveAnomalies:   int(activeAnomalies),
		KPIs:             kpiMetrics,
		RecentTelemetry:  []models.TimeSeriesResponse{}, // Would be populated from time-series
		UpdatedAt:        time.Now(),
	}, nil
}
