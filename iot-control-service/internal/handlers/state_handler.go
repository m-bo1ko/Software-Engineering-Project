package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"iot-control-service/internal/models"
	"iot-control-service/internal/service"
)

// StateHandler handles device state-related requests
type StateHandler struct {
	stateService *service.StateService
}

// NewStateHandler creates a new state handler
func NewStateHandler(stateService *service.StateService) *StateHandler {
	return &StateHandler{
		stateService: stateService,
	}
}

// GetLiveState handles live state retrieval
// GET /iot/state/live
func (h *StateHandler) GetLiveState(c *gin.Context) {
	response, err := h.stateService.GetLiveState(c.Request.Context())
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

// GetDeviceState handles device state retrieval
// GET /iot/state/{deviceId}
func (h *StateHandler) GetDeviceState(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Device ID is required",
			"",
		))
		return
	}

	response, err := h.stateService.GetDeviceState(c.Request.Context(), deviceID)
	if err != nil {
		if err.Error() == "device not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeDeviceNotFound,
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
