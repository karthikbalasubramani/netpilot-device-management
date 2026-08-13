package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines password hashing operations required by the
// authentication service.
type PasswordHasher interface {
	Hash(password string) (string, error)
}

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
