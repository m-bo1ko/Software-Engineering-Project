package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"iot-control-service/internal/middleware"
	"iot-control-service/internal/models"
	"iot-control-service/internal/service"
)

// ControlHandler handles device control-related requests
type ControlHandler struct {
	controlService *service.ControlService
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewControlHandler creates a new control handler
func NewControlHandler(
	controlService *service.ControlService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *ControlHandler {
	return &ControlHandler{
		controlService: controlService,
		securityClient: securityClient,
	}
}

// SendCommand handles command sending
// POST /iot/device-control/{deviceId}/command
func (h *ControlHandler) SendCommand(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Device ID is required",
			"",
		))
		return
	}

	var req models.SendCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	userID := middleware.GetUserID(c)
	ipAddress := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	response, err := h.controlService.SendCommand(c.Request.Context(), deviceID, &req, userID)
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "SEND_COMMAND", "command", "",
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"deviceId": deviceID, "command": req.Command},
		)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeCommandFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "SEND_COMMAND", "command", response.CommandID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"deviceId": deviceID, "command": req.Command},
	)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(response, "Command sent successfully"))
}

// ListCommands handles command listing
// GET /iot/device-control/{deviceId}/commands
func (h *ControlHandler) ListCommands(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Device ID is required",
			"",
		))
		return
	}

	var req models.ListCommandsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	responses, total, err := h.controlService.ListCommands(
		c.Request.Context(),
		deviceID,
		req.Status,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"commands": responses,
		"total":    total,
		"page":     req.Page,
		"limit":    req.Limit,
	}, ""))
}
