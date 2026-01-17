package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"security-service/internal/models"
)

// NotificationRepository handles notification database operations
type NotificationRepository struct {
	notifications *mongo.Collection
	preferences   *mongo.Collection
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(notifications, preferences *mongo.Collection) *NotificationRepository {
	return &NotificationRepository{
		notifications: notifications,
		preferences:   preferences,
	}
}

// Create inserts a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	notification.CreatedAt = time.Now()
	notification.Status = models.NotificationStatusPending

	result, err := r.notifications.InsertOne(ctx, notification)
	if err != nil {
		return nil, err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)
	return notification, nil
}

// FindByID retrieves a notification by its ID
func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*models.Notification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid notification ID format")
	}

	var notification models.Notification
	err = r.notifications.FindOne(ctx, bson.M{"_id": objectID}).Decode(&notification)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}

	return &notification, nil
}

// FindByUser retrieves notifications for a specific user with filters
func (r *NotificationRepository) FindByUser(ctx context.Context, params models.NotificationLogQueryParams) ([]*models.Notification, int64, error) {
	filter := bson.M{"user_id": params.UserID}

	// Type filter
	if params.Type != "" {
		filter["type"] = params.Type
	}

	// Status filter
	if params.Status != "" {
		filter["status"] = params.Status
	}

	// Time range filter
	if !params.From.IsZero() || !params.To.IsZero() {
		timeFilter := bson.M{}
		if !params.From.IsZero() {
			timeFilter["$gte"] = params.From
		}
		if !params.To.IsZero() {
			timeFilter["$lte"] = params.To
		}
		filter["created_at"] = timeFilter
	}

	// Set default pagination
	page := params.Page
	limit := params.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)

	// Get total count
	total, err := r.notifications.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find notifications with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.notifications.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// UpdateStatus updates the status of a notification
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid notification ID format")
	}

	updates := bson.M{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	now := time.Now()
	if status == models.NotificationStatusSent {
		updates["sent_at"] = now
	} else if status == models.NotificationStatusDelivered {
		updates["delivered_at"] = now
	}

	_, err = r.notifications.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	return err
}

// GetPreferences retrieves notification preferences for a user
func (r *NotificationRepository) GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error) {
	var prefs models.NotificationPreferences
	err := r.preferences.FindOne(ctx, bson.M{"user_id": userID}).Decode(&prefs)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Return default preferences
			return &models.NotificationPreferences{
				UserID:       userID,
				EmailEnabled: true,
				SMSEnabled:   false,
				PushEnabled:  true,
			}, nil
		}
		return nil, err
	}

	return &prefs, nil
}

// SavePreferences saves or updates notification preferences for a user
func (r *NotificationRepository) SavePreferences(ctx context.Context, prefs *models.NotificationPreferences) error {
	prefs.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	_, err := r.preferences.UpdateOne(
		ctx,
		bson.M{"user_id": prefs.UserID},
		bson.M{"$set": prefs},
		opts,
	)

	return err
}

// GetPaginatedResponse returns a paginated notifications response
func (r *NotificationRepository) GetPaginatedResponse(ctx context.Context, params models.NotificationLogQueryParams) (*models.PaginatedNotificationsResponse, error) {
	notifications, total, err := r.FindByUser(ctx, params)
	if err != nil {
		return nil, err
	}

	page := params.Page
	limit := params.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Convert to response format
	notificationResponses := make([]*models.NotificationResponse, len(notifications))
	for i, n := range notifications {
		notificationResponses[i] = n.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &models.PaginatedNotificationsResponse{
		Notifications: notificationResponses,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}, nil
}

// CountPendingNotifications counts notifications with pending status
func (r *NotificationRepository) CountPendingNotifications(ctx context.Context) (int64, error) {
	return r.notifications.CountDocuments(ctx, bson.M{"status": models.NotificationStatusPending})
}

// DeleteOlderThan removes notifications older than the specified duration
func (r *NotificationRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.notifications.DeleteMany(ctx, bson.M{
		"created_at": bson.M{"$lt": before},
	})

	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
