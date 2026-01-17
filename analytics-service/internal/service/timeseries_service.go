package service

import (
	"context"
	"fmt"
	"time"

	"analytics-service/internal/models"
	"analytics-service/internal/repository"
)

// TimeSeriesService handles time-series data business logic
type TimeSeriesService struct {
	timeSeriesRepo *repository.TimeSeriesRepository
	iotClient      interface {
		GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error)
	}
}

// NewTimeSeriesService creates a new time-series service
func NewTimeSeriesService(
	timeSeriesRepo *repository.TimeSeriesRepository,
	iotClient interface {
		GetTelemetryHistory(ctx context.Context, deviceID string, from, to time.Time, page, limit int, authToken string) ([]map[string]interface{}, error)
	},
) *TimeSeriesService {
	return &TimeSeriesService{
		timeSeriesRepo: timeSeriesRepo,
		iotClient:      iotClient,
	}
}

// QueryTimeSeries queries time-series data with aggregation
func (s *TimeSeriesService) QueryTimeSeries(ctx context.Context, req *models.TimeSeriesQueryRequest, authToken string) ([]*models.TimeSeriesResponse, error) {
	// Validate request
	if err := s.validateQueryRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert aggregation type
	aggType := models.AggregationType(req.AggregationType)

	// If we have device IDs, fetch from IoT service and aggregate
	if len(req.DeviceIDs) > 0 {
		return s.queryFromIoTService(ctx, req, aggType, authToken)
	}

	// Otherwise, query from time-series collection
	results, err := s.timeSeriesRepo.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.TimeSeriesResponse, len(results))
	for i, ts := range results {
		responses[i] = ts.ToResponse()
	}

	return responses, nil
}

// queryFromIoTService queries telemetry from IoT service and aggregates it
func (s *TimeSeriesService) queryFromIoTService(ctx context.Context, req *models.TimeSeriesQueryRequest, aggType models.AggregationType, authToken string) ([]*models.TimeSeriesResponse, error) {
	allData := make([]map[string]interface{}, 0)

	// Fetch telemetry for each device
	for _, deviceID := range req.DeviceIDs {
		telemetry, err := s.iotClient.GetTelemetryHistory(ctx, deviceID, req.From, req.To, 1, 1000, authToken)
		if err != nil {
			continue
		}
		allData = append(allData, telemetry...)
	}

	// Aggregate data based on type
	aggregated := s.aggregateData(allData, aggType, req.BuildingID)

	responses := make([]*models.TimeSeriesResponse, len(aggregated))
	for i, ts := range aggregated {
		responses[i] = ts.ToResponse()
	}

	return responses, nil
}

// aggregateData aggregates telemetry data by time period
func (s *TimeSeriesService) aggregateData(data []map[string]interface{}, aggType models.AggregationType, buildingID string) []*models.TimeSeries {
	// Group by time period
	groups := make(map[string][]map[string]interface{})

	for _, item := range data {
		timestamp, ok := item["timestamp"].(string)
		if !ok {
			continue
		}

		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			continue
		}

		var key string
		switch aggType {
		case models.AggregationTypeHourly:
			key = t.Truncate(time.Hour).Format(time.RFC3339)
		case models.AggregationTypeDaily:
			key = t.Truncate(24 * time.Hour).Format(time.RFC3339)
		case models.AggregationTypeWeekly:
			key = t.Truncate(7 * 24 * time.Hour).Format(time.RFC3339)
		case models.AggregationTypeMonthly:
			key = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).Format(time.RFC3339)
		default:
			key = t.Truncate(time.Hour).Format(time.RFC3339)
		}

		groups[key] = append(groups[key], item)
	}

	// Aggregate each group
	results := make([]*models.TimeSeries, 0)
	for key, group := range groups {
		t, _ := time.Parse(time.RFC3339, key)
		metrics := s.aggregateMetrics(group)

		deviceID := ""
		if len(group) > 0 {
			if d, ok := group[0]["deviceId"].(string); ok {
				deviceID = d
			}
		}

		ts := &models.TimeSeries{
			DeviceID:        deviceID,
			BuildingID:      buildingID,
			Timestamp:       t,
			AggregationType: aggType,
			Metrics:         metrics,
		}

		results = append(results, ts)
	}

	return results
}

// aggregateMetrics aggregates metrics from multiple data points
func (s *TimeSeriesService) aggregateMetrics(data []map[string]interface{}) map[string]interface{} {
	metrics := make(map[string]interface{})
	metricSums := make(map[string]float64)
	metricCounts := make(map[string]int)

	for _, item := range data {
		if itemMetrics, ok := item["metrics"].(map[string]interface{}); ok {
			for key, value := range itemMetrics {
				if num, ok := value.(float64); ok {
					metricSums[key] += num
					metricCounts[key]++
				}
			}
		}
	}

	// Calculate averages
	for key, sum := range metricSums {
		if count := metricCounts[key]; count > 0 {
			metrics[key] = sum / float64(count)
		}
	}

	return metrics
}

// validateQueryRequest validates time-series query request
func (s *TimeSeriesService) validateQueryRequest(req *models.TimeSeriesQueryRequest) error {
	if req.From.IsZero() || req.To.IsZero() {
		return fmt.Errorf("from and to timestamps are required")
	}
	if req.From.After(req.To) {
		return fmt.Errorf("from timestamp must be before to timestamp")
	}
	if req.AggregationType == "" {
		return fmt.Errorf("aggregation type is required")
	}
	return nil
}
