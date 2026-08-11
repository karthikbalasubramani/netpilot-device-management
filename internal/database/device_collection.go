package database

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DeviceCollectionName = "devices"
	deviceIDIndexName    = "ux_devices_device_id"
)

// Device Collection returns MongoDB devices collection
func DeviceCollection(client *mongo.Client, databaseName string) *mongo.Collection {
	return client.Database(databaseName).Collection(DeviceCollectionName)
}

// EnsureDeviceCollection: Collection will be created
// with index rule if its not exist while starting the application
func EnsureDeviceCollection(ctx context.Context, client *mongo.Client, databaseName string) error {
	if client == nil {
		return fmt.Errorf("MongoDB client is required, received nil")
	}

	databaseName = strings.TrimSpace(databaseName)
	if databaseName == "" {
		return fmt.Errorf("MongoDB databaseName can't be empty, value is invalid")
	}

	collection := DeviceCollection(
		client,
		databaseName,
	)

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{
				Key:   "device_id",
				Value: 1,
			},
		},
		Options: options.Index().SetName(deviceIDIndexName).SetUnique(true),
	}

	if _, err := collection.Indexes().CreateOne(
		ctx,
		indexModel,
	); err != nil {
		return fmt.Errorf(
			"Create devices device_id failed: %w", err,
		)
	}
	return nil
}
