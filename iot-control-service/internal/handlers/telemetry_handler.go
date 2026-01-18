package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"iot-control-service/internal/middleware"
	"iot-control-service/internal/models"
	"iot-control-service/internal/service"
)

// TelemetryHandler handles telemetry-related requests
type TelemetryHandler struct {
	telemetryService *service.TelemetryService
	securityClient   interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewTelemetryHandler creates a new telemetry handler
func NewTelemetryHandler(
	telemetryService *service.TelemetryService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *TelemetryHandler {
	return &TelemetryHandler{
		telemetryService: telemetryService,
		securityClient:   securityClient,
	}
}

// IngestTelemetry handles single telemetry ingestion
// POST /iot/telemetry
func (h *TelemetryHandler) IngestTelemetry(c *gin.Context) {
	var req models.TelemetryIngestRequest
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

	response, err := h.telemetryService.IngestTelemetry(c.Request.Context(), &req, "HTTP")
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "INGEST_TELEMETRY", "telemetry", "",
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"deviceId": req.DeviceID},
		)
		// Check if device not found error
		if strings.Contains(err.Error(), "device not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeDeviceNotFound,
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
		c.Request.Context(), userID, "", "INGEST_TELEMETRY", "telemetry", response.ID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"deviceId": req.DeviceID},
	)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(response, "Telemetry ingested successfully"))
}

// IngestBulkTelemetry handles bulk telemetry ingestion
// POST /iot/telemetry/bulk
func (h *TelemetryHandler) IngestBulkTelemetry(c *gin.Context) {
	var req models.BulkTelemetryIngestRequest
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

	responses, err := h.telemetryService.IngestBulkTelemetry(c.Request.Context(), &req, "HTTP")
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "INGEST_BULK_TELEMETRY", "telemetry", "",
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"count": len(req.Telemetry)},
		)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "INGEST_BULK_TELEMETRY", "telemetry", "",
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"count": len(responses)},
	)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(responses, "Bulk telemetry ingested successfully"))
}

// GetTelemetryHistory handles telemetry history retrieval
// GET /iot/telemetry/history
func (h *TelemetryHandler) GetTelemetryHistory(c *gin.Context) {
	var req models.TelemetryHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	if req.DeviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"deviceId query parameter is required",
			"",
		))
		return
	}

	if req.From.IsZero() {
		req.From = time.Now().AddDate(0, 0, -7) // Default to last 7 days
	}
	if req.To.IsZero() {
		req.To = time.Now()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 100
	}

	responses, total, err := h.telemetryService.GetTelemetryHistory(
		c.Request.Context(),
		req.DeviceID,
		req.From,
		req.To,
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
		"telemetry": responses,
		"total":     total,
		"page":      req.Page,
		"limit":     req.Limit,
	}, ""))
}
