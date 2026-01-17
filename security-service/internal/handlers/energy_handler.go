package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"security-service/internal/integrations"
	"security-service/internal/models"
)

// EnergyHandler handles external energy provider integration requests
type EnergyHandler struct {
	energyClient *integrations.EnergyProviderClient
}

// NewEnergyHandler creates a new energy handler
func NewEnergyHandler(energyClient *integrations.EnergyProviderClient) *EnergyHandler {
	return &EnergyHandler{energyClient: energyClient}
}

// GetConsumption retrieves energy consumption data
// GET /external-energy/consumption
func (h *EnergyHandler) GetConsumption(c *gin.Context) {
	buildingID := c.Query("buildingId")
	if buildingID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"buildingId query parameter is required",
			"",
		))
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"from and to query parameters are required",
			"Expected RFC3339 format (e.g., 2024-01-15T00:00:00Z)",
		))
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid 'from' date format",
			"Expected RFC3339 format (e.g., 2024-01-15T00:00:00Z)",
		))
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid 'to' date format",
			"Expected RFC3339 format (e.g., 2024-01-15T00:00:00Z)",
		))
		return
	}

	consumption, err := h.energyClient.GetConsumption(c.Request.Context(), buildingID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeExternalAPIError,
			"Failed to retrieve energy consumption data",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(consumption, ""))
}

// GetTariffs retrieves tariff information
// GET /external-energy/tariffs
func (h *EnergyHandler) GetTariffs(c *gin.Context) {
	region := c.Query("region")
	if region == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"region query parameter is required",
			"",
		))
		return
	}

	tariff, err := h.energyClient.GetTariffs(c.Request.Context(), region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeExternalAPIError,
			"Failed to retrieve tariff data",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(tariff, ""))
}

// RefreshToken refreshes the external API token
// POST /external-energy/refresh-token
func (h *EnergyHandler) RefreshToken(c *gin.Context) {
	var req models.ExternalTokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if req.Provider != "energy_provider" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Unsupported provider",
			"Currently only 'energy_provider' is supported",
		))
		return
	}

	response, err := h.energyClient.RefreshToken(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeExternalAPIError,
			"Failed to refresh token",
			err.Error(),
		))
		return
	}

	if !response.Success {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeExternalAPIError,
			"Token refresh failed",
			response.Message,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Token refreshed successfully"))
}
