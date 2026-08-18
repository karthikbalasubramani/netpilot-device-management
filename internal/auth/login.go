package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/karthikbalasubramani/netpilot-device-management/internal/user"
)

var (
	ErrInvalidCredentials = errors.New(
		"invalid email or password",
	)

	ErrAccountUnavailable = errors.New(
		"user account is not available for login",
	)
)

// LoginRequest contains credentials supplied by the client.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginUser contains safe authenticated-user information.
type LoginUser struct {
	UserID string      `json:"user_id"`
	Name   string      `json:"name"`
	Email  string      `json:"email"`
	Role   user.Role   `json:"role"`
	Status user.Status `json:"status"`
}

// LoginResponse contains the issued access token and authenticated user.
type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	ExpiresIn   int64     `json:"expires_in"`

	User LoginUser `json:"user"`
}

// Login authenticates a NetPilot user and issues an access token.
func (service *Service) Login(
	ctx context.Context,
	request LoginRequest,
) (*LoginResponse, error) {
	email := strings.ToLower(
		strings.TrimSpace(request.Email),
	)

	if email == "" ||
		strings.TrimSpace(request.Password) == "" {
		return nil, ErrInvalidCredentials
	}

	storedUser, err := service.userRepository.GetByEmail(
		ctx,
		email,
	)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf(
			"retrieve user for authentication: %w",
			err,
		)
	}

	err = service.passwordHasher.Verify(
		storedUser.PasswordHash,
		request.Password,
	)
	if err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf(
			"verify user password: %w",
			err,
		)
	}

	if storedUser.Status != user.StatusActive {
		return nil, ErrAccountUnavailable
	}

	accessToken, err := service.tokenIssuer.Generate(
		storedUser.UserID,
		storedUser.Role,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"generate access token: %w",
			err,
		)
	}

	lastLoginAt := time.Now().UTC()

	if err := service.userRepository.UpdateLastLogin(
		ctx,
		storedUser.UserID,
		lastLoginAt,
	); err != nil {
		return nil, fmt.Errorf(
			"update user last login: %w",
			err,
		)
	}

	expiresIn := max(int64(
		time.Until(accessToken.ExpiresAt).Seconds(),
	), 0)

	return &LoginResponse{
		AccessToken: accessToken.Token,
		TokenType:   accessTokenType,
		ExpiresAt:   accessToken.ExpiresAt,
		ExpiresIn:   expiresIn,

		User: LoginUser{
			UserID: storedUser.UserID,
			Name:   storedUser.Name,
			Email:  storedUser.Email,
			Role:   storedUser.Role,
			Status: storedUser.Status,
		},
	}, nil
}
