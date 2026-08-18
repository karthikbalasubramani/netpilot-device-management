package user

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when the requested user does not exist.
	ErrNotFound = errors.New("user not found")

	// ErrAlreadyExists is returned when a user conflicts with an existing
	// unique user identifier or email address.
	ErrAlreadyExists = errors.New("user already exists")
)

// Repository defines the persistence operations required by the User domain.
//
// Implementations can use MongoDB or another persistence technology without
// exposing database-specific details to the authentication/business layer.
type Repository interface {
	Create(
		ctx context.Context,
		user *User,
	) error

	GetByEmail(
		ctx context.Context,
		email string,
	) (*User, error)

	GetByUserID(
		ctx context.Context,
		userID string,
	) (*User, error)

	UpdateLastLogin(
		ctx context.Context,
		userID string,
		lastLoginAt time.Time,
	) error
}
