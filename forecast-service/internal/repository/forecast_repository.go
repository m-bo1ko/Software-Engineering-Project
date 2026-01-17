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

// ForecastRepository handles forecast database operations
type ForecastRepository struct {
	collection *mongo.Collection
}

// NewForecastRepository creates a new forecast repository
func NewForecastRepository(collection *mongo.Collection) *ForecastRepository {
	return &ForecastRepository{collection: collection}
}

// Create inserts a new forecast into the database
func (r *ForecastRepository) Create(ctx context.Context, forecast *models.Forecast) (*models.Forecast, error) {
	forecast.CreatedAt = time.Now()
	forecast.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, forecast)
	if err != nil {
		return nil, err
	}

	forecast.ID = result.InsertedID.(primitive.ObjectID)
	return forecast, nil
}

// FindByID retrieves a forecast by its ID
func (r *ForecastRepository) FindByID(ctx context.Context, id string) (*models.Forecast, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid forecast ID format")
	}

	var forecast models.Forecast
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&forecast)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("forecast not found")
		}
		return nil, err
	}

	return &forecast, nil
}

// FindLatestByBuilding retrieves the latest forecast for a building
func (r *ForecastRepository) FindLatestByBuilding(ctx context.Context, buildingID string, forecastType models.ForecastType) (*models.Forecast, error) {
	filter := bson.M{
		"building_id": buildingID,
		"status":      models.ForecastStatusCompleted,
	}

	if forecastType != "" {
		filter["type"] = forecastType
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var forecast models.Forecast
	err := r.collection.FindOne(ctx, filter, opts).Decode(&forecast)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("no forecasts found for this building")
		}
		return nil, err
	}

	return &forecast, nil
}

// FindByBuilding retrieves forecasts for a building with pagination
func (r *ForecastRepository) FindByBuilding(ctx context.Context, buildingID string, page, limit int) ([]*models.Forecast, int64, error) {
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

	// Find forecasts with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var forecasts []*models.Forecast
	if err := cursor.All(ctx, &forecasts); err != nil {
		return nil, 0, err
	}

	return forecasts, total, nil
}

// FindByDevice retrieves forecasts for a specific device
func (r *ForecastRepository) FindByDevice(ctx context.Context, deviceID string) ([]*models.Forecast, error) {
	filter := bson.M{"device_id": deviceID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(10)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var forecasts []*models.Forecast
	if err := cursor.All(ctx, &forecasts); err != nil {
		return nil, err
	}

	return forecasts, nil
}

// Update updates an existing forecast
func (r *ForecastRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Forecast, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid forecast ID format")
	}

	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var forecast models.Forecast
	if err := result.Decode(&forecast); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("forecast not found")
		}
		return nil, err
	}

	return &forecast, nil
}

// UpdateStatus updates the status of a forecast
func (r *ForecastRepository) UpdateStatus(ctx context.Context, id string, status models.ForecastStatus, errorMsg string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid forecast ID format")
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

// UpdatePredictions updates the predictions for a forecast
func (r *ForecastRepository) UpdatePredictions(ctx context.Context, id string, predictions []models.ForecastPrediction, accuracy *models.ForecastAccuracy) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid forecast ID format")
	}

	updates := bson.M{
		"predictions": predictions,
		"status":      models.ForecastStatusCompleted,
		"updated_at":  time.Now(),
	}

	if accuracy != nil {
		updates["accuracy"] = accuracy
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	return err
}

// Delete removes a forecast from the database
func (r *ForecastRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid forecast ID format")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("forecast not found")
	}

	return nil
}

// CountByStatus counts forecasts by status
func (r *ForecastRepository) CountByStatus(ctx context.Context, status models.ForecastStatus) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"status": status})
}
