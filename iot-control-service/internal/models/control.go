package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommandStatus represents the status of a command
type CommandStatus string

const (
	CommandStatusPending   CommandStatus = "PENDING"
	CommandStatusSent      CommandStatus = "SENT"
	CommandStatusApplied   CommandStatus = "APPLIED"
	CommandStatusFailed    CommandStatus = "FAILED"
	CommandStatusCancelled CommandStatus = "CANCELLED"
	CommandStatusTimeout   CommandStatus = "TIMEOUT"
)

// DeviceCommand represents a command sent to a device
type DeviceCommand struct {
	ID          primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	CommandID   string                      `bson:"command_id" json:"commandId"`
	DeviceID    string                      `bson:"device_id" json:"deviceId"`
	Command     string                      `bson:"command" json:"command"`
	Params      map[string]interface{}      `bson:"params" json:"params"`
	Status      CommandStatus               `bson:"status" json:"status"`
	IssuedBy    string                      `bson:"issued_by" json:"issuedBy"`
	ErrorMsg    string                      `bson:"error_msg,omitempty" json:"errorMsg,omitempty"`
	SentAt      *time.Time                  `bson:"sent_at,omitempty" json:"sentAt,omitempty"`
	AppliedAt   *time.Time                  `bson:"applied_at,omitempty" json:"appliedAt,omitempty"`
	CreatedAt   time.Time                   `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time                   `bson:"updated_at" json:"updatedAt"`
}

// CommandResponse represents command data in API responses
type CommandResponse struct {
	ID        string                 `json:"id"`
	CommandID string                 `json:"commandId"`
	DeviceID  string                 `json:"deviceId"`
	Command   string                 `json:"command"`
	Params    map[string]interface{} `json:"params"`
	Status    string                 `json:"status"`
	IssuedBy  string                 `json:"issuedBy"`
	ErrorMsg  string                 `json:"errorMsg,omitempty"`
	SentAt    *time.Time             `json:"sentAt,omitempty"`
	AppliedAt *time.Time             `json:"appliedAt,omitempty"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

// ToResponse converts a DeviceCommand to CommandResponse
func (c *DeviceCommand) ToResponse() *CommandResponse {
	return &CommandResponse{
		ID:        c.ID.Hex(),
		CommandID: c.CommandID,
		DeviceID:  c.DeviceID,
		Command:   c.Command,
		Params:    c.Params,
		Status:    string(c.Status),
		IssuedBy:  c.IssuedBy,
		ErrorMsg:  c.ErrorMsg,
		SentAt:    c.SentAt,
		AppliedAt: c.AppliedAt,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// SendCommandRequest represents a request to send a command
type SendCommandRequest struct {
	Command string                 `json:"command" binding:"required"`
	Params  map[string]interface{} `json:"params"`
}

// ListCommandsRequest represents query parameters for listing commands
type ListCommandsRequest struct {
	Status   string `form:"status"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}

// CommandAck represents an acknowledgment from a device
type CommandAck struct {
	CommandID string    `json:"commandId"`
	DeviceID  string    `json:"deviceId"`
	Status    string    `json:"status"` // "APPLIED" or "FAILED"
	ErrorMsg  string    `json:"errorMsg,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
