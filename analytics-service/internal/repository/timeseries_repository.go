package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"analytics-service/internal/models"
)

// TimeSeriesRepository handles time-series database operations
type TimeSeriesRepository struct {
	collection *mongo.Collection
}

// NewTimeSeriesRepository creates a new time-series repository
func NewTimeSeriesRepository(collection *mongo.Collection) *TimeSeriesRepository {
	return &TimeSeriesRepository{collection: collection}
}

// Create inserts a new time-series record
func (r *TimeSeriesRepository) Create(ctx context.Context, ts *models.TimeSeries) (*models.TimeSeries, error) {
	ts.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, ts)
	if err != nil {
		return nil, err
	}

	ts.ID = result.InsertedID.(primitive.ObjectID)
	return ts, nil
}

// CreateMany inserts multiple time-series records
func (r *TimeSeriesRepository) CreateMany(ctx context.Context, records []*models.TimeSeries) error {
	if len(records) == 0 {
		return nil
	}

	now := time.Now()
	docs := make([]interface{}, len(records))
	for i, ts := range records {
		ts.CreatedAt = now
		docs[i] = ts
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

// Query performs a time-series query with aggregation
func (r *TimeSeriesRepository) Query(ctx context.Context, req *models.TimeSeriesQueryRequest) ([]*models.TimeSeries, error) {
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": req.From,
			"$lte": req.To,
		},
		"aggregation_type": req.AggregationType,
	}

	if len(req.DeviceIDs) > 0 {
		filter["device_id"] = bson.M{"$in": req.DeviceIDs}
	}
	if req.BuildingID != "" {
		filter["building_id"] = req.BuildingID
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*models.TimeSeries
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Aggregate performs MongoDB aggregation pipeline for time-series data
func (r *TimeSeriesRepository) Aggregate(ctx context.Context, pipeline []bson.M) ([]bson.M, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// FindLatestByDevice retrieves the latest time-series record for a device
func (r *TimeSeriesRepository) FindLatestByDevice(ctx context.Context, deviceID string) (*models.TimeSeries, error) {
	filter := bson.M{"device_id": deviceID}
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var ts models.TimeSeries
	err := r.collection.FindOne(ctx, filter, opts).Decode(&ts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("no time-series data found for device")
		}
		return nil, err
	}

	return &ts, nil
}
