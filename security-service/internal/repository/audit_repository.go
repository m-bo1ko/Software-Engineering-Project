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

// AuditRepository handles audit log database operations
type AuditRepository struct {
	collection *mongo.Collection
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(collection *mongo.Collection) *AuditRepository {
	return &AuditRepository{collection: collection}
}

// Create inserts a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	log.Timestamp = time.Now()

	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return nil, err
	}

	log.ID = result.InsertedID.(primitive.ObjectID)
	return log, nil
}

// FindByID retrieves an audit log by its ID
func (r *AuditRepository) FindByID(ctx context.Context, id string) (*models.AuditLog, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid audit log ID format")
	}

	var log models.AuditLog
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&log)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("audit log not found")
		}
		return nil, err
	}

	return &log, nil
}

// Find retrieves audit logs with filters and pagination
func (r *AuditRepository) Find(ctx context.Context, params models.AuditLogQueryParams) ([]*models.AuditLog, int64, error) {
	// Build filter
	filter := bson.M{}

	// Time range filter
	if !params.From.IsZero() || !params.To.IsZero() {
		timeFilter := bson.M{}
		if !params.From.IsZero() {
			timeFilter["$gte"] = params.From
		}
		if !params.To.IsZero() {
			timeFilter["$lte"] = params.To
		}
		filter["timestamp"] = timeFilter
	}

	// User ID filter
	if params.UserID != "" {
		filter["user_id"] = params.UserID
	}

	// Service filter
	if params.Service != "" {
		filter["service"] = params.Service
	}

	// Action filter
	if params.Action != "" {
		filter["action"] = params.Action
	}

	// Resource filter
	if params.Resource != "" {
		filter["resource"] = params.Resource
	}

	// Status filter
	if params.Status != "" {
		filter["status"] = params.Status
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
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find logs with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []*models.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByUser retrieves audit logs for a specific user
func (r *AuditRepository) FindByUser(ctx context.Context, userID string, page, limit int) ([]*models.AuditLog, int64, error) {
	return r.Find(ctx, models.AuditLogQueryParams{
		UserID: userID,
		Page:   page,
		Limit:  limit,
	})
}

// FindByService retrieves audit logs for a specific service
func (r *AuditRepository) FindByService(ctx context.Context, service string, page, limit int) ([]*models.AuditLog, int64, error) {
	return r.Find(ctx, models.AuditLogQueryParams{
		Service: service,
		Page:    page,
		Limit:   limit,
	})
}

// FindByDateRange retrieves audit logs within a date range
func (r *AuditRepository) FindByDateRange(ctx context.Context, from, to time.Time, page, limit int) ([]*models.AuditLog, int64, error) {
	return r.Find(ctx, models.AuditLogQueryParams{
		From:  from,
		To:    to,
		Page:  page,
		Limit: limit,
	})
}

// GetPaginatedResponse returns a paginated audit logs response
func (r *AuditRepository) GetPaginatedResponse(ctx context.Context, params models.AuditLogQueryParams) (*models.PaginatedAuditLogsResponse, error) {
	logs, total, err := r.Find(ctx, params)
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
	logResponses := make([]*models.AuditLogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &models.PaginatedAuditLogsResponse{
		Logs:       logResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// DeleteOlderThan removes audit logs older than the specified duration
func (r *AuditRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{
		"timestamp": bson.M{"$lt": before},
	})

	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// CountByAction counts audit logs by action type
func (r *AuditRepository) CountByAction(ctx context.Context, action string, from, to time.Time) (int64, error) {
	filter := bson.M{"action": action}

	if !from.IsZero() || !to.IsZero() {
		timeFilter := bson.M{}
		if !from.IsZero() {
			timeFilter["$gte"] = from
		}
		if !to.IsZero() {
			timeFilter["$lte"] = to
		}
		filter["timestamp"] = timeFilter
	}

	return r.collection.CountDocuments(ctx, filter)
}
