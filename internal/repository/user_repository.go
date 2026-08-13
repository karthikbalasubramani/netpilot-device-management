package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

// mongoUserRepository implements user.Repository using MongoDB.
type mongoUserRepository struct {
	collection *mongo.Collection
}

// Compile-time verification that mongoUserRepository satisfies
// the user.Repository interface.
var _ user.Repository = (*mongoUserRepository)(nil)

// NewUserRepository creates a MongoDB-backed implementation of
// user.Repository.
func NewUserRepository(
	collection *mongo.Collection,
) (user.Repository, error) {
	if collection == nil {
		return nil, fmt.Errorf(
			"user collection is required",
		)
	}

	return &mongoUserRepository{
		collection: collection,
	}, nil
}

// Create persists a new NetPilot user in MongoDB.
func (repository *mongoUserRepository) Create(
	ctx context.Context,
	userToCreate *user.User,
) error {
	if userToCreate == nil {
		return fmt.Errorf(
			"user to create is required",
		)
	}

	result, err := repository.collection.InsertOne(
		ctx,
		userToCreate,
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf(
				"%w: user_id %q or email %q",
				user.ErrAlreadyExists,
				userToCreate.UserID,
				userToCreate.Email,
			)
		}

		return fmt.Errorf(
			"insert user %q: %w",
			userToCreate.UserID,
			err,
		)
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		userToCreate.ID = objectID
	}

	return nil
}

// GetByEmail retrieves a user by normalized email address.
func (repository *mongoUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {
	normalizedEmail := strings.ToLower(
		strings.TrimSpace(email),
	)

	if normalizedEmail == "" {
		return nil, fmt.Errorf(
			"user email is required",
		)
	}

	filter := bson.D{
		{
			Key:   "email",
			Value: normalizedEmail,
		},
	}

	var storedUser user.User

	err := repository.collection.
		FindOne(ctx, filter).
		Decode(&storedUser)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(
				"%w: email %q",
				user.ErrNotFound,
				normalizedEmail,
			)
		}

		return nil, fmt.Errorf(
			"find user by email %q: %w",
			normalizedEmail,
			err,
		)
	}

	return &storedUser, nil
}

// GetByUserID retrieves a user by the NetPilot application-level user ID.
func (repository *mongoUserRepository) GetByUserID(
	ctx context.Context,
	userID string,
) (*user.User, error) {
	normalizedUserID := strings.TrimSpace(userID)

	if normalizedUserID == "" {
		return nil, fmt.Errorf(
			"user ID is required",
		)
	}

	filter := bson.D{
		{
			Key:   "user_id",
			Value: normalizedUserID,
		},
	}

	var storedUser user.User

	err := repository.collection.
		FindOne(ctx, filter).
		Decode(&storedUser)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf(
				"%w: user_id %q",
				user.ErrNotFound,
				normalizedUserID,
			)
		}

		return nil, fmt.Errorf(
			"find user by user_id %q: %w",
			normalizedUserID,
			err,
		)
	}

	return &storedUser, nil
}
