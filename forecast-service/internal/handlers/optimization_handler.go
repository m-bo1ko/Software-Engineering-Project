// Package handlers contains HTTP request handlers
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"forecast-service/internal/middleware"
	"forecast-service/internal/models"
	"forecast-service/internal/service"
)

// OptimizationHandler handles optimization-related requests
type OptimizationHandler struct {
	optimizationService *service.OptimizationService
	securityClient      interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	}
}

// NewOptimizationHandler creates a new optimization handler
func NewOptimizationHandler(
	optimizationService *service.OptimizationService,
	securityClient interface {
		AuditLog(ctx interface{}, userID, username, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{})
	},
) *OptimizationHandler {
	return &OptimizationHandler{
		optimizationService: optimizationService,
		securityClient:      securityClient,
	}
}

// GenerateOptimization handles optimization scenario generation
// POST /optimization/generate
func (h *OptimizationHandler) GenerateOptimization(c *gin.Context) {
	var req models.OptimizationGenerateRequest
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

	response, err := h.optimizationService.GenerateOptimization(c.Request.Context(), &req, userID, token)
	if err != nil {
		h.securityClient.AuditLog(c.Request.Context(), userID, "", "GENERATE_OPTIMIZATION", "optimization", "", "FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, map[string]interface{}{"buildingId": req.BuildingID})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeOptimizationFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(c.Request.Context(), userID, "", "GENERATE_OPTIMIZATION", "optimization", response.ID, "SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, map[string]interface{}{"buildingId": req.BuildingID})
	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Optimization scenario generated successfully"))
}

// GetRecommendations retrieves energy-saving recommendations for a building
// GET /optimization/recommendations/:buildingId
func (h *OptimizationHandler) GetRecommendations(c *gin.Context) {
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
	response, err := h.optimizationService.GetRecommendations(c.Request.Context(), buildingID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeOptimizationFailed,
			err.Error(),
			"",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, ""))
}

// GetScenario retrieves an optimization scenario by ID
// GET /optimization/scenario/:scenarioId
func (h *OptimizationHandler) GetScenario(c *gin.Context) {
	scenarioID := c.Param("scenarioId")
	if scenarioID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Scenario ID is required",
			"",
		))
		return
	}

	response, err := h.optimizationService.GetScenario(c.Request.Context(), scenarioID)
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

// SendToIoT sends an optimization scenario to IoT service
// POST /optimization/send-to-iot
func (h *OptimizationHandler) SendToIoT(c *gin.Context) {
	var req models.SendToIoTRequest
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

	response, err := h.optimizationService.SendToIoT(c.Request.Context(), &req, userID, token)
	if err != nil {
		h.securityClient.AuditLog(c.Request.Context(), userID, "", "SEND_TO_IOT", "optimization", req.ScenarioID, "FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, nil)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeOptimizationFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(c.Request.Context(), userID, "", "SEND_TO_IOT", "optimization", req.ScenarioID, "SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method, nil)
	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Scenario sent to IoT service successfully"))
}

// GetDeviceOptimization retrieves optimization recommendations for a device
// GET /forecast/optimization/:deviceId
func (h *OptimizationHandler) GetDeviceOptimization(c *gin.Context) {
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
	response, err := h.optimizationService.GetDeviceOptimization(c.Request.Context(), deviceID, token)
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
