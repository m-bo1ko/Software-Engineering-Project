package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"analytics-service/internal/middleware"
	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	dashboardService *service.DashboardService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetOverviewDashboard handles overview dashboard retrieval
// GET /analytics/dashboards/overview
func (h *DashboardHandler) GetOverviewDashboard(c *gin.Context) {
	token := middleware.GetToken(c)

	response, err := h.dashboardService.GetOverviewDashboard(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}

// GetBuildingDashboard handles building dashboard retrieval
// GET /analytics/dashboards/building/{buildingId}
func (h *DashboardHandler) GetBuildingDashboard(c *gin.Context) {
	buildingID := c.Param("buildingId")
	if buildingID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Building ID is required",
			"",
		))
		return
	}

	token := middleware.GetToken(c)

	response, err := h.dashboardService.GetBuildingDashboard(c.Request.Context(), buildingID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}
