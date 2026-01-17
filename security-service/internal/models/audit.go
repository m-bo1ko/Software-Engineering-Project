package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID      string                 `bson:"user_id" json:"userId"`
	Username    string                 `bson:"username" json:"username"`
	Service     string                 `bson:"service" json:"service"`     // e.g., "auth-service", "building-service"
	Action      string                 `bson:"action" json:"action"`       // e.g., "LOGIN", "CREATE_USER", "DELETE_BUILDING"
	Resource    string                 `bson:"resource" json:"resource"`   // e.g., "user", "building", "report"
	ResourceID  string                 `bson:"resource_id" json:"resourceId,omitempty"`
	Details     map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
	IPAddress   string                 `bson:"ip_address" json:"ipAddress"`
	UserAgent   string                 `bson:"user_agent" json:"userAgent"`
	Status      string                 `bson:"status" json:"status"` // "SUCCESS", "FAILURE"
	ErrorMsg    string                 `bson:"error_msg,omitempty" json:"errorMsg,omitempty"`
	Timestamp   time.Time              `bson:"timestamp" json:"timestamp"`
	RequestPath string                 `bson:"request_path" json:"requestPath"`
	Method      string                 `bson:"method" json:"method"` // HTTP method
}

// AuditLogCreateRequest represents the request body for creating an audit log
type AuditLogCreateRequest struct {
	UserID      string                 `json:"userId"`
	Username    string                 `json:"username"`
	Service     string                 `json:"service" binding:"required"`
	Action      string                 `json:"action" binding:"required"`
	Resource    string                 `json:"resource" binding:"required"`
	ResourceID  string                 `json:"resourceId"`
	Details     map[string]interface{} `json:"details"`
	IPAddress   string                 `json:"ipAddress"`
	UserAgent   string                 `json:"userAgent"`
	Status      string                 `json:"status" binding:"required,oneof=SUCCESS FAILURE"`
	ErrorMsg    string                 `json:"errorMsg"`
	RequestPath string                 `json:"requestPath"`
	Method      string                 `json:"method"`
}

// AuditLogQueryParams represents the query parameters for filtering audit logs
type AuditLogQueryParams struct {
	From     time.Time `form:"from"`
	To       time.Time `form:"to"`
	UserID   string    `form:"userId"`
	Service  string    `form:"service"`
	Action   string    `form:"action"`
	Resource string    `form:"resource"`
	Status   string    `form:"status"`
	Page     int       `form:"page"`
	Limit    int       `form:"limit"`
}

// AuditLogResponse represents the audit log data returned in API responses
type AuditLogResponse struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"userId"`
	Username    string                 `json:"username"`
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resourceId,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IPAddress   string                 `json:"ipAddress"`
	UserAgent   string                 `json:"userAgent"`
	Status      string                 `json:"status"`
	ErrorMsg    string                 `json:"errorMsg,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestPath string                 `json:"requestPath"`
	Method      string                 `json:"method"`
}

// ToResponse converts an AuditLog to AuditLogResponse
func (a *AuditLog) ToResponse() *AuditLogResponse {
	return &AuditLogResponse{
		ID:          a.ID.Hex(),
		UserID:      a.UserID,
		Username:    a.Username,
		Service:     a.Service,
		Action:      a.Action,
		Resource:    a.Resource,
		ResourceID:  a.ResourceID,
		Details:     a.Details,
		IPAddress:   a.IPAddress,
		UserAgent:   a.UserAgent,
		Status:      a.Status,
		ErrorMsg:    a.ErrorMsg,
		Timestamp:   a.Timestamp,
		RequestPath: a.RequestPath,
		Method:      a.Method,
	}
}

// PaginatedAuditLogsResponse represents a paginated list of audit logs
type PaginatedAuditLogsResponse struct {
	Logs       []*AuditLogResponse `json:"logs"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"totalPages"`
}
