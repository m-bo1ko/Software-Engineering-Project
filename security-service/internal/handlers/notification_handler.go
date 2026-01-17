package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"security-service/internal/models"
	"security-service/internal/service"
)

// NotificationHandler handles notification requests
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// SendNotification sends a notification
// POST /notifications/send
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req models.NotificationSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	notification, err := h.notificationService.SendNotification(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrNotificationDisabled {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrCodeInvalidRequest,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to send notification",
			err.Error(),
		))
		return
	}

	// Check if notification was actually sent
	statusCode := http.StatusOK
	message := "Notification sent successfully"
	if notification.Status == models.NotificationStatusFailed {
		message = "Notification queued but delivery failed"
	}

	c.JSON(statusCode, models.NewSuccessResponse(notification, message))
}

// UpdatePreferences updates user notification preferences
// POST /notifications/preferences
func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	var req models.NotificationPreferencesUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	prefs, err := h.notificationService.UpdatePreferences(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to update notification preferences",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(prefs, "Notification preferences updated successfully"))
}

// GetLogs retrieves notification history for a user
// GET /notifications/logs
func (h *NotificationHandler) GetLogs(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"userId query parameter is required",
			"",
		))
		return
	}

	var params models.NotificationLogQueryParams
	params.UserID = userID
	params.Type = models.NotificationType(c.Query("type"))
	params.Status = c.Query("status")

	// Parse time parameters
	if from := c.Query("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrCodeValidationFailed,
				"Invalid 'from' date format",
				"Expected RFC3339 format",
			))
			return
		}
		params.From = t
	}

	if to := c.Query("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrCodeValidationFailed,
				"Invalid 'to' date format",
				"Expected RFC3339 format",
			))
			return
		}
		params.To = t
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	params.Page = page
	params.Limit = limit

	result, err := h.notificationService.GetLogs(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve notification logs",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result, ""))
}
