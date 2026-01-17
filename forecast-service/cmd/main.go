// Package main is the entry point for the forecast service
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

	"forecast-service/internal/config"
	"forecast-service/internal/handlers"
	"forecast-service/internal/integrations"
	"forecast-service/internal/middleware"
	"forecast-service/internal/repository"
	"forecast-service/internal/service"
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
	forecastRepo := repository.NewForecastRepository(collections.Forecasts)
	peakLoadRepo := repository.NewPeakLoadRepository(collections.PeakLoads)
	optimizationRepo := repository.NewOptimizationRepository(collections.OptimizationScenarios)
	recommendationRepo := repository.NewRecommendationRepository(collections.Recommendations)

	// Initialize external integrations
	securityClient := integrations.NewSecurityClient(cfg)
	externalClient := integrations.NewExternalClient(cfg)
	iotClient := integrations.NewIoTClient(cfg)

	// Initialize services
	forecastService := service.NewForecastService(
		forecastRepo,
		peakLoadRepo,
		securityClient,
		externalClient,
		cfg,
	)

	optimizationService := service.NewOptimizationService(
		optimizationRepo,
		forecastRepo,
		recommendationRepo,
		iotClient,
		externalClient,
		securityClient,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(securityClient)

	// Initialize handlers
	forecastHandler := handlers.NewForecastHandler(forecastService, securityClient)
	optimizationHandler := handlers.NewOptimizationHandler(optimizationService, securityClient)

	// Create router
	router := handlers.NewRouter(
		forecastHandler,
		optimizationHandler,
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
		log.Printf("Starting Forecast Service on %s:%s", cfg.Server.Host, cfg.Server.Port)
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

