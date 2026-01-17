// Package main is the entry point for the Analytics service
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

	"analytics-service/internal/config"
	"analytics-service/internal/handlers"
	"analytics-service/internal/integrations"
	"analytics-service/internal/middleware"
	"analytics-service/internal/repository"
	"analytics-service/internal/service"
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
	reportRepo := repository.NewReportRepository(collections.Reports)
	anomalyRepo := repository.NewAnomalyRepository(collections.Anomalies)
	timeSeriesRepo := repository.NewTimeSeriesRepository(collections.TimeSeries)
	kpiRepo := repository.NewKPIRepository(collections.KPIs)

	// Initialize external integrations
	securityClient := integrations.NewSecurityClient(cfg)
	iotClient := integrations.NewIoTClient(cfg)
	forecastClient := integrations.NewForecastClient(cfg)

	// Initialize services
	reportService := service.NewReportService(reportRepo, iotClient, forecastClient)
	anomalyService := service.NewAnomalyService(anomalyRepo, iotClient)
	timeSeriesService := service.NewTimeSeriesService(timeSeriesRepo, iotClient)
	kpiService := service.NewKPIService(kpiRepo, anomalyRepo, iotClient)
	dashboardService := service.NewDashboardService(anomalyRepo, kpiRepo, iotClient, forecastClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(securityClient)

	// Initialize handlers
	reportHandler := handlers.NewReportHandler(reportService, securityClient)
	anomalyHandler := handlers.NewAnomalyHandler(anomalyService, securityClient)
	timeSeriesHandler := handlers.NewTimeSeriesHandler(timeSeriesService)
	kpiHandler := handlers.NewKPIHandler(kpiService, securityClient)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	// Create router
	router := handlers.NewRouter(
		reportHandler,
		anomalyHandler,
		timeSeriesHandler,
		kpiHandler,
		dashboardHandler,
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
		log.Printf("Starting Analytics Service on %s:%s", cfg.Server.Host, cfg.Server.Port)
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
