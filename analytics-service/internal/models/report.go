package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "PENDING"
	ReportStatusGenerating ReportStatus = "GENERATING"
	ReportStatusCompleted ReportStatus = "COMPLETED"
	ReportStatusFailed    ReportStatus = "FAILED"
)

// Report represents an analytical report
type Report struct {
	ID          primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	ReportID    string                      `bson:"report_id" json:"reportId"`
	BuildingID string                      `bson:"building_id,omitempty" json:"buildingId,omitempty"`
	Type        string                      `bson:"type" json:"type"`
	Status      ReportStatus                `bson:"status" json:"status"`
	Content     map[string]interface{}      `bson:"content" json:"content"`
	GeneratedAt time.Time                   `bson:"generated_at" json:"generatedAt"`
	GeneratedBy string                      `bson:"generated_by" json:"generatedBy"`
	CreatedAt   time.Time                   `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time                   `bson:"updated_at" json:"updatedAt"`
}

// ReportResponse represents report data in API responses
type ReportResponse struct {
	ID          string                 `json:"id"`
	ReportID    string                 `json:"reportId"`
	BuildingID  string                 `json:"buildingId,omitempty"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Content     map[string]interface{} `json:"content"`
	GeneratedAt time.Time              `json:"generatedAt"`
	GeneratedBy string                 `json:"generatedBy"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// ToResponse converts a Report to ReportResponse
func (r *Report) ToResponse() *ReportResponse {
	return &ReportResponse{
		ID:          r.ID.Hex(),
		ReportID:    r.ReportID,
		BuildingID:  r.BuildingID,
		Type:        r.Type,
		Status:      string(r.Status),
		Content:     r.Content,
		GeneratedAt: r.GeneratedAt,
		GeneratedBy: r.GeneratedBy,
		CreatedAt:   r.CreatedAt,
	}
}

// GenerateReportRequest represents a request to generate a report
type GenerateReportRequest struct {
	BuildingID string    `json:"buildingId,omitempty"`
	Type       string    `json:"type" binding:"required"`
	From       time.Time `json:"from,omitempty"`
	To         time.Time `json:"to,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// ListReportsRequest represents query parameters for listing reports
type ListReportsRequest struct {
	BuildingID string `form:"buildingId"`
	Type       string `form:"type"`
	Status     string `form:"status"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}
