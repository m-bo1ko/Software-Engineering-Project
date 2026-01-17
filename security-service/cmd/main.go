// Package main is the entry point for the security service
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"security-service/internal/config"
	"security-service/internal/handlers"
	"security-service/internal/integrations"
	"security-service/internal/middleware"
	"security-service/internal/repository"
	"security-service/internal/service"
	"security-service/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoDB, err := repository.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := mongoDB.Close(shutdownCtx); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
	}()

	// Create indexes
	if err := mongoDB.CreateIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Get collections
	collections := mongoDB.GetCollections()

	// Initialize repositories
	userRepo := repository.NewUserRepository(collections.Users)
	roleRepo := repository.NewRoleRepository(collections.Roles)
	authRepo := repository.NewAuthRepository(collections.RefreshTokens, collections.AuthCredentials)
	auditRepo := repository.NewAuditRepository(collections.AuditLogs)
	notificationRepo := repository.NewNotificationRepository(collections.Notifications, collections.NotificationPrefs)

	// Initialize default roles
	roleService := service.NewRoleService(roleRepo, auditRepo)
	if err := roleService.InitializeDefaultRoles(ctx); err != nil {
		log.Printf("Warning: Failed to initialize default roles: %v", err)
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Initialize services
	authService := service.NewAuthService(userRepo, roleRepo, authRepo, auditRepo, jwtManager)
	userService := service.NewUserService(userRepo, roleRepo, auditRepo)
	auditService := service.NewAuditService(auditRepo)

	// Initialize external integrations
	notificationClient := integrations.NewNotificationClient(cfg)
	energyClient, err := integrations.NewEnergyProviderClient(cfg, authRepo)
	if err != nil {
		log.Printf("Warning: Failed to initialize energy client: %v", err)
	}

	notificationService := service.NewNotificationService(notificationRepo, notificationClient)

	// Initialize default admin user
	if err := userService.InitializeAdminUser(ctx); err != nil {
		log.Printf("Warning: Failed to initialize admin user: %v", err)
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	roleHandler := handlers.NewRoleHandler(roleService)
	auditHandler := handlers.NewAuditHandler(auditService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	energyHandler := handlers.NewEnergyHandler(energyClient)

	// Create router
	router := handlers.NewRouter(
		authHandler,
		userHandler,
		roleHandler,
		auditHandler,
		notificationHandler,
		energyHandler,
		authMiddleware,
	)

	// Create Gin engine and setup routes
	engine := gin.New()
	router.SetupRoutes(engine)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting Security Service on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
