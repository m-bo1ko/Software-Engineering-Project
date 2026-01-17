package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"analytics-service/internal/middleware"
	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// KPIHandler handles KPI-related requests
type KPIHandler struct {
	kpiService *service.KPIService
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewKPIHandler creates a new KPI handler
func NewKPIHandler(
	kpiService *service.KPIService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *KPIHandler {
	return &KPIHandler{
		kpiService:    kpiService,
		securityClient: securityClient,
	}
}

// GetKPIs handles KPI retrieval
// GET /analytics/kpi or GET /analytics/kpi/{buildingId}
func (h *KPIHandler) GetKPIs(c *gin.Context) {
	buildingID := c.Param("buildingId")
	period := c.DefaultQuery("period", "DAILY")

	response, err := h.kpiService.GetKPIs(c.Request.Context(), buildingID, period)
	if err != nil {
		if err.Error() == "KPI not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				err.Error(),
				"",
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeKPICalculationFailed,
				err.Error(),
				"",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}

// CalculateKPIs handles KPI calculation
// POST /analytics/kpi/calculate
func (h *KPIHandler) CalculateKPIs(c *gin.Context) {
	buildingID := c.Query("buildingId")
	period := c.DefaultQuery("period", "DAILY")

	token := middleware.GetToken(c)
	userID := middleware.GetUserID(c)
	ipAddress := middleware.GetClientIP(c)
	userAgent := middleware.GetUserAgent(c)

	response, err := h.kpiService.CalculateKPIs(c.Request.Context(), buildingID, period, token)
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "CALCULATE_KPI", "kpi", buildingID,
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"period": period},
		)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeKPICalculationFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "CALCULATE_KPI", "kpi", buildingID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"period": period},
	)
	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "KPIs calculated successfully"))
}
