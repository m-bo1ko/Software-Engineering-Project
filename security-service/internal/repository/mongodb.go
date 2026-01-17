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

	"security-service/internal/config"
)

// MongoDB holds the database connection and collections
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	config   *config.Config
}

// Collections holds references to all MongoDB collections
type Collections struct {
	Users              *mongo.Collection
	Roles              *mongo.Collection
	AuthCredentials    *mongo.Collection
	AuditLogs          *mongo.Collection
	RefreshTokens      *mongo.Collection
	Notifications      *mongo.Collection
	NotificationPrefs  *mongo.Collection
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
		Users:              m.Database.Collection("users"),
		Roles:              m.Database.Collection("roles"),
		AuthCredentials:    m.Database.Collection("auth_credentials"),
		AuditLogs:          m.Database.Collection("audit_logs"),
		RefreshTokens:      m.Database.Collection("refresh_tokens"),
		Notifications:      m.Database.Collection("notifications"),
		NotificationPrefs:  m.Database.Collection("notification_preferences"),
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

	// Users collection indexes
	userIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"username": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := collections.Users.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Roles collection indexes
	roleIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"name": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := collections.Roles.Indexes().CreateMany(ctx, roleIndexes); err != nil {
		return fmt.Errorf("failed to create role indexes: %w", err)
	}

	// Refresh tokens indexes
	refreshTokenIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"token": 1},
		},
		{
			Keys: map[string]interface{}{"user_id": 1},
		},
		{
			Keys:    map[string]interface{}{"expires_at": 1},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
	}
	if _, err := collections.RefreshTokens.Indexes().CreateMany(ctx, refreshTokenIndexes); err != nil {
		return fmt.Errorf("failed to create refresh token indexes: %w", err)
	}

	// Audit logs indexes
	auditIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "timestamp": -1},
		},
		{
			Keys: map[string]interface{}{"service": 1, "timestamp": -1},
		},
		{
			Keys: map[string]interface{}{"timestamp": -1},
		},
	}
	if _, err := collections.AuditLogs.Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return fmt.Errorf("failed to create audit log indexes: %w", err)
	}

	// Notifications indexes
	notificationIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "created_at": -1},
		},
	}
	if _, err := collections.Notifications.Indexes().CreateMany(ctx, notificationIndexes); err != nil {
		return fmt.Errorf("failed to create notification indexes: %w", err)
	}

	// Auth credentials indexes
	authCredIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"service_name": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := collections.AuthCredentials.Indexes().CreateMany(ctx, authCredIndexes); err != nil {
		return fmt.Errorf("failed to create auth credentials indexes: %w", err)
	}

	log.Println("MongoDB indexes created successfully")
	return nil
}
