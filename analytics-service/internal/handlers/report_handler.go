// Package handlers contains HTTP request handlers
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"analytics-service/internal/middleware"
	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// ReportHandler handles report-related requests
type ReportHandler struct {
	reportService *service.ReportService
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	reportService *service.ReportService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *ReportHandler {
	return &ReportHandler{
		reportService:  reportService,
		securityClient: securityClient,
	}
}

// GenerateReport handles report generation
// POST /analytics/reports/generate
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	var req models.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	userID := middleware.GetUserID(c)
	token := middleware.GetToken(c)
	ipAddress := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	response, err := h.reportService.GenerateReport(c.Request.Context(), &req, userID, token)
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "GENERATE_REPORT", "report", "",
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"type": req.Type, "buildingId": req.BuildingID},
		)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "GENERATE_REPORT", "report", response.ReportID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"type": req.Type, "buildingId": req.BuildingID},
	)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(response, "Report generation started"))
}

// GetReport handles report retrieval
// GET /analytics/reports/{reportId}
func (h *ReportHandler) GetReport(c *gin.Context) {
	reportID := c.Param("reportId")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Report ID is required",
			"",
		))
		return
	}

	response, err := h.reportService.GetReport(c.Request.Context(), reportID)
	if err != nil {
		if err.Error() == "report not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeReportNotFound,
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

// ListReports handles report listing
// GET /analytics/reports
func (h *ReportHandler) ListReports(c *gin.Context) {
	var req models.ListReportsRequest
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

	responses, total, err := h.reportService.ListReports(
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
		"reports": responses,
		"total":   total,
		"page":    req.Page,
		"limit":   req.Limit,
	}, ""))
}
