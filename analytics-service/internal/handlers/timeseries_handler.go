package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"analytics-service/internal/middleware"
	"analytics-service/internal/models"
	"analytics-service/internal/service"
)

// TimeSeriesHandler handles time-series related requests
type TimeSeriesHandler struct {
	timeSeriesService *service.TimeSeriesService
}

// NewTimeSeriesHandler creates a new time-series handler
func NewTimeSeriesHandler(timeSeriesService *service.TimeSeriesService) *TimeSeriesHandler {
	return &TimeSeriesHandler{
		timeSeriesService: timeSeriesService,
	}
}

// QueryTimeSeries handles time-series query
// POST /analytics/time-series/query
func (h *TimeSeriesHandler) QueryTimeSeries(c *gin.Context) {
	var req models.TimeSeriesQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	token := middleware.GetToken(c)

	responses, err := h.timeSeriesService.QueryTimeSeries(c.Request.Context(), &req, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(responses, ""))
}
