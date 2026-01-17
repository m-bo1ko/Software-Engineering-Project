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

// CommandRepository handles device command database operations
type CommandRepository struct {
	collection *mongo.Collection
}

// NewCommandRepository creates a new command repository
func NewCommandRepository(collection *mongo.Collection) *CommandRepository {
	return &CommandRepository{collection: collection}
}

// Create inserts a new command
func (r *CommandRepository) Create(ctx context.Context, command *models.DeviceCommand) (*models.DeviceCommand, error) {
	command.CreatedAt = time.Now()
	command.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, command)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("command with this ID already exists")
		}
		return nil, err
	}

	command.ID = result.InsertedID.(primitive.ObjectID)
	return command, nil
}

// FindByID retrieves a command by its MongoDB ID
func (r *CommandRepository) FindByID(ctx context.Context, id string) (*models.DeviceCommand, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid command ID format")
	}

	var command models.DeviceCommand
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&command)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("command not found")
		}
		return nil, err
	}

	return &command, nil
}

// FindByCommandID retrieves a command by its command_id field
func (r *CommandRepository) FindByCommandID(ctx context.Context, commandID string) (*models.DeviceCommand, error) {
	var command models.DeviceCommand
	err := r.collection.FindOne(ctx, bson.M{"command_id": commandID}).Decode(&command)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("command not found")
		}
		return nil, err
	}
	return &command, nil
}

// FindByDeviceID retrieves commands for a device
func (r *CommandRepository) FindByDeviceID(ctx context.Context, deviceID string, status string, page, limit int) ([]*models.DeviceCommand, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{"device_id": deviceID}

	if status != "" {
		filter["status"] = status
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find commands with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var commands []*models.DeviceCommand
	if err := cursor.All(ctx, &commands); err != nil {
		return nil, 0, err
	}

	return commands, total, nil
}

// Update updates a command
func (r *CommandRepository) Update(ctx context.Context, id string, updates bson.M) (*models.DeviceCommand, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid command ID format")
	}

	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var command models.DeviceCommand
	if err := result.Decode(&command); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("command not found")
		}
		return nil, err
	}

	return &command, nil
}

// UpdateStatus updates the status of a command
func (r *CommandRepository) UpdateStatus(ctx context.Context, commandID string, status models.CommandStatus, errorMsg string) error {
	updates := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == models.CommandStatusSent {
		now := time.Now()
		updates["sent_at"] = now
	}

	if status == models.CommandStatusApplied {
		now := time.Now()
		updates["applied_at"] = now
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"command_id": commandID},
		bson.M{"$set": updates},
	)
	return err
}
