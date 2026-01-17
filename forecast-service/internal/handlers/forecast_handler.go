// Package handlers contains HTTP request handlers
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"forecast-service/internal/middleware"
	"forecast-service/internal/models"
	"forecast-service/internal/service"
)

// ForecastHandler handles forecast-related requests
type ForecastHandler struct {
	forecastService *service.ForecastService
	securityClient  interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewForecastHandler creates a new forecast handler
func NewForecastHandler(forecastService *service.ForecastService, securityClient interface {
	AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
}) *ForecastHandler {
	return &ForecastHandler{
		forecastService: forecastService,
		securityClient:  securityClient,
	}
}

// GenerateForecast handles forecast generation
// POST /forecast/generate
func (h *ForecastHandler) GenerateForecast(c *gin.Context) {
	var req models.ForecastGenerateRequest
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

	response, err := h.forecastService.GenerateForecast(c.Request.Context(), &req, userID, token)
	if err != nil {
		h.securityClient.AuditLog(c.Request.Context(), userID, "", "GENERATE_FORECAST", "forecast", "", "FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, map[string]interface{}{"buildingId": req.BuildingID})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeForecastFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(c.Request.Context(), userID, "", "GENERATE_FORECAST", "forecast", response.ID, "SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, map[string]interface{}{"buildingId": req.BuildingID})
	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Forecast generated successfully"))
}

// GeneratePeakLoad handles peak load prediction
// POST /forecast/peak-load
func (h *ForecastHandler) GeneratePeakLoad(c *gin.Context) {
	var req models.PeakLoadRequest
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

	response, err := h.forecastService.GeneratePeakLoad(c.Request.Context(), &req, userID, token)
	if err != nil {
		h.securityClient.AuditLog(c.Request.Context(), userID, "", "GENERATE_PEAK_LOAD", "peak_load", "", "FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, map[string]interface{}{"buildingId": req.BuildingID})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeForecastFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(c.Request.Context(), userID, "", "GENERATE_PEAK_LOAD", "peak_load", response.ID, "SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, map[string]interface{}{"buildingId": req.BuildingID})
	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Peak load prediction generated successfully"))
}

// GetLatestForecast retrieves the latest forecast for a building
// GET /forecast/latest?buildingId=
func (h *ForecastHandler) GetLatestForecast(c *gin.Context) {
	buildingID := c.Query("buildingId")
	if buildingID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"buildingId query parameter is required",
			"",
		))
		return
	}

	forecastType := models.ForecastType(c.DefaultQuery("type", ""))
	if forecastType == "" {
		forecastType = models.ForecastTypeDemand
	}

	response, err := h.forecastService.GetLatestForecast(c.Request.Context(), buildingID, forecastType)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}

// GetForecastByID retrieves a forecast by ID
// GET /forecast/:id
func (h *ForecastHandler) GetForecastByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Forecast ID is required",
			"",
		))
		return
	}

	response, err := h.forecastService.GetForecastByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}

// GetDevicePrediction retrieves predicted consumption for a device
// GET /forecast/prediction/:deviceId
func (h *ForecastHandler) GetDevicePrediction(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Device ID is required",
			"",
		))
		return
	}

	token := middleware.GetToken(c)
	response, err := h.forecastService.GetDevicePrediction(c.Request.Context(), deviceID, token)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}
