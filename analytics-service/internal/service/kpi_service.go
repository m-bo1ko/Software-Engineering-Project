package service

import (
	"context"
	"fmt"
	"time"

	"analytics-service/internal/models"
	"analytics-service/internal/repository"
)

// KPIService handles KPI calculation business logic
type KPIService struct {
	kpiRepo   *repository.KPIRepository
	anomalyRepo *repository.AnomalyRepository
	iotClient interface {
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	}
}

// NewKPIService creates a new KPI service
func NewKPIService(
	kpiRepo *repository.KPIRepository,
	anomalyRepo *repository.AnomalyRepository,
	iotClient interface {
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	},
) *KPIService {
	return &KPIService{
		kpiRepo:    kpiRepo,
		anomalyRepo: anomalyRepo,
		iotClient:  iotClient,
	}
}

// CalculateKPIs calculates KPIs for a building or system-wide
func (s *KPIService) CalculateKPIs(ctx context.Context, buildingID, period string, authToken string) (*models.KPIResponse, error) {
	metrics := make(map[string]interface{})

	// Get devices
	devices, err := s.iotClient.GetDevices(ctx, buildingID, authToken)
	if err == nil {
		onlineCount := 0
		for _, device := range devices {
			if status, ok := device["status"].(string); ok && status == "ONLINE" {
				onlineCount++
			}
		}

		metrics["totalDevices"] = len(devices)
		metrics["onlineDevices"] = onlineCount
		metrics["deviceAvailability"] = float64(onlineCount) / float64(len(devices)) * 100
	}

	// Get anomaly count
	anomalyCount, _ := s.anomalyRepo.CountByStatus(ctx, "NEW")
	metrics["activeAnomalies"] = anomalyCount

	// Create or update KPI
	kpi := &models.KPI{
		BuildingID:   buildingID,
		CalculatedAt: time.Now(),
		Metrics:     metrics,
		Period:      period,
	}

	updated, err := s.kpiRepo.UpdateOrCreate(ctx, kpi)
	if err != nil {
		return nil, fmt.Errorf("failed to save KPI: %w", err)
	}

	return updated.ToResponse(), nil
}

// GetKPIs retrieves KPIs for a building or system-wide
func (s *KPIService) GetKPIs(ctx context.Context, buildingID, period string) (*models.KPIResponse, error) {
	if period == "" {
		period = "DAILY"
	}

	kpi, err := s.kpiRepo.FindLatest(ctx, buildingID, period)
	if err != nil {
		return nil, err
	}

	return kpi.ToResponse(), nil
}
