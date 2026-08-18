package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines password hashing operations required by the
// authentication service.
type PasswordHasher interface {
	Hash(
		password string,
	) (string, error)

	Verify(
		passwordHash string,
		password string,
	) error
}

var ErrPasswordMismatch = errors.New(
	"Password does not match",
)

// bcryptPasswordHasher implements PasswordHasher using bcrypt.
type bcryptPasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a bcrypt-backed password hasher.
func NewPasswordHasher(
	cost int,
) (PasswordHasher, error) {
	if cost < bcrypt.MinCost ||
		cost > bcrypt.MaxCost {
		return nil, fmt.Errorf(
			"bcrypt cost must be between %d and %d",
			bcrypt.MinCost,
			bcrypt.MaxCost,
		)
	}

	return &bcryptPasswordHasher{
		cost: cost,
	}, nil
}

// Hash creates a bcrypt hash from the provided plain-text password.
func (hasher *bcryptPasswordHasher) Hash(
	password string,
) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		hasher.cost,
	)
	if err != nil {
		return "", fmt.Errorf(
			"hash password: %w",
			err,
		)
	}

	return string(hashedPassword), nil
}

// Verify compares a stored bcrypt hash against a plain-text password.
func (hasher *bcryptPasswordHasher) Verify(
	passwordHash string,
	password string,
) error {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
	if err != nil {
		if errors.Is(
			err,
			bcrypt.ErrMismatchedHashAndPassword,
		) {
			return ErrPasswordMismatch
		}
		return fmt.Errorf(
			"Verify password hash: %w",
			err,
		)
	}
	return nil
}
