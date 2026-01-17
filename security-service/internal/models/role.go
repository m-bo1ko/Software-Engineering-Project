package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Permission represents a specific permission for a resource action
type Permission struct {
	Resource string   `bson:"resource" json:"resource"` // e.g., "users", "buildings", "reports"
	Actions  []string `bson:"actions" json:"actions"`   // e.g., ["read", "write", "delete"]
}

// Role represents a role with its associated permissions
type Role struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" binding:"required"`
	Description string             `bson:"description" json:"description"`
	Permissions []Permission       `bson:"permissions" json:"permissions"`
	IsSystem    bool               `bson:"is_system" json:"isSystem"` // System roles cannot be deleted
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// RoleCreateRequest represents the request body for creating a new role
type RoleCreateRequest struct {
	Name        string       `json:"name" binding:"required,min=2,max=50"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// RoleUpdateRequest represents the request body for updating a role
type RoleUpdateRequest struct {
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// RoleResponse represents the role data returned in API responses
type RoleResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	IsSystem    bool         `json:"isSystem"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// ToResponse converts a Role to RoleResponse
func (r *Role) ToResponse() *RoleResponse {
	return &RoleResponse{
		ID:          r.ID.Hex(),
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// HasPermission checks if the role has a specific permission for a resource action
func (r *Role) HasPermission(resource, action string) bool {
	for _, perm := range r.Permissions {
		if perm.Resource == resource || perm.Resource == "*" {
			for _, a := range perm.Actions {
				if a == action || a == "*" {
					return true
				}
			}
		}
	}
	return false
}
