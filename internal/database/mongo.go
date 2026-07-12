package database

import (
	"context"
	"fmt"
	"time"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/config"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the MongoDB client and database instance used by the application.
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// ConnectMongoDB creates a MongoDB client using the configured Mongo URI,
// verifies the connection using Ping, and returns the MongoDB wrapper.
func ConnectMongoDB(cfg *config.Config) (*MongoDB, error) {
	// Create a context with timeout to avoid hanging during MongoDB connection.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("connecting to MongoDB", "database", cfg.MongoDatabase)

	// Prepare MongoDB client options using the configured connection URI.
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	// Create MongoDB client.
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Verify MongoDB connection by sending a ping request.
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB client: %w", err)
	}

	// Return MongoDB wrapper with client and selected database reference.
	return &MongoDB{
		Client:   client,
		Database: client.Database(cfg.MongoDatabase),
	}, nil
}

// Disconnect closes the MongoDB client connection gracefully.
func Disconnect(m *MongoDB) error {
	if m == nil || m.Client == nil {
		return nil
	}

	// Create a context with timeout to avoid hanging during MongoDB disconnect.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Disconnect MongoDB client.
	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect MongoDB: %w", err)
	}

	return nil
}
