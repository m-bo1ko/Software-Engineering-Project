package handlers

import (
	"github.com/gin-gonic/gin"

	"security-service/internal/middleware"
)

// Router holds all handler dependencies
type Router struct {
	AuthHandler         *AuthHandler
	UserHandler         *UserHandler
	RoleHandler         *RoleHandler
	AuditHandler        *AuditHandler
	NotificationHandler *NotificationHandler
	EnergyHandler       *EnergyHandler
	AuthMiddleware      *middleware.AuthMiddleware
}

// NewRouter creates a new router with all handlers
func NewRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	roleHandler *RoleHandler,
	auditHandler *AuditHandler,
	notificationHandler *NotificationHandler,
	energyHandler *EnergyHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		RoleHandler:         roleHandler,
		AuditHandler:        auditHandler,
		NotificationHandler: notificationHandler,
		EnergyHandler:       energyHandler,
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
			"service": "security-service",
		})
	})

	// API v1 routes
	api := engine.Group("/api/v1")
	{
		r.setupAuthRoutes(api)
		r.setupUserRoutes(api)
		r.setupRoleRoutes(api)
		r.setupAuditRoutes(api)
		r.setupNotificationRoutes(api)
		r.setupEnergyRoutes(api)
	}

	// Legacy routes (without /api/v1 prefix for backward compatibility)
	r.setupLegacyRoutes(engine)
}

// setupAuthRoutes configures authentication routes
func (r *Router) setupAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		// Public routes
		auth.POST("/login", r.AuthHandler.Login)
		auth.POST("/refresh", r.AuthHandler.RefreshToken)

		// Token validation (for internal microservices)
		auth.GET("/validate-token", r.AuthHandler.ValidateToken)

		// Permission check (for internal microservices)
		auth.POST("/check-permissions", r.AuthHandler.CheckPermissions)

		// Protected routes
		protected := auth.Group("")
		protected.Use(r.AuthMiddleware.RequireAuth())
		{
			protected.POST("/logout", r.AuthHandler.Logout)
			protected.GET("/user-info", r.AuthHandler.GetUserInfo)
		}
	}
}

// setupUserRoutes configures user management routes
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.Use(r.AuthMiddleware.RequireAuth())
	{
		// Admin only routes
		users.GET("", r.AuthMiddleware.RequireAdmin(), r.UserHandler.ListUsers)
		users.POST("", r.AuthMiddleware.RequireAdmin(), r.UserHandler.CreateUser)
		users.DELETE("/:id", r.AuthMiddleware.RequireAdmin(), r.UserHandler.DeleteUser)

		// Protected routes (user can view their own details or admin can view any)
		users.GET("/:id", r.UserHandler.GetUser)
		users.PUT("/:id", r.UserHandler.UpdateUser)
	}
}

// setupRoleRoutes configures role management routes
func (r *Router) setupRoleRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	roles.Use(r.AuthMiddleware.RequireAuth())
	{
		// Admin only routes
		roles.GET("", r.RoleHandler.ListRoles)
		roles.POST("", r.AuthMiddleware.RequireAdmin(), r.RoleHandler.CreateRole)
		roles.PUT("/:roleName", r.AuthMiddleware.RequireAdmin(), r.RoleHandler.UpdateRole)
		roles.DELETE("/:roleName", r.AuthMiddleware.RequireAdmin(), r.RoleHandler.DeleteRole)
	}
}

// setupAuditRoutes configures audit logging routes
func (r *Router) setupAuditRoutes(rg *gin.RouterGroup) {
	audit := rg.Group("/audit")
	{
		// Allow internal services to log without full auth
		audit.POST("/log", r.AuditHandler.CreateLog)

		// Protected routes for viewing logs
		protected := audit.Group("")
		protected.Use(r.AuthMiddleware.RequireAuth())
		protected.Use(r.AuthMiddleware.RequireAdmin())
		{
			protected.GET("/logs", r.AuditHandler.GetLogs)
			protected.GET("/logs/:id", r.AuditHandler.GetLog)
		}
	}
}

// setupNotificationRoutes configures notification routes
func (r *Router) setupNotificationRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	notifications.Use(r.AuthMiddleware.RequireAuth())
	{
		notifications.POST("/send", r.NotificationHandler.SendNotification)
		notifications.POST("/preferences", r.NotificationHandler.UpdatePreferences)
		notifications.GET("/preferences/:userId", r.NotificationHandler.GetPreferences)
		notifications.PUT("/preferences/:userId", r.NotificationHandler.UpdatePreferencesByUserID)
		notifications.GET("/logs", r.NotificationHandler.GetLogs)
	}
}

// setupEnergyRoutes configures external energy provider routes
func (r *Router) setupEnergyRoutes(rg *gin.RouterGroup) {
	energy := rg.Group("/external-energy")
	energy.Use(r.AuthMiddleware.RequireAuth())
	{
		energy.GET("/consumption", r.EnergyHandler.GetConsumption)
		energy.GET("/tariffs", r.EnergyHandler.GetTariffs)
		energy.POST("/refresh-token", r.AuthMiddleware.RequireAdmin(), r.EnergyHandler.RefreshToken)
	}
}

// setupLegacyRoutes configures legacy routes without /api/v1 prefix
func (r *Router) setupLegacyRoutes(engine *gin.Engine) {
	// Auth routes
	auth := engine.Group("/auth")
	{
		auth.POST("/login", r.AuthHandler.Login)
		auth.POST("/refresh", r.AuthHandler.RefreshToken)
		auth.GET("/validate-token", r.AuthHandler.ValidateToken)
		auth.POST("/check-permissions", r.AuthHandler.CheckPermissions)

		protected := auth.Group("")
		protected.Use(r.AuthMiddleware.RequireAuth())
		{
			protected.POST("/logout", r.AuthHandler.Logout)
			protected.GET("/user-info", r.AuthHandler.GetUserInfo)
		}
	}

	// User routes
	users := engine.Group("/users")
	users.Use(r.AuthMiddleware.RequireAuth())
	{
		users.GET("", r.AuthMiddleware.RequireAdmin(), r.UserHandler.ListUsers)
		users.POST("", r.AuthMiddleware.RequireAdmin(), r.UserHandler.CreateUser)
		users.GET("/:id", r.UserHandler.GetUser)
		users.PUT("/:id", r.UserHandler.UpdateUser)
		users.DELETE("/:id", r.AuthMiddleware.RequireAdmin(), r.UserHandler.DeleteUser)
	}

	// Role routes
	roles := engine.Group("/roles")
	roles.Use(r.AuthMiddleware.RequireAuth())
	{
		roles.GET("", r.RoleHandler.ListRoles)
		roles.POST("", r.AuthMiddleware.RequireAdmin(), r.RoleHandler.CreateRole)
		roles.PUT("/:roleName", r.AuthMiddleware.RequireAdmin(), r.RoleHandler.UpdateRole)
		roles.DELETE("/:roleName", r.AuthMiddleware.RequireAdmin(), r.RoleHandler.DeleteRole)
	}

	// Notification routes
	notifications := engine.Group("/notifications")
	notifications.Use(r.AuthMiddleware.RequireAuth())
	{
		notifications.POST("/send", r.NotificationHandler.SendNotification)
		notifications.POST("/preferences", r.NotificationHandler.UpdatePreferences)
		notifications.GET("/preferences/:userId", r.NotificationHandler.GetPreferences)
		notifications.PUT("/preferences/:userId", r.NotificationHandler.UpdatePreferencesByUserID)
		notifications.GET("/logs", r.NotificationHandler.GetLogs)
	}

	// Audit routes
	audit := engine.Group("/audit")
	{
		audit.POST("/log", r.AuditHandler.CreateLog)

		protected := audit.Group("")
		protected.Use(r.AuthMiddleware.RequireAuth())
		protected.Use(r.AuthMiddleware.RequireAdmin())
		{
			protected.GET("/logs", r.AuditHandler.GetLogs)
		}
	}

	// External energy routes
	energy := engine.Group("/external-energy")
	energy.Use(r.AuthMiddleware.RequireAuth())
	{
		energy.GET("/consumption", r.EnergyHandler.GetConsumption)
		energy.GET("/tariffs", r.EnergyHandler.GetTariffs)
		energy.POST("/refresh-token", r.AuthMiddleware.RequireAdmin(), r.EnergyHandler.RefreshToken)
	}
}
