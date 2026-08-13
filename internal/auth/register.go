package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

const (
	minimumUserNameLength = 2
	maximumUserNameLength = 100

	minimumPasswordLength = 12
	maximumPasswordBytes  = 72

	maximumEmailLength = 254
)

var (
	ErrInvalidName = errors.New(
		"name must contain between 2 and 100 characters",
	)

	ErrInvalidEmail = errors.New(
		"email address is invalid",
	)

	ErrPasswordTooShort = errors.New(
		"password must contain at least 12 characters",
	)

	ErrPasswordTooLong = errors.New(
		"password must not exceed 72 bytes",
	)
)

// RegisterRequest contains client-supplied user registration data.
//
// Role, status, user ID, timestamps, and password hash are intentionally
// absent because those values are controlled by NetPilot.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse contains the safe account information returned after
// successful registration.
type RegisterResponse struct {
	UserID string      `json:"user_id"`
	Name   string      `json:"name"`
	Email  string      `json:"email"`
	Role   user.Role   `json:"role"`
	Status user.Status `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

// Register creates a new NetPilot user account.
func (service *Service) Register(
	ctx context.Context,
	request RegisterRequest,
) (*RegisterResponse, error) {
	name, err := validateUserName(request.Name)
	if err != nil {
		return nil, err
	}

	email, err := normalizeAndValidateEmail(
		request.Email,
	)
	if err != nil {
		return nil, err
	}

	if err := validatePassword(request.Password); err != nil {
		return nil, err
	}

	passwordHash, err := service.passwordHasher.Hash(
		request.Password,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create password hash: %w",
			err,
		)
	}

	now := time.Now().UTC()

	newUser := &user.User{
		UserID: "usr_" + uuid.NewString(),

		Name:  name,
		Email: email,

		PasswordHash: passwordHash,

		Role:   user.RoleViewer,
		Status: user.StatusActive,

		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := service.userRepository.Create(
		ctx,
		newUser,
	); err != nil {
		return nil, fmt.Errorf(
			"create user account: %w",
			err,
		)
	}

	return &RegisterResponse{
		UserID:    newUser.UserID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Role:      newUser.Role,
		Status:    newUser.Status,
		CreatedAt: newUser.CreatedAt,
	}, nil
}

func validateUserName(
	name string,
) (string, error) {
	normalizedName := strings.TrimSpace(name)

	nameLength := utf8.RuneCountInString(
		normalizedName,
	)

	if nameLength < minimumUserNameLength ||
		nameLength > maximumUserNameLength {
		return "", ErrInvalidName
	}

	return normalizedName, nil
}

func normalizeAndValidateEmail(
	email string,
) (string, error) {
	normalizedEmail := strings.ToLower(
		strings.TrimSpace(email),
	)

	if normalizedEmail == "" ||
		len(normalizedEmail) > maximumEmailLength {
		return "", ErrInvalidEmail
	}

	parsedAddress, err := mail.ParseAddress(
		normalizedEmail,
	)
	if err != nil {
		return "", ErrInvalidEmail
	}

	if parsedAddress.Address != normalizedEmail {
		return "", ErrInvalidEmail
	}

	return normalizedEmail, nil
}

func validatePassword(
	password string,
) error {
	if utf8.RuneCountInString(password) <
		minimumPasswordLength {
		return ErrPasswordTooShort
	}

	if len([]byte(password)) > maximumPasswordBytes {
		return ErrPasswordTooLong
	}

	return nil
}
