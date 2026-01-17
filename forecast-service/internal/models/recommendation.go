package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeImmediate  RecommendationType = "IMMEDIATE"
	RecommendationTypeScheduled  RecommendationType = "SCHEDULED"
	RecommendationTypeLongTerm   RecommendationType = "LONG_TERM"
	RecommendationTypeBehavioral RecommendationType = "BEHAVIORAL"
)

// RecommendationPriority represents the priority level
type RecommendationPriority string

const (
	RecommendationPriorityLow    RecommendationPriority = "LOW"
	RecommendationPriorityMedium RecommendationPriority = "MEDIUM"
	RecommendationPriorityHigh   RecommendationPriority = "HIGH"
	RecommendationPriorityCritical RecommendationPriority = "CRITICAL"
)

// Recommendation represents an energy-saving recommendation
type Recommendation struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	BuildingID      string                 `bson:"building_id" json:"buildingId"`
	DeviceID        string                 `bson:"device_id,omitempty" json:"deviceId,omitempty"`
	Type            RecommendationType     `bson:"type" json:"type"`
	Priority        RecommendationPriority `bson:"priority" json:"priority"`
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description" json:"description"`
	ActionRequired  string                 `bson:"action_required" json:"actionRequired"`
	ExpectedSavings Savings                `bson:"expected_savings" json:"expectedSavings"`
	ImplementationSteps []string           `bson:"implementation_steps" json:"implementationSteps"`
	AutomationAvailable bool               `bson:"automation_available" json:"automationAvailable"`
	Status          string                 `bson:"status" json:"status"` // NEW, VIEWED, IMPLEMENTED, DISMISSED
	Category        string                 `bson:"category" json:"category"` // HVAC, LIGHTING, EQUIPMENT, etc.
	ValidFrom       time.Time              `bson:"valid_from" json:"validFrom"`
	ValidTo         *time.Time             `bson:"valid_to,omitempty" json:"validTo,omitempty"`
	CreatedAt       time.Time              `bson:"created_at" json:"createdAt"`
	ViewedAt        *time.Time             `bson:"viewed_at,omitempty" json:"viewedAt,omitempty"`
	ImplementedAt   *time.Time             `bson:"implemented_at,omitempty" json:"implementedAt,omitempty"`
}

// RecommendationsResponse represents the recommendations for a building
type RecommendationsResponse struct {
	BuildingID        string            `json:"buildingId"`
	TotalRecommendations int            `json:"totalRecommendations"`
	TotalPotentialSavings Savings       `json:"totalPotentialSavings"`
	ByPriority        PrioritySummary   `json:"byPriority"`
	ByCategory        map[string]int    `json:"byCategory"`
	Recommendations   []RecommendationItem `json:"recommendations"`
	GeneratedAt       time.Time         `json:"generatedAt"`
}

// PrioritySummary summarizes recommendations by priority
type PrioritySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// RecommendationItem represents a single recommendation in the list
type RecommendationItem struct {
	ID                  string                 `json:"id"`
	Type                RecommendationType     `json:"type"`
	Priority            RecommendationPriority `json:"priority"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	ActionRequired      string                 `json:"actionRequired"`
	ExpectedSavings     Savings                `json:"expectedSavings"`
	Category            string                 `json:"category"`
	DeviceID            string                 `json:"deviceId,omitempty"`
	AutomationAvailable bool                   `json:"automationAvailable"`
	ImplementationSteps []string               `json:"implementationSteps"`
}

// ToRecommendationItem converts a Recommendation to RecommendationItem
func (r *Recommendation) ToRecommendationItem() RecommendationItem {
	return RecommendationItem{
		ID:                  r.ID.Hex(),
		Type:                r.Type,
		Priority:            r.Priority,
		Title:               r.Title,
		Description:         r.Description,
		ActionRequired:      r.ActionRequired,
		ExpectedSavings:     r.ExpectedSavings,
		Category:            r.Category,
		DeviceID:            r.DeviceID,
		AutomationAvailable: r.AutomationAvailable,
		ImplementationSteps: r.ImplementationSteps,
	}
}
