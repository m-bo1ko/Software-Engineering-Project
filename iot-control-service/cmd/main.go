// Package main is the entry point for the IoT Control service
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

	"iot-control-service/internal/config"
	"iot-control-service/internal/handlers"
	"iot-control-service/internal/integrations"
	"iot-control-service/internal/middleware"
	"iot-control-service/internal/models"
	"iot-control-service/internal/mqtt"
	"iot-control-service/internal/repository"
	"iot-control-service/internal/service"
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
	deviceRepo := repository.NewDeviceRepository(collections.Devices)
	telemetryRepo := repository.NewTelemetryRepository(collections.Telemetry)
	commandRepo := repository.NewCommandRepository(collections.DeviceCommands)
	optimizationRepo := repository.NewOptimizationRepository(collections.OptimizationScenarios)

	// Initialize external integrations
	securityClient := integrations.NewSecurityClient(cfg)
	// Integration: ForecastClient enables fetching device predictions for optimization timing
	forecastClient := integrations.NewForecastClient(cfg)
	// Integration: AnalyticsClient enables checking anomalies before applying optimizations
	analyticsClient := integrations.NewAnalyticsClient(cfg)

	// Initialize MQTT client
	mqttClient, err := mqtt.NewClient(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to MQTT broker: %v", err)
	} else {
		defer mqttClient.Disconnect()
		// Subscribe to MQTT telemetry and acks
		setupMQTTSubscriptions(mqttClient, telemetryRepo, deviceRepo, commandRepo)
	}

	// Initialize services
	deviceService := service.NewDeviceService(deviceRepo)
	telemetryService := service.NewTelemetryService(telemetryRepo, deviceRepo)
	controlService := service.NewControlService(commandRepo, deviceRepo, mqttClient, cfg.IoT.CommandTimeout)
	// Integration: OptimizationService now uses ForecastClient and AnalyticsClient
	// to fetch predictions and check anomalies before executing optimization scenarios
	optimizationService := service.NewOptimizationService(optimizationRepo, commandRepo, deviceRepo, forecastClient, analyticsClient)
	stateService := service.NewStateService(deviceRepo, telemetryRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(securityClient)

	// Initialize handlers
	deviceHandler := handlers.NewDeviceHandler(deviceService, securityClient)
	telemetryHandler := handlers.NewTelemetryHandler(telemetryService, securityClient)
	controlHandler := handlers.NewControlHandler(controlService, securityClient)
	optimizationHandler := handlers.NewOptimizationHandler(optimizationService, securityClient)
	stateHandler := handlers.NewStateHandler(stateService)

	// Create router
	router := handlers.NewRouter(
		deviceHandler,
		telemetryHandler,
		controlHandler,
		optimizationHandler,
		stateHandler,
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
		log.Printf("Starting IoT Control Service on %s:%s", cfg.Server.Host, cfg.Server.Port)
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

// setupMQTTSubscriptions sets up MQTT subscriptions for telemetry and command acks
func setupMQTTSubscriptions(
	mqttClient *mqtt.Client,
	telemetryRepo *repository.TelemetryRepository,
	deviceRepo *repository.DeviceRepository,
	commandRepo *repository.CommandRepository,
) {
	// Subscribe to all telemetry
	mqttClient.SubscribeToAllTelemetry(func(deviceID string, telemetry *models.Telemetry) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		telemetry.Source = "MQTT"
		_, err := telemetryRepo.Create(ctx, telemetry)
		if err != nil {
			log.Printf("Failed to save MQTT telemetry: %v", err)
			return
		}

		// Update device last seen
		deviceRepo.UpdateLastSeen(ctx, deviceID)
	})

	// Subscribe to all command acks
	mqttClient.SubscribeToAllAcks(func(deviceID string, ack *models.CommandAck) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := commandRepo.FindByCommandID(ctx, ack.CommandID)
		if err != nil {
			log.Printf("Command not found for ack: %s", ack.CommandID)
			return
		}

		status := models.CommandStatusApplied
		if ack.Status == "FAILED" {
			status = models.CommandStatusFailed
		}

		commandRepo.UpdateStatus(ctx, ack.CommandID, status, ack.ErrorMsg)
	})
}
