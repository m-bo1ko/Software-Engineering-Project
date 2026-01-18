// Package handlers contains HTTP request handlers
package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"iot-control-service/internal/middleware"
	"iot-control-service/internal/models"
	"iot-control-service/internal/service"
)

// DeviceHandler handles device-related requests
type DeviceHandler struct {
	deviceService  *service.DeviceService
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(
	deviceService *service.DeviceService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *DeviceHandler {
	return &DeviceHandler{
		deviceService:  deviceService,
		securityClient: securityClient,
	}
}

// RegisterDevice handles device registration
// POST /iot/devices/register
func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var req models.RegisterDeviceRequest
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

	response, err := h.deviceService.RegisterDevice(c.Request.Context(), &req, userID)
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "REGISTER_DEVICE", "device", "",
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"deviceId": req.DeviceID},
		)
		// Check for duplicate device error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.ErrCodeDeviceExists,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "REGISTER_DEVICE", "device", response.ID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"deviceId": response.DeviceID},
	)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(response, "Device registered successfully"))
}

// GetDevice handles device retrieval
// GET /iot/devices/{deviceId}
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Device ID is required",
			"",
		))
		return
	}

	response, err := h.deviceService.GetDevice(c.Request.Context(), deviceID)
	if err != nil {
		if err.Error() == "device not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeDeviceNotFound,
				err.Error(),
				"",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeInternalError,
				err.Error(),
				"",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}

// ListDevices handles device listing
// GET /iot/devices
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	var req models.ListDevicesRequest
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

	responses, total, err := h.deviceService.ListDevices(
		c.Request.Context(),
		req.BuildingID,
		req.Type,
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
		"devices": responses,
		"total":   total,
		"page":    req.Page,
		"limit":   req.Limit,
	}, ""))
}
