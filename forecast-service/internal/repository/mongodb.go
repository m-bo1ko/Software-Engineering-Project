// Package repository handles database operations
package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"forecast-service/internal/config"
)

// MongoDB holds the database connection and collections
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	config   *config.Config
}

// Collections holds references to all MongoDB collections
type Collections struct {
	Forecasts             *mongo.Collection
	PeakLoads             *mongo.Collection
	OptimizationScenarios *mongo.Collection
	Recommendations       *mongo.Collection
	Devices               *mongo.Collection
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MongoDB.Timeout)
	defer cancel()

	// Create client options
	clientOptions := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("Connected to MongoDB: %s", cfg.MongoDB.Database)

	return &MongoDB{
		Client:   client,
		Database: client.Database(cfg.MongoDB.Database),
		config:   cfg,
	}, nil
}

// GetCollections returns all collection references
func (m *MongoDB) GetCollections() *Collections {
	return &Collections{
		Forecasts:             m.Database.Collection("forecasts"),
		PeakLoads:             m.Database.Collection("peak_loads"),
		OptimizationScenarios: m.Database.Collection("optimization_scenarios"),
		Recommendations:       m.Database.Collection("recommendations"),
		Devices:               m.Database.Collection("devices"),
	}
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	log.Println("Disconnected from MongoDB")
	return nil
}

// CreateIndexes creates necessary indexes for optimal query performance
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	collections := m.GetCollections()

	// Forecasts collection indexes
	forecastIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"building_id": 1, "created_at": -1},
		},
		{
			Keys: map[string]interface{}{"building_id": 1, "type": 1, "status": 1},
		},
		{
			Keys: map[string]interface{}{"device_id": 1, "created_at": -1},
		},
	}
	if _, err := collections.Forecasts.Indexes().CreateMany(ctx, forecastIndexes); err != nil {
		return fmt.Errorf("failed to create forecast indexes: %w", err)
	}

	// Peak loads collection indexes
	peakLoadIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"building_id": 1, "created_at": -1},
		},
		{
			Keys: map[string]interface{}{"forecast_id": 1},
		},
	}
	if _, err := collections.PeakLoads.Indexes().CreateMany(ctx, peakLoadIndexes); err != nil {
		return fmt.Errorf("failed to create peak load indexes: %w", err)
	}

	// Optimization scenarios collection indexes
	optimizationIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"building_id": 1, "status": 1, "created_at": -1},
		},
		{
			Keys: map[string]interface{}{"forecast_id": 1},
		},
		{
			Keys: map[string]interface{}{"scheduled_start": 1, "status": 1},
		},
	}
	if _, err := collections.OptimizationScenarios.Indexes().CreateMany(ctx, optimizationIndexes); err != nil {
		return fmt.Errorf("failed to create optimization scenario indexes: %w", err)
	}

	// Recommendations collection indexes
	recommendationIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"building_id": 1, "status": 1, "priority": 1},
		},
		{
			Keys: map[string]interface{}{"device_id": 1},
		},
	}
	if _, err := collections.Recommendations.Indexes().CreateMany(ctx, recommendationIndexes); err != nil {
		return fmt.Errorf("failed to create recommendation indexes: %w", err)
	}

	// Devices collection indexes
	deviceIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"device_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"building_id": 1},
		},
	}
	if _, err := collections.Devices.Indexes().CreateMany(ctx, deviceIndexes); err != nil {
		return fmt.Errorf("failed to create device indexes: %w", err)
	}

	log.Println("MongoDB indexes created successfully")
	return nil
}
