package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "PENDING"
	NotificationStatusSent      NotificationStatus = "SENT"
	NotificationStatusFailed    NotificationStatus = "FAILED"
	NotificationStatusDelivered NotificationStatus = "DELIVERED"
)

// Notification represents a notification record
type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"userId"`
	Type        NotificationType   `bson:"type" json:"type"`
	Subject     string             `bson:"subject" json:"subject"`
	Content     string             `bson:"content" json:"content"`
	Recipient   string             `bson:"recipient" json:"recipient"` // email address, phone number, or device token
	Status      NotificationStatus `bson:"status" json:"status"`
	ErrorMsg    string             `bson:"error_msg,omitempty" json:"errorMsg,omitempty"`
	Metadata    map[string]string  `bson:"metadata,omitempty" json:"metadata,omitempty"`
	SentAt      *time.Time         `bson:"sent_at,omitempty" json:"sentAt,omitempty"`
	DeliveredAt *time.Time         `bson:"delivered_at,omitempty" json:"deliveredAt,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}

// NotificationSendRequest represents the request to send a notification
type NotificationSendRequest struct {
	UserID    string            `json:"userId" binding:"required"`
	Type      NotificationType  `json:"type" binding:"required,oneof=email sms push"`
	Subject   string            `json:"subject"`
	Content   string            `json:"content" binding:"required"`
	Recipient string            `json:"recipient" binding:"required"`
	Metadata  map[string]string `json:"metadata"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID              string             `bson:"user_id" json:"userId"`
	EmailEnabled        bool               `bson:"email_enabled" json:"emailEnabled"`
	SMSEnabled          bool               `bson:"sms_enabled" json:"smsEnabled"`
	PushEnabled         bool               `bson:"push_enabled" json:"pushEnabled"`
	EmailAddress        string             `bson:"email_address,omitempty" json:"emailAddress,omitempty"`
	PhoneNumber         string             `bson:"phone_number,omitempty" json:"phoneNumber,omitempty"`
	PushDeviceTokens    []string           `bson:"push_device_tokens,omitempty" json:"pushDeviceTokens,omitempty"`
	QuietHoursEnabled   bool               `bson:"quiet_hours_enabled" json:"quietHoursEnabled"`
	QuietHoursStart     string             `bson:"quiet_hours_start,omitempty" json:"quietHoursStart,omitempty"` // e.g., "22:00"
	QuietHoursEnd       string             `bson:"quiet_hours_end,omitempty" json:"quietHoursEnd,omitempty"`     // e.g., "08:00"
	NotificationTypes   []string           `bson:"notification_types,omitempty" json:"notificationTypes,omitempty"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updatedAt"`
}

// NotificationPreferencesUpdateRequest represents the request to update notification preferences
type NotificationPreferencesUpdateRequest struct {
	UserID              string   `json:"userId" binding:"required"`
	EmailEnabled        *bool    `json:"emailEnabled"`
	SMSEnabled          *bool    `json:"smsEnabled"`
	PushEnabled         *bool    `json:"pushEnabled"`
	EmailAddress        string   `json:"emailAddress"`
	PhoneNumber         string   `json:"phoneNumber"`
	PushDeviceTokens    []string `json:"pushDeviceTokens"`
	QuietHoursEnabled   *bool    `json:"quietHoursEnabled"`
	QuietHoursStart     string   `json:"quietHoursStart"`
	QuietHoursEnd       string   `json:"quietHoursEnd"`
	NotificationTypes   []string `json:"notificationTypes"`
}

// NotificationLogQueryParams represents query parameters for notification logs
type NotificationLogQueryParams struct {
	UserID string           `form:"userId" binding:"required"`
	Type   NotificationType `form:"type"`
	Status string           `form:"status"`
	From   time.Time        `form:"from"`
	To     time.Time        `form:"to"`
	Page   int              `form:"page"`
	Limit  int              `form:"limit"`
}

// NotificationResponse represents the notification data returned in API responses
type NotificationResponse struct {
	ID          string            `json:"id"`
	UserID      string            `json:"userId"`
	Type        NotificationType  `json:"type"`
	Subject     string            `json:"subject"`
	Content     string            `json:"content"`
	Recipient   string            `json:"recipient"`
	Status      NotificationStatus `json:"status"`
	ErrorMsg    string            `json:"errorMsg,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	SentAt      *time.Time        `json:"sentAt,omitempty"`
	DeliveredAt *time.Time        `json:"deliveredAt,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// ToResponse converts a Notification to NotificationResponse
func (n *Notification) ToResponse() *NotificationResponse {
	return &NotificationResponse{
		ID:          n.ID.Hex(),
		UserID:      n.UserID,
		Type:        n.Type,
		Subject:     n.Subject,
		Content:     n.Content,
		Recipient:   n.Recipient,
		Status:      n.Status,
		ErrorMsg:    n.ErrorMsg,
		Metadata:    n.Metadata,
		SentAt:      n.SentAt,
		DeliveredAt: n.DeliveredAt,
		CreatedAt:   n.CreatedAt,
	}
}

// PaginatedNotificationsResponse represents a paginated list of notifications
type PaginatedNotificationsResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
	Total         int64                   `json:"total"`
	Page          int                     `json:"page"`
	Limit         int                     `json:"limit"`
	TotalPages    int                     `json:"totalPages"`
}
