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

// AnomalyRepository handles anomaly database operations
type AnomalyRepository struct {
	collection *mongo.Collection
}

// NewAnomalyRepository creates a new anomaly repository
func NewAnomalyRepository(collection *mongo.Collection) *AnomalyRepository {
	return &AnomalyRepository{collection: collection}
}

// Create inserts a new anomaly
func (r *AnomalyRepository) Create(ctx context.Context, anomaly *models.Anomaly) (*models.Anomaly, error) {
	anomaly.CreatedAt = time.Now()
	anomaly.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, anomaly)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("anomaly with this ID already exists")
		}
		return nil, err
	}

	anomaly.ID = result.InsertedID.(primitive.ObjectID)
	return anomaly, nil
}

// FindByID retrieves an anomaly by its MongoDB ID
func (r *AnomalyRepository) FindByID(ctx context.Context, id string) (*models.Anomaly, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid anomaly ID format")
	}

	var anomaly models.Anomaly
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&anomaly)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("anomaly not found")
		}
		return nil, err
	}

	return &anomaly, nil
}

// FindByAnomalyID retrieves an anomaly by its anomaly_id field
func (r *AnomalyRepository) FindByAnomalyID(ctx context.Context, anomalyID string) (*models.Anomaly, error) {
	var anomaly models.Anomaly
	err := r.collection.FindOne(ctx, bson.M{"anomaly_id": anomalyID}).Decode(&anomaly)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("anomaly not found")
		}
		return nil, err
	}
	return &anomaly, nil
}

// FindAll retrieves anomalies with filters and pagination
func (r *AnomalyRepository) FindAll(ctx context.Context, deviceID, buildingID, anomalyType, severity, status string, page, limit int) ([]*models.Anomaly, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{}

	if deviceID != "" {
		filter["device_id"] = deviceID
	}
	if buildingID != "" {
		filter["building_id"] = buildingID
	}
	if anomalyType != "" {
		filter["type"] = anomalyType
	}
	if severity != "" {
		filter["severity"] = severity
	}
	if status != "" {
		filter["status"] = status
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find anomalies with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "detected_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var anomalies []*models.Anomaly
	if err := cursor.All(ctx, &anomalies); err != nil {
		return nil, 0, err
	}

	return anomalies, total, nil
}

// Update updates an anomaly
func (r *AnomalyRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Anomaly, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid anomaly ID format")
	}

	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var anomaly models.Anomaly
	if err := result.Decode(&anomaly); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("anomaly not found")
		}
		return nil, err
	}

	return &anomaly, nil
}

// CountByStatus counts anomalies by status
func (r *AnomalyRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	filter := bson.M{"status": status}
	return r.collection.CountDocuments(ctx, filter)
}

// CountByBuildingAndStatus counts anomalies by building and status
func (r *AnomalyRepository) CountByBuildingAndStatus(ctx context.Context, buildingID, status string) (int64, error) {
	filter := bson.M{"building_id": buildingID, "status": status}
	return r.collection.CountDocuments(ctx, filter)
}
