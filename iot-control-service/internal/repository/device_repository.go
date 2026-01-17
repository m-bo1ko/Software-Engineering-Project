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

// DeviceRepository handles device database operations
type DeviceRepository struct {
	collection *mongo.Collection
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(collection *mongo.Collection) *DeviceRepository {
	return &DeviceRepository{collection: collection}
}

// Create inserts a new device into the database
func (r *DeviceRepository) Create(ctx context.Context, device *models.Device) (*models.Device, error) {
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, device)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("device with this ID already exists")
		}
		return nil, err
	}

	device.ID = result.InsertedID.(primitive.ObjectID)
	return device, nil
}

// FindByID retrieves a device by its MongoDB ID
func (r *DeviceRepository) FindByID(ctx context.Context, id string) (*models.Device, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid device ID format")
	}

	var device models.Device
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&device)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}

	return &device, nil
}

// FindByDeviceID retrieves a device by its device_id field
func (r *DeviceRepository) FindByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	var device models.Device
	err := r.collection.FindOne(ctx, bson.M{"device_id": deviceID}).Decode(&device)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}
	return &device, nil
}

// FindAll retrieves devices with filters and pagination
func (r *DeviceRepository) FindAll(ctx context.Context, buildingID, deviceType, status string, page, limit int) ([]*models.Device, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{}

	if buildingID != "" {
		filter["location.building_id"] = buildingID
	}
	if deviceType != "" {
		filter["type"] = deviceType
	}
	if status != "" {
		filter["status"] = status
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find devices with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var devices []*models.Device
	if err := cursor.All(ctx, &devices); err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// Update updates an existing device
func (r *DeviceRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Device, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid device ID format")
	}

	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var device models.Device
	if err := result.Decode(&device); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}

	return &device, nil
}

// UpdateLastSeen updates the last_seen timestamp for a device
func (r *DeviceRepository) UpdateLastSeen(ctx context.Context, deviceID string) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"device_id": deviceID},
		bson.M{
			"$set": bson.M{
				"last_seen": now,
				"updated_at": now,
				"status": models.DeviceStatusOnline,
			},
		},
	)
	return err
}

// UpdateStatus updates the status of a device
func (r *DeviceRepository) UpdateStatus(ctx context.Context, deviceID string, status models.DeviceStatus) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"device_id": deviceID},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

// Delete removes a device from the database
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid device ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("device not found")
	}

	return nil
}
