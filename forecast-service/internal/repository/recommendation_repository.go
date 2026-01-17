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

// RecommendationRepository handles recommendation database operations
type RecommendationRepository struct {
	collection *mongo.Collection
}

// NewRecommendationRepository creates a new recommendation repository
func NewRecommendationRepository(collection *mongo.Collection) *RecommendationRepository {
	return &RecommendationRepository{collection: collection}
}

// Create inserts a new recommendation into the database
func (r *RecommendationRepository) Create(ctx context.Context, rec *models.Recommendation) (*models.Recommendation, error) {
	rec.CreatedAt = time.Now()
	rec.Status = "NEW"

	result, err := r.collection.InsertOne(ctx, rec)
	if err != nil {
		return nil, err
	}

	rec.ID = result.InsertedID.(primitive.ObjectID)
	return rec, nil
}

// CreateMany inserts multiple recommendations
func (r *RecommendationRepository) CreateMany(ctx context.Context, recs []*models.Recommendation) error {
	if len(recs) == 0 {
		return nil
	}

	docs := make([]interface{}, len(recs))
	for i, rec := range recs {
		rec.CreatedAt = time.Now()
		rec.Status = "NEW"
		docs[i] = rec
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

// FindByID retrieves a recommendation by its ID
func (r *RecommendationRepository) FindByID(ctx context.Context, id string) (*models.Recommendation, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid recommendation ID format")
	}

	var rec models.Recommendation
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&rec)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("recommendation not found")
		}
		return nil, err
	}

	return &rec, nil
}

// FindByBuilding retrieves active recommendations for a building
func (r *RecommendationRepository) FindByBuilding(ctx context.Context, buildingID string) ([]*models.Recommendation, error) {
	filter := bson.M{
		"building_id": buildingID,
		"status":      bson.M{"$in": []string{"NEW", "VIEWED"}},
		"$or": []bson.M{
			{"valid_to": nil},
			{"valid_to": bson.M{"$gte": time.Now()}},
		},
	}

	opts := options.Find().SetSort(bson.D{
		{Key: "priority", Value: -1},
		{Key: "created_at", Value: -1},
	})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recs []*models.Recommendation
	if err := cursor.All(ctx, &recs); err != nil {
		return nil, err
	}

	return recs, nil
}

// FindByDevice retrieves recommendations for a specific device
func (r *RecommendationRepository) FindByDevice(ctx context.Context, deviceID string) ([]*models.Recommendation, error) {
	filter := bson.M{
		"device_id": deviceID,
		"status":    bson.M{"$in": []string{"NEW", "VIEWED"}},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recs []*models.Recommendation
	if err := cursor.All(ctx, &recs); err != nil {
		return nil, err
	}

	return recs, nil
}

// UpdateStatus updates the status of a recommendation
func (r *RecommendationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid recommendation ID format")
	}

	updates := bson.M{"status": status}
	now := time.Now()

	switch status {
	case "VIEWED":
		updates["viewed_at"] = now
	case "IMPLEMENTED":
		updates["implemented_at"] = now
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	return err
}

// Delete removes a recommendation from the database
func (r *RecommendationRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid recommendation ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("recommendation not found")
	}

	return nil
}

// DeleteByBuilding removes all recommendations for a building
func (r *RecommendationRepository) DeleteByBuilding(ctx context.Context, buildingID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"building_id": buildingID})
	return err
}

// GetStatsByBuilding returns recommendation statistics for a building
func (r *RecommendationRepository) GetStatsByBuilding(ctx context.Context, buildingID string) (map[string]int, error) {
	recs, err := r.FindByBuilding(ctx, buildingID)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total":    len(recs),
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, rec := range recs {
		switch rec.Priority {
		case models.RecommendationPriorityCritical:
			stats["critical"]++
		case models.RecommendationPriorityHigh:
			stats["high"]++
		case models.RecommendationPriorityMedium:
			stats["medium"]++
		case models.RecommendationPriorityLow:
			stats["low"]++
		}
	}

	return stats, nil
}
