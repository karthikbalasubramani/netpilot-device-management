package database

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var errMongoDBClientNotInitialized = errors.New(
	"MongoDB client is not initialized",
)

// CheckMongoDBReadiness verifies that the MongoDB deployment is reachable.
func CheckMongoDBReadiness(
	ctx context.Context,
	client *mongo.Client,
) error {
	if client == nil {
		return errMongoDBClientNotInitialized
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf(
			"ping MongoDB: %w",
			err,
		)
	}
	return nil
}
