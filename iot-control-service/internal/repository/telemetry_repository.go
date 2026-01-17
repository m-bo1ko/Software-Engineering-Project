package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"iot-control-service/internal/models"
)

// TelemetryRepository handles telemetry database operations
type TelemetryRepository struct {
	collection *mongo.Collection
}

// NewTelemetryRepository creates a new telemetry repository
func NewTelemetryRepository(collection *mongo.Collection) *TelemetryRepository {
	return &TelemetryRepository{collection: collection}
}

// Create inserts a new telemetry record
func (r *TelemetryRepository) Create(ctx context.Context, telemetry *models.Telemetry) (*models.Telemetry, error) {
	if telemetry.Timestamp.IsZero() {
		telemetry.Timestamp = time.Now()
	}
	telemetry.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, telemetry)
	if err != nil {
		return nil, err
	}

	telemetry.ID = result.InsertedID.(primitive.ObjectID)
	return telemetry, nil
}

// CreateMany inserts multiple telemetry records
func (r *TelemetryRepository) CreateMany(ctx context.Context, telemetry []*models.Telemetry) error {
	if len(telemetry) == 0 {
		return nil
	}

	now := time.Now()
	docs := make([]interface{}, len(telemetry))
	for i, t := range telemetry {
		if t.Timestamp.IsZero() {
			t.Timestamp = now
		}
		t.CreatedAt = now
		docs[i] = t
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

// FindByDeviceID retrieves telemetry for a device with pagination
func (r *TelemetryRepository) FindByDeviceID(ctx context.Context, deviceID string, from, to time.Time, page, limit int) ([]*models.Telemetry, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 100
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{"device_id": deviceID}

	if !from.IsZero() {
		filter["timestamp"] = bson.M{"$gte": from}
	}
	if !to.IsZero() {
		if filter["timestamp"] == nil {
			filter["timestamp"] = bson.M{"$lte": to}
		} else {
			filter["timestamp"].(bson.M)["$lte"] = to
		}
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find telemetry with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var telemetry []*models.Telemetry
	if err := cursor.All(ctx, &telemetry); err != nil {
		return nil, 0, err
	}

	return telemetry, total, nil
}

// FindLatestByDevice retrieves the latest telemetry for a device
func (r *TelemetryRepository) FindLatestByDevice(ctx context.Context, deviceID string) (*models.Telemetry, error) {
	filter := bson.M{"device_id": deviceID}
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var telemetry models.Telemetry
	err := r.collection.FindOne(ctx, filter, opts).Decode(&telemetry)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("no telemetry found for device")
		}
		return nil, err
	}

	return &telemetry, nil
}

// FindLatestMetricsByDevice retrieves latest metrics for multiple devices
func (r *TelemetryRepository) FindLatestMetricsByDevice(ctx context.Context, deviceIDs []string) (map[string]*models.Telemetry, error) {
	if len(deviceIDs) == 0 {
		return make(map[string]*models.Telemetry), nil
	}

	pipeline := []bson.M{
		{"$match": bson.M{"device_id": bson.M{"$in": deviceIDs}}},
		{"$sort": bson.M{"timestamp": -1}},
		{
			"$group": bson.M{
				"_id": "$device_id",
				"latest": bson.M{"$first": "$$ROOT"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]*models.Telemetry)
	for cursor.Next(ctx) {
		var doc struct {
			ID     string            `bson:"_id"`
			Latest models.Telemetry `bson:"latest"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		result[doc.ID] = &doc.Latest
	}

	return result, nil
}
