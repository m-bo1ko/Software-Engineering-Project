package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"iot-control-service/internal/middleware"
	"iot-control-service/internal/models"
	"iot-control-service/internal/service"
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
		securityClient:       securityClient,
	}
}

// ApplyOptimization handles optimization scenario application
// POST /iot/optimization/apply
func (h *OptimizationHandler) ApplyOptimization(c *gin.Context) {
	var req models.ApplyOptimizationRequest
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

	response, err := h.optimizationService.ApplyOptimization(c.Request.Context(), &req, userID)
	if err != nil {
		h.securityClient.AuditLog(
			c.Request.Context(), userID, "", "APPLY_OPTIMIZATION", "optimization", "",
			"FAILURE", err.Error(), ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
			map[string]interface{}{"scenarioId": req.ScenarioID, "buildingId": req.BuildingID},
		)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeOptimizationFailed,
			err.Error(),
			"",
		))
		return
	}

	h.securityClient.AuditLog(
		c.Request.Context(), userID, "", "APPLY_OPTIMIZATION", "optimization", response.ScenarioID,
		"SUCCESS", "", ipAddress, userAgent, c.Request.URL.Path, c.Request.Method,
		map[string]interface{}{"scenarioId": response.ScenarioID, "buildingId": req.BuildingID},
	)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(response, "Optimization scenario applied successfully"))
}

// GetOptimizationStatus handles optimization status retrieval
// GET /iot/optimization/status/{scenarioId}
func (h *OptimizationHandler) GetOptimizationStatus(c *gin.Context) {
	scenarioID := c.Param("scenarioId")
	if scenarioID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Scenario ID is required",
			"",
		))
		return
	}

	response, err := h.optimizationService.GetOptimizationStatus(c.Request.Context(), scenarioID)
	if err != nil {
		if err.Error() == "scenario not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
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
