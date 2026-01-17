package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"security-service/internal/models"
	"security-service/internal/service"
)

// AuditHandler handles audit logging requests
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// CreateLog creates a new audit log entry
// POST /audit/log
func (h *AuditHandler) CreateLog(c *gin.Context) {
	var req models.AuditLogCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Add request context if not provided
	if req.IPAddress == "" {
		req.IPAddress = c.ClientIP()
	}
	if req.UserAgent == "" {
		req.UserAgent = c.GetHeader("User-Agent")
	}

	log, err := h.auditService.CreateLog(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create audit log",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(log, "Audit log created successfully"))
}

// GetLogs retrieves audit logs with filters
// GET /audit/logs
func (h *AuditHandler) GetLogs(c *gin.Context) {
	var params models.AuditLogQueryParams

	// Parse query parameters
	if from := c.Query("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrCodeValidationFailed,
				"Invalid 'from' date format",
				"Expected RFC3339 format (e.g., 2024-01-15T10:00:00Z)",
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
				"Expected RFC3339 format (e.g., 2024-01-15T10:00:00Z)",
			))
			return
		}
		params.To = t
	}

	params.UserID = c.Query("userId")
	params.Service = c.Query("service")
	params.Action = c.Query("action")
	params.Resource = c.Query("resource")
	params.Status = c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	params.Page = page
	params.Limit = limit

	result, err := h.auditService.GetLogs(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve audit logs",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(result, ""))
}

// GetLog retrieves a specific audit log by ID
// GET /audit/logs/:id
func (h *AuditHandler) GetLog(c *gin.Context) {
	id := c.Param("id")

	log, err := h.auditService.GetLogByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "audit log not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"Audit log not found",
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve audit log",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(log, ""))
}
