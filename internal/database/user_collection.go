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
	UserCollectionName = "users"

	userIDIndexName = "ux_users_user_id"
	emailIndexName  = "ux_users_email"
)

// UserCollection returns the MongoDB collection used for NetPilot user
// accounts.
func UserCollection(
	client *mongo.Client,
	databaseName string,
) *mongo.Collection {
	return client.
		Database(databaseName).
		Collection(UserCollectionName)
}

// EnsureUserCollection prepares the users collection and its required indexes.
//
// MongoDB creates the collection automatically when an index is created if the
// collection does not already exist.
func EnsureUserCollection(
	ctx context.Context,
	client *mongo.Client,
	databaseName string,
) error {
	if client == nil {
		return fmt.Errorf("MongoDB client is required")
	}

	databaseName = strings.TrimSpace(databaseName)
	if databaseName == "" {
		return fmt.Errorf("MongoDB database name is required")
	}

	collection := UserCollection(
		client,
		databaseName,
	)

	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   "user_id",
					Value: 1,
				},
			},
			Options: options.Index().
				SetName(userIDIndexName).
				SetUnique(true),
		},
		{
			Keys: bson.D{
				{
					Key:   "email",
					Value: 1,
				},
			},
			Options: options.Index().
				SetName(emailIndexName).
				SetUnique(true),
		},
	}

	if _, err := collection.Indexes().CreateMany(
		ctx,
		indexModels,
	); err != nil {
		return fmt.Errorf(
			"create users collection indexes: %w",
			err,
		)
	}

	return nil
}
