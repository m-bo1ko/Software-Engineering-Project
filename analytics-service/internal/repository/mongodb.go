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

	"analytics-service/internal/config"
)

// MongoDB holds the database connection and collections
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	config   *config.Config
}

// Collections holds references to all MongoDB collections
type Collections struct {
	Reports    *mongo.Collection
	Anomalies  *mongo.Collection
	TimeSeries *mongo.Collection
	KPIs       *mongo.Collection
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
		Reports:    m.Database.Collection("reports"),
		Anomalies:  m.Database.Collection("anomalies"),
		TimeSeries: m.Database.Collection("time_series"),
		KPIs:       m.Database.Collection("kpis"),
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

	// Reports collection indexes
	reportIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"report_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"building_id": 1, "generated_at": -1},
		},
		{
			Keys: map[string]interface{}{"type": 1, "status": 1},
		},
		{
			Keys:    map[string]interface{}{"generated_at": 1},
			Options: options.Index().SetExpireAfterSeconds(7776000), // 90 days TTL
		},
	}
	if _, err := collections.Reports.Indexes().CreateMany(ctx, reportIndexes); err != nil {
		return fmt.Errorf("failed to create report indexes: %w", err)
	}

	// Anomalies collection indexes
	anomalyIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"anomaly_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"device_id": 1, "detected_at": -1},
		},
		{
			Keys: map[string]interface{}{"building_id": 1, "status": 1, "detected_at": -1},
		},
		{
			Keys: map[string]interface{}{"status": 1, "severity": 1},
		},
	}
	if _, err := collections.Anomalies.Indexes().CreateMany(ctx, anomalyIndexes); err != nil {
		return fmt.Errorf("failed to create anomaly indexes: %w", err)
	}

	// Time series collection indexes
	timeSeriesIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"device_id": 1, "timestamp": -1},
		},
		{
			Keys: map[string]interface{}{"building_id": 1, "timestamp": -1},
		},
		{
			Keys: map[string]interface{}{"aggregation_type": 1, "timestamp": -1},
		},
		{
			Keys:    map[string]interface{}{"timestamp": 1},
			Options: options.Index().SetExpireAfterSeconds(31536000), // 1 year TTL
		},
	}
	if _, err := collections.TimeSeries.Indexes().CreateMany(ctx, timeSeriesIndexes); err != nil {
		return fmt.Errorf("failed to create time series indexes: %w", err)
	}

	// KPIs collection indexes
	kpiIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"building_id": 1, "calculated_at": -1},
		},
		{
			Keys: map[string]interface{}{"period": 1, "calculated_at": -1},
		},
	}
	if _, err := collections.KPIs.Indexes().CreateMany(ctx, kpiIndexes); err != nil {
		return fmt.Errorf("failed to create KPI indexes: %w", err)
	}

	log.Println("MongoDB indexes created successfully")
	return nil
}
