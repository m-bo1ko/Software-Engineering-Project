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

// OptimizationRepository handles optimization scenario database operations
type OptimizationRepository struct {
	collection *mongo.Collection
}

// NewOptimizationRepository creates a new optimization repository
func NewOptimizationRepository(collection *mongo.Collection) *OptimizationRepository {
	return &OptimizationRepository{collection: collection}
}

// Create inserts a new optimization scenario
func (r *OptimizationRepository) Create(ctx context.Context, scenario *models.OptimizationScenario) (*models.OptimizationScenario, error) {
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, scenario)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("scenario with this ID already exists")
		}
		return nil, err
	}

	scenario.ID = result.InsertedID.(primitive.ObjectID)
	return scenario, nil
}

// FindByID retrieves a scenario by its MongoDB ID
func (r *OptimizationRepository) FindByID(ctx context.Context, id string) (*models.OptimizationScenario, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid scenario ID format")
	}

	var scenario models.OptimizationScenario
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&scenario)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("scenario not found")
		}
		return nil, err
	}

	return &scenario, nil
}

// FindByScenarioID retrieves a scenario by its scenario_id field
func (r *OptimizationRepository) FindByScenarioID(ctx context.Context, scenarioID string) (*models.OptimizationScenario, error) {
	var scenario models.OptimizationScenario
	err := r.collection.FindOne(ctx, bson.M{"scenario_id": scenarioID}).Decode(&scenario)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("scenario not found")
		}
		return nil, err
	}
	return &scenario, nil
}

// Update updates a scenario
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
			return nil, errors.New("scenario not found")
		}
		return nil, err
	}

	return &scenario, nil
}

// UpdateProgress updates the progress and status of a scenario
func (r *OptimizationRepository) UpdateProgress(ctx context.Context, scenarioID string, progress float64, status models.OptimizationExecutionStatus) error {
	updates := bson.M{
		"progress":     progress,
		"execution_status": status,
		"updated_at":   time.Now(),
	}

	if status == models.OptimizationStatusRunning && progress == 0 {
		now := time.Now()
		updates["started_at"] = now
	}

	if status == models.OptimizationStatusCompleted || status == models.OptimizationStatusFailed {
		now := time.Now()
		updates["completed_at"] = now
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"scenario_id": scenarioID},
		bson.M{"$set": updates},
	)
	return err
}

// UpdateActionStatus updates the status of a specific action in a scenario
func (r *OptimizationRepository) UpdateActionStatus(ctx context.Context, scenarioID string, deviceID string, status string, commandID string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"scenario_id": scenarioID,
			"actions.device_id": deviceID,
		},
		bson.M{
			"$set": bson.M{
				"actions.$.status":     status,
				"actions.$.command_id": commandID,
				"updated_at":           time.Now(),
			},
		},
	)
	return err
}
