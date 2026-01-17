// Package service contains business logic for the application
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"analytics-service/internal/models"
	"analytics-service/internal/repository"
)

// ReportService handles report business logic
type ReportService struct {
	reportRepo   *repository.ReportRepository
	iotClient    interface {
		GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error)
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	}
	forecastClient interface {
		GetLatestForecast(ctx context.Context, buildingID string, authToken string) (map[string]interface{}, error)
	}
}

// NewReportService creates a new report service
func NewReportService(
	reportRepo *repository.ReportRepository,
	iotClient interface {
		GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error)
		GetDevices(ctx context.Context, buildingID string, authToken string) ([]map[string]interface{}, error)
	},
	forecastClient interface {
		GetLatestForecast(ctx context.Context, buildingID string, authToken string) (map[string]interface{}, error)
	},
) *ReportService {
	return &ReportService{
		reportRepo:    reportRepo,
		iotClient:     iotClient,
		forecastClient: forecastClient,
	}
}

// GenerateReport generates an analytical report
func (s *ReportService) GenerateReport(ctx context.Context, req *models.GenerateReportRequest, userID, authToken string) (*models.ReportResponse, error) {
	// Validate request
	if err := s.validateGenerateReport(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate report ID
	reportID := uuid.New().String()

	// Create report record in pending state
	report := &models.Report{
		ReportID:    reportID,
		BuildingID:  req.BuildingID,
		Type:        req.Type,
		Status:      models.ReportStatusGenerating,
		Content:     make(map[string]interface{}),
		GeneratedAt: time.Now(),
		GeneratedBy: userID,
	}

	createdReport, err := s.reportRepo.Create(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Generate report content asynchronously
	go s.generateReportContent(context.Background(), createdReport, req, authToken)

	return createdReport.ToResponse(), nil
}

// generateReportContent generates the actual report content
func (s *ReportService) generateReportContent(ctx context.Context, report *models.Report, req *models.GenerateReportRequest, authToken string) {
	content := make(map[string]interface{})

	// Get devices for the building
	devices, err := s.iotClient.GetDevices(ctx, req.BuildingID, authToken)
	if err == nil {
		content["devices"] = devices
		content["deviceCount"] = len(devices)
	}

	// Get forecast data if available
	if req.BuildingID != "" {
		forecast, err := s.forecastClient.GetLatestForecast(ctx, req.BuildingID, authToken)
		if err == nil {
			content["forecast"] = forecast
		}
	}

	// Generate report based on type
	switch req.Type {
	case "ENERGY_CONSUMPTION":
		content = s.generateEnergyConsumptionReport(ctx, req, devices, authToken)
	case "DEVICE_PERFORMANCE":
		content = s.generateDevicePerformanceReport(ctx, req, devices, authToken)
	case "ANOMALY_SUMMARY":
		content = s.generateAnomalySummaryReport(ctx, req)
	default:
		content["summary"] = "General report"
		content["generatedAt"] = time.Now()
	}

	// Update report with content
	updates := bson.M{
		"content":      content,
		"status":        models.ReportStatusCompleted,
		"generated_at":  time.Now(),
	}

	_, err := s.reportRepo.Update(ctx, report.ID.Hex(), updates)
	if err != nil {
		log.Printf("Failed to update report: %v", err)
	}
}

// generateEnergyConsumptionReport generates energy consumption report
func (s *ReportService) generateEnergyConsumptionReport(ctx context.Context, req *models.GenerateReportRequest, devices []map[string]interface{}, authToken string) map[string]interface{} {
	content := make(map[string]interface{})
	content["type"] = "ENERGY_CONSUMPTION"
	content["period"] = map[string]interface{}{
		"from": req.From,
		"to":   req.To,
	}

	totalConsumption := 0.0
	deviceConsumptions := make([]map[string]interface{}, 0)

	for _, device := range devices {
		deviceID, _ := device["deviceId"].(string)
		if deviceID == "" {
			continue
		}

		// Get telemetry history for device
		telemetry, err := s.iotClient.GetTelemetryHistory(ctx, deviceID, req.From, req.To, 1, 100, authToken)
		if err != nil {
			continue
		}

		deviceTotal := 0.0
		for _, t := range telemetry {
			if metrics, ok := t["metrics"].(map[string]interface{}); ok {
				if consumption, ok := metrics["consumption"].(float64); ok {
					deviceTotal += consumption
				}
			}
		}

		totalConsumption += deviceTotal
		deviceConsumptions = append(deviceConsumptions, map[string]interface{}{
			"deviceId":   deviceID,
			"consumption": deviceTotal,
		})
	}

	content["totalConsumption"] = totalConsumption
	content["deviceConsumptions"] = deviceConsumptions
	content["averageConsumption"] = totalConsumption / float64(len(devices))

	return content
}

// generateDevicePerformanceReport generates device performance report
func (s *ReportService) generateDevicePerformanceReport(ctx context.Context, req *models.GenerateReportRequest, devices []map[string]interface{}, authToken string) map[string]interface{} {
	content := make(map[string]interface{})
	content["type"] = "DEVICE_PERFORMANCE"
	content["period"] = map[string]interface{}{
		"from": req.From,
		"to":   req.To,
	}

	devicePerformances := make([]map[string]interface{}, 0)
	for _, device := range devices {
		deviceID, _ := device["deviceId"].(string)
		if deviceID == "" {
			continue
		}

		devicePerformances = append(devicePerformances, map[string]interface{}{
			"deviceId": deviceID,
			"status":    device["status"],
			"lastSeen": device["lastSeen"],
		})
	}

	content["devices"] = devicePerformances
	return content
}

// generateAnomalySummaryReport generates anomaly summary report
func (s *ReportService) generateAnomalySummaryReport(ctx context.Context, req *models.GenerateReportRequest) map[string]interface{} {
	content := make(map[string]interface{})
	content["type"] = "ANOMALY_SUMMARY"
	// This would typically query the anomaly repository
	content["summary"] = "Anomaly summary report"
	return content
}

// GetReport retrieves a report by ID
func (s *ReportService) GetReport(ctx context.Context, reportID string) (*models.ReportResponse, error) {
	report, err := s.reportRepo.FindByReportID(ctx, reportID)
	if err != nil {
		return nil, err
	}
	return report.ToResponse(), nil
}

// ListReports lists reports with filters
func (s *ReportService) ListReports(ctx context.Context, buildingID, reportType, status string, page, limit int) ([]*models.ReportResponse, int64, error) {
	reports, total, err := s.reportRepo.FindAll(ctx, buildingID, reportType, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = report.ToResponse()
	}

	return responses, total, nil
}

// validateGenerateReport validates report generation request
func (s *ReportService) validateGenerateReport(req *models.GenerateReportRequest) error {
	if req.Type == "" {
		return fmt.Errorf("report type is required")
	}
	return nil
}
