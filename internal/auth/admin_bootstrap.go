package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

var ErrAdminAlreadyBootstrapped = errors.New(
	"An administrator already exists",
)

type AdminBootstrapRequest struct {
	Name     string
	Email    string
	Password string
}

type AdminBootstrapper struct {
	userRepository user.Repository
	passwordHasher PasswordHasher
}

func NewAdminBootstrapper(
	userRepository user.Repository,
	passwordHasher PasswordHasher,
) *AdminBootstrapper {
	return &AdminBootstrapper{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
	}
}

func (bootstrapper *AdminBootstrapper) Bootstrap(
	ctx context.Context,
	request AdminBootstrapRequest,
) (*user.User, error) {
	adminCount, err :=
		bootstrapper.userRepository.CountByRole(
			ctx,
			user.Role("admin"),
		)
	if err != nil {
		return nil, fmt.Errorf(
			"check existing administrator: %w",
			err,
		)
	}

	if adminCount > 0 {
		return nil, ErrAdminAlreadyBootstrapped
	}

	name, err := validateUserName(
		request.Name,
	)
	if err != nil {
		return nil, err
	}

	email, err := normalizeAndValidateEmail(
		request.Email,
	)
	if err != nil {
		return nil, err
	}

	if err := validatePassword(
		request.Password,
	); err != nil {
		return nil, err
	}

	passwordHash, err :=
		bootstrapper.passwordHasher.Hash(
			request.Password,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"hash administrator password: %w",
			err,
		)
	}

	now := time.Now().UTC()

	adminUser := &user.User{
		UserID: "usr_" + uuid.NewString(),

		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,

		Role:   user.RoleAdmin,
		Status: user.StatusActive,

		CreatedAt: now,
		UpdatedAt: now,
	}

	err = bootstrapper.userRepository.Create(
		ctx,
		adminUser,
	)
	if err != nil {
		if errors.Is(
			err,
			user.ErrAlreadyExists,
		) {
			return nil, user.ErrAlreadyExists
		}

		return nil, fmt.Errorf(
			"create bootstrap administrator: %w",
			err,
		)
	}

	return adminUser, nil
}
