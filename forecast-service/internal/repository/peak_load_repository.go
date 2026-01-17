package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"forecast-service/internal/models"
)

// PeakLoadRepository handles peak load database operations
type PeakLoadRepository struct {
	collection *mongo.Collection
}

// NewPeakLoadRepository creates a new peak load repository
func NewPeakLoadRepository(collection *mongo.Collection) *PeakLoadRepository {
	return &PeakLoadRepository{collection: collection}
}

// Create inserts a new peak load prediction into the database
func (r *PeakLoadRepository) Create(ctx context.Context, peakLoad *models.PeakLoad) (*models.PeakLoad, error) {
	peakLoad.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, peakLoad)
	if err != nil {
		return nil, err
	}

	peakLoad.ID = result.InsertedID.(primitive.ObjectID)
	return peakLoad, nil
}

// FindByID retrieves a peak load by its ID
func (r *PeakLoadRepository) FindByID(ctx context.Context, id string) (*models.PeakLoad, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid peak load ID format")
	}

	var peakLoad models.PeakLoad
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&peakLoad)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("peak load not found")
		}
		return nil, err
	}

	return &peakLoad, nil
}

// FindLatestByBuilding retrieves the latest peak load for a building
func (r *PeakLoadRepository) FindLatestByBuilding(ctx context.Context, buildingID string) (*models.PeakLoad, error) {
	filter := bson.M{"building_id": buildingID}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var peakLoad models.PeakLoad
	err := r.collection.FindOne(ctx, filter, opts).Decode(&peakLoad)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("no peak load predictions found for this building")
		}
		return nil, err
	}

	return &peakLoad, nil
}

// FindByBuilding retrieves peak loads for a building with pagination
func (r *PeakLoadRepository) FindByBuilding(ctx context.Context, buildingID string, page, limit int) ([]*models.PeakLoad, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{"building_id": buildingID}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find peak loads with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var peakLoads []*models.PeakLoad
	if err := cursor.All(ctx, &peakLoads); err != nil {
		return nil, 0, err
	}

	return peakLoads, total, nil
}

// FindByForecast retrieves peak loads associated with a forecast
func (r *PeakLoadRepository) FindByForecast(ctx context.Context, forecastID string) ([]*models.PeakLoad, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"forecast_id": forecastID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var peakLoads []*models.PeakLoad
	if err := cursor.All(ctx, &peakLoads); err != nil {
		return nil, err
	}

	return peakLoads, nil
}

// FindUpcomingPeaks retrieves upcoming peak predictions within a time range
func (r *PeakLoadRepository) FindUpcomingPeaks(ctx context.Context, buildingID string, from, to time.Time) ([]*models.PeakLoad, error) {
	filter := bson.M{
		"building_id": buildingID,
		"predicted_peaks": bson.M{
			"$elemMatch": bson.M{
				"start_time": bson.M{"$gte": from, "$lte": to},
			},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var peakLoads []*models.PeakLoad
	if err := cursor.All(ctx, &peakLoads); err != nil {
		return nil, err
	}

	return peakLoads, nil
}

// Delete removes a peak load from the database
func (r *PeakLoadRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid peak load ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("peak load not found")
	}

	return nil
}

// GetPeakLoadSummary generates a summary of peak loads for a building
func (r *PeakLoadRepository) GetPeakLoadSummary(ctx context.Context, buildingID string, from, to time.Time) (*models.PeakLoadSummary, error) {
	peakLoads, err := r.FindUpcomingPeaks(ctx, buildingID, from, to)
	if err != nil {
		return nil, err
	}

	summary := &models.PeakLoadSummary{
		BuildingID: buildingID,
	}

	var totalDuration float64
	var maxPeak float64

	for _, pl := range peakLoads {
		for _, peak := range pl.PredictedPeaks {
			summary.TotalPeaksDetected++

			duration := peak.EndTime.Sub(peak.StartTime).Hours()
			totalDuration += duration

			if peak.PeakValue > maxPeak {
				maxPeak = peak.PeakValue
			}

			switch peak.Severity {
			case models.PeakLoadSeverityCritical:
				summary.CriticalPeaks++
			case models.PeakLoadSeverityHigh:
				summary.HighPeaks++
			case models.PeakLoadSeverityMedium:
				summary.MediumPeaks++
			case models.PeakLoadSeverityLow:
				summary.LowPeaks++
			}
		}
	}

	if summary.TotalPeaksDetected > 0 {
		summary.AveragePeakDuration = totalDuration / float64(summary.TotalPeaksDetected)
	}
	summary.MaxPeakValue = maxPeak

	return summary, nil
}
