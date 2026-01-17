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

// KPIRepository handles KPI database operations
type KPIRepository struct {
	collection *mongo.Collection
}

// NewKPIRepository creates a new KPI repository
func NewKPIRepository(collection *mongo.Collection) *KPIRepository {
	return &KPIRepository{collection: collection}
}

// Create inserts a new KPI record
func (r *KPIRepository) Create(ctx context.Context, kpi *models.KPI) (*models.KPI, error) {
	kpi.CreatedAt = time.Now()
	kpi.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, kpi)
	if err != nil {
		return nil, err
	}

	kpi.ID = result.InsertedID.(primitive.ObjectID)
	return kpi, nil
}

// FindLatest retrieves the latest KPI for a building (or system-wide if buildingID is empty)
func (r *KPIRepository) FindLatest(ctx context.Context, buildingID, period string) (*models.KPI, error) {
	filter := bson.M{"period": period}
	if buildingID != "" {
		filter["building_id"] = buildingID
	} else {
		filter["building_id"] = bson.M{"$exists": false}
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "calculated_at", Value: -1}})

	var kpi models.KPI
	err := r.collection.FindOne(ctx, filter, opts).Decode(&kpi)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("KPI not found")
		}
		return nil, err
	}

	return &kpi, nil
}

// UpdateOrCreate updates an existing KPI or creates a new one
func (r *KPIRepository) UpdateOrCreate(ctx context.Context, kpi *models.KPI) (*models.KPI, error) {
	filter := bson.M{
		"period": kpi.Period,
	}
	if kpi.BuildingID != "" {
		filter["building_id"] = kpi.BuildingID
	} else {
		filter["building_id"] = bson.M{"$exists": false}
	}

	update := bson.M{
		"$set": bson.M{
			"calculated_at": kpi.CalculatedAt,
			"metrics":       kpi.Metrics,
			"updated_at":    time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	// Retrieve the updated/created document
	return r.FindLatest(ctx, kpi.BuildingID, kpi.Period)
}
