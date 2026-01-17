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

// ReportRepository handles report database operations
type ReportRepository struct {
	collection *mongo.Collection
}

// NewReportRepository creates a new report repository
func NewReportRepository(collection *mongo.Collection) *ReportRepository {
	return &ReportRepository{collection: collection}
}

// Create inserts a new report
func (r *ReportRepository) Create(ctx context.Context, report *models.Report) (*models.Report, error) {
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, report)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("report with this ID already exists")
		}
		return nil, err
	}

	report.ID = result.InsertedID.(primitive.ObjectID)
	return report, nil
}

// FindByID retrieves a report by its MongoDB ID
func (r *ReportRepository) FindByID(ctx context.Context, id string) (*models.Report, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid report ID format")
	}

	var report models.Report
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&report)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}

	return &report, nil
}

// FindByReportID retrieves a report by its report_id field
func (r *ReportRepository) FindByReportID(ctx context.Context, reportID string) (*models.Report, error) {
	var report models.Report
	err := r.collection.FindOne(ctx, bson.M{"report_id": reportID}).Decode(&report)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}
	return &report, nil
}

// FindAll retrieves reports with filters and pagination
func (r *ReportRepository) FindAll(ctx context.Context, buildingID, reportType, status string, page, limit int) ([]*models.Report, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	skip := int64((page - 1) * limit)
	filter := bson.M{}

	if buildingID != "" {
		filter["building_id"] = buildingID
	}
	if reportType != "" {
		filter["type"] = reportType
	}
	if status != "" {
		filter["status"] = status
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find reports with pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "generated_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reports []*models.Report
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// Update updates a report
func (r *ReportRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Report, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid report ID format")
	}

	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var report models.Report
	if err := result.Decode(&report); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}

	return &report, nil
}
