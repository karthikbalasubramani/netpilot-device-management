package auth

import (
	"fmt"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

// Service contains authentication and account-management business logic.
type Service struct {
	userRepository user.Repository
	passwordHasher PasswordHasher
	tokenIssuer    AccessTokenIssuer
}

// NewService creates an authentication service.
func NewService(
	userRepository user.Repository,
	passwordHasher PasswordHasher,
	tokenIssuer AccessTokenIssuer,
) (*Service, error) {
	if userRepository == nil {
		return nil, fmt.Errorf(
			"user repository is required",
		)
	}

	if passwordHasher == nil {
		return nil, fmt.Errorf(
			"password hasher is required",
		)
	}

	if tokenIssuer == nil {
		return nil, fmt.Errorf(
			"access token issuer is required",
		)
	}

	return &Service{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		tokenIssuer:    tokenIssuer,
	}, nil
}
