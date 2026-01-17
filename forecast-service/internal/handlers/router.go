package handlers

import (
	"github.com/gin-gonic/gin"

	"forecast-service/internal/middleware"
)

// Router holds all handler dependencies
type Router struct {
	ForecastHandler      *ForecastHandler
	OptimizationHandler  *OptimizationHandler
	AuthMiddleware       *middleware.AuthMiddleware
}

// NewRouter creates a new router with all handlers
func NewRouter(
	forecastHandler *ForecastHandler,
	optimizationHandler *OptimizationHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		ForecastHandler:     forecastHandler,
		OptimizationHandler: optimizationHandler,
		AuthMiddleware:      authMiddleware,
	}
}

// SetupRoutes configures all API routes
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// Apply common middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS())
	engine.Use(middleware.SecurityHeaders())
	engine.Use(middleware.RequestLogger())

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "forecast-service",
		})
	})

	// API v1 routes
	api := engine.Group("/api/v1")
	{
		r.setupForecastRoutes(api)
		r.setupOptimizationRoutes(api)
	}

	// Legacy routes (without /api/v1 prefix for backward compatibility)
	r.setupLegacyRoutes(engine)
}

// setupForecastRoutes configures forecast routes
func (r *Router) setupForecastRoutes(rg *gin.RouterGroup) {
	forecast := rg.Group("/forecast")
	forecast.Use(r.AuthMiddleware.RequireAuth())
	{
		forecast.POST("/generate", r.ForecastHandler.GenerateForecast)
		forecast.POST("/peak-load", r.ForecastHandler.GeneratePeakLoad)
		forecast.GET("/latest", r.ForecastHandler.GetLatestForecast)
		forecast.GET("/prediction/:deviceId", r.ForecastHandler.GetDevicePrediction)
	}

	// Optimization endpoint for device (used in forecast routes)
	forecast.GET("/optimization/:deviceId", r.AuthMiddleware.RequireAuth(), r.OptimizationHandler.GetDeviceOptimization)
}

// setupOptimizationRoutes configures optimization routes
func (r *Router) setupOptimizationRoutes(rg *gin.RouterGroup) {
	optimization := rg.Group("/optimization")
	optimization.Use(r.AuthMiddleware.RequireAuth())
	{
		optimization.POST("/generate", r.OptimizationHandler.GenerateOptimization)
		optimization.GET("/recommendations/:buildingId", r.OptimizationHandler.GetRecommendations)
		optimization.GET("/scenario/:scenarioId", r.OptimizationHandler.GetScenario)
		optimization.POST("/send-to-iot", r.OptimizationHandler.SendToIoT)
	}
}

// setupLegacyRoutes configures legacy routes without /api/v1 prefix
func (r *Router) setupLegacyRoutes(engine *gin.Engine) {
	// Forecast routes
	forecast := engine.Group("/forecast")
	forecast.Use(r.AuthMiddleware.RequireAuth())
	{
		forecast.POST("/generate", r.ForecastHandler.GenerateForecast)
		forecast.POST("/peak-load", r.ForecastHandler.GeneratePeakLoad)
		forecast.GET("/latest", r.ForecastHandler.GetLatestForecast)
		forecast.GET("/prediction/:deviceId", r.ForecastHandler.GetDevicePrediction)
		forecast.GET("/optimization/:deviceId", r.OptimizationHandler.GetDeviceOptimization)
	}

	// Optimization routes
	optimization := engine.Group("/optimization")
	optimization.Use(r.AuthMiddleware.RequireAuth())
	{
		optimization.POST("/generate", r.OptimizationHandler.GenerateOptimization)
		optimization.GET("/recommendations/:buildingId", r.OptimizationHandler.GetRecommendations)
		optimization.GET("/scenario/:scenarioId", r.OptimizationHandler.GetScenario)
		optimization.POST("/send-to-iot", r.OptimizationHandler.SendToIoT)
	}
}

