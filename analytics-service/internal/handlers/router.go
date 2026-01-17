package handlers

import (
	"github.com/gin-gonic/gin"

	"analytics-service/internal/middleware"
)

// Router holds all handler dependencies
type Router struct {
	ReportHandler      *ReportHandler
	AnomalyHandler     *AnomalyHandler
	TimeSeriesHandler  *TimeSeriesHandler
	KPIHandler         *KPIHandler
	DashboardHandler   *DashboardHandler
	AuthMiddleware     *middleware.AuthMiddleware
}

// NewRouter creates a new router with all handlers
func NewRouter(
	reportHandler *ReportHandler,
	anomalyHandler *AnomalyHandler,
	timeSeriesHandler *TimeSeriesHandler,
	kpiHandler *KPIHandler,
	dashboardHandler *DashboardHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		ReportHandler:     reportHandler,
		AnomalyHandler:    anomalyHandler,
		TimeSeriesHandler: timeSeriesHandler,
		KPIHandler:        kpiHandler,
		DashboardHandler:  dashboardHandler,
		AuthMiddleware:    authMiddleware,
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
			"service": "analytics-service",
		})
	})

	// API v1 routes
	api := engine.Group("/api/v1")
	{
		r.setupReportRoutes(api)
		r.setupAnomalyRoutes(api)
		r.setupTimeSeriesRoutes(api)
		r.setupKPIRoutes(api)
		r.setupDashboardRoutes(api)
	}

	// Legacy routes (without /api/v1 prefix for backward compatibility)
	r.setupLegacyRoutes(engine)
}

// setupReportRoutes configures report routes
func (r *Router) setupReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/analytics/reports")
	reports.Use(r.AuthMiddleware.RequireAuth())
	{
		reports.GET("", r.ReportHandler.ListReports)
		reports.GET("/:reportId", r.ReportHandler.GetReport)
		reports.POST("/generate", r.ReportHandler.GenerateReport)
	}
}

// setupAnomalyRoutes configures anomaly routes
func (r *Router) setupAnomalyRoutes(rg *gin.RouterGroup) {
	anomalies := rg.Group("/analytics/anomalies")
	anomalies.Use(r.AuthMiddleware.RequireAuth())
	{
		anomalies.GET("", r.AnomalyHandler.ListAnomalies)
		anomalies.GET("/:anomalyId", r.AnomalyHandler.GetAnomaly)
		anomalies.POST("/acknowledge", r.AnomalyHandler.AcknowledgeAnomaly)
	}
}

// setupTimeSeriesRoutes configures time-series routes
func (r *Router) setupTimeSeriesRoutes(rg *gin.RouterGroup) {
	timeseries := rg.Group("/analytics/time-series")
	timeseries.Use(r.AuthMiddleware.RequireAuth())
	{
		timeseries.POST("/query", r.TimeSeriesHandler.QueryTimeSeries)
	}
}

// setupKPIRoutes configures KPI routes
func (r *Router) setupKPIRoutes(rg *gin.RouterGroup) {
	kpi := rg.Group("/analytics/kpi")
	kpi.Use(r.AuthMiddleware.RequireAuth())
	{
		kpi.GET("", r.KPIHandler.GetKPIs)
		kpi.GET("/:buildingId", r.KPIHandler.GetKPIs)
		kpi.POST("/calculate", r.KPIHandler.CalculateKPIs)
	}
}

// setupDashboardRoutes configures dashboard routes
func (r *Router) setupDashboardRoutes(rg *gin.RouterGroup) {
	dashboards := rg.Group("/analytics/dashboards")
	dashboards.Use(r.AuthMiddleware.RequireAuth())
	{
		dashboards.GET("/overview", r.DashboardHandler.GetOverviewDashboard)
		dashboards.GET("/building/:buildingId", r.DashboardHandler.GetBuildingDashboard)
	}
}

// setupLegacyRoutes configures legacy routes without /api/v1 prefix
func (r *Router) setupLegacyRoutes(engine *gin.Engine) {
	// Report routes
	reports := engine.Group("/analytics/reports")
	reports.Use(r.AuthMiddleware.RequireAuth())
	{
		reports.GET("", r.ReportHandler.ListReports)
		reports.GET("/:reportId", r.ReportHandler.GetReport)
		reports.POST("/generate", r.ReportHandler.GenerateReport)
	}

	// Anomaly routes
	anomalies := engine.Group("/analytics/anomalies")
	anomalies.Use(r.AuthMiddleware.RequireAuth())
	{
		anomalies.GET("", r.AnomalyHandler.ListAnomalies)
		anomalies.GET("/:anomalyId", r.AnomalyHandler.GetAnomaly)
		anomalies.POST("/acknowledge", r.AnomalyHandler.AcknowledgeAnomaly)
	}

	// Time-series routes
	timeseries := engine.Group("/analytics/time-series")
	timeseries.Use(r.AuthMiddleware.RequireAuth())
	{
		timeseries.POST("/query", r.TimeSeriesHandler.QueryTimeSeries)
	}

	// KPI routes
	kpi := engine.Group("/analytics/kpi")
	kpi.Use(r.AuthMiddleware.RequireAuth())
	{
		kpi.GET("", r.KPIHandler.GetKPIs)
		kpi.GET("/:buildingId", r.KPIHandler.GetKPIs)
		kpi.POST("/calculate", r.KPIHandler.CalculateKPIs)
	}

	// Dashboard routes
	dashboards := engine.Group("/analytics/dashboards")
	dashboards.Use(r.AuthMiddleware.RequireAuth())
	{
		dashboards.GET("/overview", r.DashboardHandler.GetOverviewDashboard)
		dashboards.GET("/building/:buildingId", r.DashboardHandler.GetBuildingDashboard)
	}
}
