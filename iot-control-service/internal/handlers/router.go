package handlers

import (
	"github.com/gin-gonic/gin"

	"iot-control-service/internal/middleware"
)

// Router holds all handler dependencies
type Router struct {
	DeviceHandler       *DeviceHandler
	TelemetryHandler    *TelemetryHandler
	ControlHandler      *ControlHandler
	OptimizationHandler *OptimizationHandler
	StateHandler        *StateHandler
	AuthMiddleware      *middleware.AuthMiddleware
}

// NewRouter creates a new router with all handlers
func NewRouter(
	deviceHandler *DeviceHandler,
	telemetryHandler *TelemetryHandler,
	controlHandler *ControlHandler,
	optimizationHandler *OptimizationHandler,
	stateHandler *StateHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		DeviceHandler:       deviceHandler,
		TelemetryHandler:    telemetryHandler,
		ControlHandler:      controlHandler,
		OptimizationHandler: optimizationHandler,
		StateHandler:        stateHandler,
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
			"service": "iot-control-service",
		})
	})

	// API v1 routes
	api := engine.Group("/api/v1")
	{
		r.setupTelemetryRoutes(api)
		r.setupDeviceRoutes(api)
		r.setupControlRoutes(api)
		r.setupOptimizationRoutes(api)
		r.setupStateRoutes(api)
	}

	// Legacy routes (without /api/v1 prefix for backward compatibility)
	r.setupLegacyRoutes(engine)
}

// setupTelemetryRoutes configures telemetry routes
func (r *Router) setupTelemetryRoutes(rg *gin.RouterGroup) {
	telemetry := rg.Group("/iot/telemetry")
	telemetry.Use(r.AuthMiddleware.RequireAuth())
	{
		telemetry.POST("", r.TelemetryHandler.IngestTelemetry)
		telemetry.POST("/bulk", r.TelemetryHandler.IngestBulkTelemetry)
		telemetry.GET("/history", r.TelemetryHandler.GetTelemetryHistory)
	}
}

// setupDeviceRoutes configures device routes
func (r *Router) setupDeviceRoutes(rg *gin.RouterGroup) {
	devices := rg.Group("/iot/devices")
	devices.Use(r.AuthMiddleware.RequireAuth())
	{
		devices.GET("", r.DeviceHandler.ListDevices)
		devices.GET("/:deviceId", r.DeviceHandler.GetDevice)
		devices.POST("/register", r.DeviceHandler.RegisterDevice)
	}
}

// setupControlRoutes configures control routes
func (r *Router) setupControlRoutes(rg *gin.RouterGroup) {
	control := rg.Group("/iot/device-control")
	control.Use(r.AuthMiddleware.RequireAuth())
	{
		control.POST("/:deviceId/command", r.ControlHandler.SendCommand)
		control.GET("/:deviceId/commands", r.ControlHandler.ListCommands)
	}
}

// setupOptimizationRoutes configures optimization routes
func (r *Router) setupOptimizationRoutes(rg *gin.RouterGroup) {
	optimization := rg.Group("/iot/optimization")
	optimization.Use(r.AuthMiddleware.RequireAuth())
	{
		optimization.POST("/apply", r.OptimizationHandler.ApplyOptimization)
		optimization.GET("/status/:scenarioId", r.OptimizationHandler.GetOptimizationStatus)
	}
}

// setupStateRoutes configures state routes
func (r *Router) setupStateRoutes(rg *gin.RouterGroup) {
	state := rg.Group("/iot/state")
	state.Use(r.AuthMiddleware.RequireAuth())
	{
		state.GET("/live", r.StateHandler.GetLiveState)
		state.GET("/:deviceId", r.StateHandler.GetDeviceState)
	}
}

// setupLegacyRoutes configures legacy routes without /api/v1 prefix
func (r *Router) setupLegacyRoutes(engine *gin.Engine) {
	// Telemetry routes
	telemetry := engine.Group("/iot/telemetry")
	telemetry.Use(r.AuthMiddleware.RequireAuth())
	{
		telemetry.POST("", r.TelemetryHandler.IngestTelemetry)
		telemetry.POST("/bulk", r.TelemetryHandler.IngestBulkTelemetry)
		telemetry.GET("/history", r.TelemetryHandler.GetTelemetryHistory)
	}

	// Device routes
	devices := engine.Group("/iot/devices")
	devices.Use(r.AuthMiddleware.RequireAuth())
	{
		devices.GET("", r.DeviceHandler.ListDevices)
		devices.GET("/:deviceId", r.DeviceHandler.GetDevice)
		devices.POST("/register", r.DeviceHandler.RegisterDevice)
	}

	// Control routes
	control := engine.Group("/iot/device-control")
	control.Use(r.AuthMiddleware.RequireAuth())
	{
		control.POST("/:deviceId/command", r.ControlHandler.SendCommand)
		control.GET("/:deviceId/commands", r.ControlHandler.ListCommands)
	}

	// Optimization routes
	optimization := engine.Group("/iot/optimization")
	optimization.Use(r.AuthMiddleware.RequireAuth())
	{
		optimization.POST("/apply", r.OptimizationHandler.ApplyOptimization)
		optimization.GET("/status/:scenarioId", r.OptimizationHandler.GetOptimizationStatus)
	}

	// State routes
	state := engine.Group("/iot/state")
	state.Use(r.AuthMiddleware.RequireAuth())
	{
		state.GET("/live", r.StateHandler.GetLiveState)
		state.GET("/:deviceId", r.StateHandler.GetDeviceState)
	}
}
