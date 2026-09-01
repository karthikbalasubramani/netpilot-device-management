package device

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("device not found")
	ErrAlreadyExists = errors.New("device already exists")
)

// Repository defines the persistence operations required by the
// Device domain.
//
// Implementations may use MongoDB or another persistence mechanism,
// while callers depend only on this domain-level contract.
type Repository interface {
	Create(ctx context.Context, device *Device) error
	GetByDeviceID(ctx context.Context, deviceID string) (*Device, error)
}
