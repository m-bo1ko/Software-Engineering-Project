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

// OptimizationRepository handles optimization scenario database operations
type OptimizationRepository struct {
	collection *mongo.Collection
}

// NewOptimizationRepository creates a new optimization repository
func NewOptimizationRepository(collection *mongo.Collection) *OptimizationRepository {
	return &OptimizationRepository{collection: collection}
}

// Create inserts a new optimization scenario into the database
func (r *OptimizationRepository) Create(ctx context.Context, scenario *models.OptimizationScenario) (*models.OptimizationScenario, error) {
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, scenario)
	if err != nil {
		return nil, err
	}

	scenario.ID = result.InsertedID.(primitive.ObjectID)
	return scenario, nil
}

// FindByID retrieves an optimization scenario by its ID
func (r *OptimizationRepository) FindByID(ctx context.Context, id string) (*models.OptimizationScenario, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid scenario ID format")
	}

	var scenario models.OptimizationScenario
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&scenario)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("optimization scenario not found")
		}
		return nil, err
	}

	return &scenario, nil
}

// FindByBuilding retrieves optimization scenarios for a building
func (r *OptimizationRepository) FindByBuilding(ctx context.Context, buildingID string, status models.OptimizationStatus, page, limit int) ([]*models.OptimizationScenario, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{"building_id": buildingID}

	if status != "" {
		filter["status"] = status
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find scenarios with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var scenarios []*models.OptimizationScenario
	if err := cursor.All(ctx, &scenarios); err != nil {
		return nil, 0, err
	}

	return scenarios, total, nil
}

// FindPendingScenarios retrieves scenarios ready for execution
func (r *OptimizationRepository) FindPendingScenarios(ctx context.Context) ([]*models.OptimizationScenario, error) {
	filter := bson.M{
		"status": models.OptimizationStatusApproved,
		"scheduled_start": bson.M{"$lte": time.Now()},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var scenarios []*models.OptimizationScenario
	if err := cursor.All(ctx, &scenarios); err != nil {
		return nil, err
	}

	return scenarios, nil
}

// Update updates an existing optimization scenario
func (r *OptimizationRepository) Update(ctx context.Context, id string, updates bson.M) (*models.OptimizationScenario, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid scenario ID format")
	}

	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var scenario models.OptimizationScenario
	if err := result.Decode(&scenario); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("optimization scenario not found")
		}
		return nil, err
	}

	return &scenario, nil
}

// UpdateStatus updates the status of an optimization scenario
func (r *OptimizationRepository) UpdateStatus(ctx context.Context, id string, status models.OptimizationStatus, errorMsg string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid scenario ID format")
	}

	updates := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	return err
}

// ApproveScenario approves a scenario for execution
func (r *OptimizationRepository) ApproveScenario(ctx context.Context, id, approverID string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid scenario ID format")
	}

	now := time.Now()
	updates := bson.M{
		"status":      models.OptimizationStatusApproved,
		"approved_by": approverID,
		"approved_at": now,
		"updated_at":  now,
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	return err
}

// AddExecutionLog adds a log entry to the scenario
func (r *OptimizationRepository) AddExecutionLog(ctx context.Context, id string, entry models.ExecutionLogEntry) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid scenario ID format")
	}

	entry.Timestamp = time.Now()

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$push": bson.M{"execution_log": entry},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// UpdateActionStatus updates the status of a specific action
func (r *OptimizationRepository) UpdateActionStatus(ctx context.Context, scenarioID, actionID, status string, actualImpact *float64, errorMsg string) error {
	objectID, err := primitive.ObjectIDFromHex(scenarioID)
	if err != nil {
		return errors.New("invalid scenario ID format")
	}

	now := time.Now()
	update := bson.M{
		"actions.$.status":      status,
		"actions.$.executed_at": now,
		"updated_at":            now,
	}

	if actualImpact != nil {
		update["actions.$.actual_impact"] = *actualImpact
	}

	if errorMsg != "" {
		update["actions.$.error_message"] = errorMsg
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "actions.id": actionID},
		bson.M{"$set": update},
	)
	return err
}

// Delete removes an optimization scenario from the database
func (r *OptimizationRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid scenario ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("optimization scenario not found")
	}

	return nil
}

// CountByStatus counts scenarios by status
func (r *OptimizationRepository) CountByStatus(ctx context.Context, buildingID string, status models.OptimizationStatus) (int64, error) {
	filter := bson.M{"status": status}
	if buildingID != "" {
		filter["building_id"] = buildingID
	}
	return r.collection.CountDocuments(ctx, filter)
}
