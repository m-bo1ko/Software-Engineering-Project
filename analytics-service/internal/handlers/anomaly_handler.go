package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"analytics-service/internal/middleware"
	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// AnomalyHandler handles anomaly-related requests
type AnomalyHandler struct {
	anomalyService *service.AnomalyService
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewAnomalyHandler creates a new anomaly handler
func NewAnomalyHandler(
	anomalyService *service.AnomalyService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *AnomalyHandler {
	return &AnomalyHandler{
		anomalyService: anomalyService,
		securityClient: securityClient,
	}
}

// GetAnomaly handles anomaly retrieval
// GET /analytics/anomalies/{anomalyId}
func (h *AnomalyHandler) GetAnomaly(c *gin.Context) {
	anomalyID := c.Param("anomalyId")
	if anomalyID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Anomaly ID is required",
			"",
		))
		return
	}

	response, err := h.anomalyService.GetAnomaly(c.Request.Context(), anomalyID)
	if err != nil {
		if err.Error() == "anomaly not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeAnomalyNotFound,
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

// ListAnomalies handles anomaly listing
// GET /analytics/anomalies
func (h *AnomalyHandler) ListAnomalies(c *gin.Context) {
	var req models.ListAnomaliesRequest
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

	responses, total, err := h.anomalyService.ListAnomalies(
		c.Request.Context(),
		req.DeviceID,
		req.BuildingID,
		req.Type,
		req.Severity,
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
		"anomalies": responses,
		"total":     total,
		"page":      req.Page,
		"limit":     req.Limit,
	}, ""))
}

// AcknowledgeAnomaly handles anomaly acknowledgment
// POST /analytics/anomalies/acknowledge
func (h *AnomalyHandler) AcknowledgeAnomaly(c *gin.Context) {
	var req models.AcknowledgeAnomalyRequest
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

	response, err := h.anomalyService.AcknowledgeAnomaly(c.Request.Context(), req.AnomalyID, userID)
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "ACKNOWLEDGE_ANOMALY", "anomaly", req.AnomalyID,
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			nil,
		)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "ACKNOWLEDGE_ANOMALY", "anomaly", req.AnomalyID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		nil,
	)
	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Anomaly acknowledged successfully"))
}
