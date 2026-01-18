package service

import (
	"context"

	"security-service/internal/integrations"
	"security-service/internal/models"
	"security-service/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	client           *integrations.NotificationClient
}

// NewNotificationService creates a new notification service
func NewNotificationService(notificationRepo *repository.NotificationRepository, client *integrations.NotificationClient) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		client:           client,
	}
}

// SendNotification sends a notification to a user
func (s *NotificationService) SendNotification(ctx context.Context, req *models.NotificationSendRequest) (*models.NotificationResponse, error) {
	// Check user preferences
	prefs, err := s.notificationRepo.GetPreferences(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Check if notification type is enabled
	switch req.Type {
	case models.NotificationTypeEmail:
		if !prefs.EmailEnabled {
			return nil, ErrNotificationDisabled
		}
	case models.NotificationTypeSMS:
		if !prefs.SMSEnabled {
			return nil, ErrNotificationDisabled
		}
	case models.NotificationTypePush:
		if !prefs.PushEnabled {
			return nil, ErrNotificationDisabled
		}
	}

	// Create notification record
	notification := &models.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Subject:   req.Subject,
		Content:   req.Content,
		Recipient: req.Recipient,
		Metadata:  req.Metadata,
	}

	createdNotification, err := s.notificationRepo.Create(ctx, notification)
	if err != nil {
		return nil, err
	}

	// Send notification via external service
	var sendErr error
	switch req.Type {
	case models.NotificationTypeEmail:
		sendErr = s.client.SendEmail(ctx, req.Recipient, req.Subject, req.Content)
	case models.NotificationTypeSMS:
		sendErr = s.client.SendSMS(ctx, req.Recipient, req.Content)
	case models.NotificationTypePush:
		sendErr = s.client.SendPush(ctx, req.Recipient, req.Subject, req.Content)
	}

	// Update notification status
	if sendErr != nil {
		s.notificationRepo.UpdateStatus(ctx, createdNotification.ID.Hex(), models.NotificationStatusFailed, sendErr.Error())
		createdNotification.Status = models.NotificationStatusFailed
		createdNotification.ErrorMsg = sendErr.Error()
	} else {
		s.notificationRepo.UpdateStatus(ctx, createdNotification.ID.Hex(), models.NotificationStatusSent, "")
		createdNotification.Status = models.NotificationStatusSent
	}

	return createdNotification.ToResponse(), nil
}

// UpdatePreferences updates user notification preferences
func (s *NotificationService) UpdatePreferences(ctx context.Context, req *models.NotificationPreferencesUpdateRequest) (*models.NotificationPreferences, error) {
	// Get existing preferences
	prefs, err := s.notificationRepo.GetPreferences(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Update preferences
	if req.EmailEnabled != nil {
		prefs.EmailEnabled = *req.EmailEnabled
	}
	if req.SMSEnabled != nil {
		prefs.SMSEnabled = *req.SMSEnabled
	}
	if req.PushEnabled != nil {
		prefs.PushEnabled = *req.PushEnabled
	}
	if req.EmailAddress != "" {
		prefs.EmailAddress = req.EmailAddress
	}
	if req.PhoneNumber != "" {
		prefs.PhoneNumber = req.PhoneNumber
	}
	if req.PushDeviceTokens != nil {
		prefs.PushDeviceTokens = req.PushDeviceTokens
	}
	if req.QuietHoursEnabled != nil {
		prefs.QuietHoursEnabled = *req.QuietHoursEnabled
	}
	if req.QuietHoursStart != "" {
		prefs.QuietHoursStart = req.QuietHoursStart
	}
	if req.QuietHoursEnd != "" {
		prefs.QuietHoursEnd = req.QuietHoursEnd
	}
	if req.NotificationTypes != nil {
		prefs.NotificationTypes = req.NotificationTypes
	}

	// Save preferences
	if err := s.notificationRepo.SavePreferences(ctx, prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// GetPreferences retrieves user notification preferences
func (s *NotificationService) GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error) {
	return s.notificationRepo.GetPreferences(ctx, userID)
}

// GetLogs retrieves notification history for a user
func (s *NotificationService) GetLogs(ctx context.Context, params models.NotificationLogQueryParams) (*models.PaginatedNotificationsResponse, error) {
	return s.notificationRepo.GetPaginatedResponse(ctx, params)
}

// Custom errors
var (
	ErrNotificationDisabled = NewServiceError("notification type is disabled for this user")
)

// ServiceError represents a service-level error
type ServiceError struct {
	message string
}

func NewServiceError(message string) *ServiceError {
	return &ServiceError{message: message}
}

func (e *ServiceError) Error() string {
	return e.message
}
