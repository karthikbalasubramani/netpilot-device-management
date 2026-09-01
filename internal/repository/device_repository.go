package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/device"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoDeviceRepository struct {
	collection *mongo.Collection
}

var _ device.Repository = (*mongoDeviceRepository)(nil)

// NewDeviceRepository creates a MongoDB-backed Device repository.
func NewDeviceRepository(
	collection *mongo.Collection,
) device.Repository {
	return &mongoDeviceRepository{
		collection: collection,
	}
}

// Create persists a new Device.
//
// MongoDB's unique device_id index is the final authority for duplicate
// device IDs. Duplicate-key errors are translated into a domain error
// so callers do not depend on MongoDB-specific error types.
func (repository *mongoDeviceRepository) Create(
	ctx context.Context, deviceModel *device.Device,
) error {
	_, err := repository.collection.InsertOne(
		ctx,
		deviceModel,
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return device.ErrAlreadyExists
		}
		return fmt.Errorf(
			"Create Device: %w", err,
		)
	}
	return nil
}

// GetByDeviceID retrieves a Device using NetPilot's application-level
// device_id rather than MongoDB's internal _id.
func (repository *mongoDeviceRepository) GetByDeviceID(
	ctx context.Context, deviceID string,
) (*device.Device, error) {
	var deviceModel device.Device

	err := repository.collection.FindOne(
		ctx,
		bson.M{
			"device_id": deviceID,
		},
	).Decode(&deviceModel)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, device.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf(
			"Get Device By ID: %w", err,
		)
	}
	return &deviceModel, nil
}
