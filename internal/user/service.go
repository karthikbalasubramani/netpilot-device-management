package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidRole = errors.New(
		"Invalid user role",
	)

	ErrSelfRoleChange = errors.New(
		"Administrator cannot update their own role",
	)
)

type Service struct {
	repository Repository
}

func NewService(
	repository Repository,
) *Service {
	return &Service{
		repository: repository,
	}
}

func (service *Service) UpdateRole(
	ctx context.Context,
	actorUserID string,
	targetUserID string,
	role Role,
) (*User, error) {
	actorUserID = strings.TrimSpace(actorUserID)
	targetUserID = strings.TrimSpace(targetUserID)

	if actorUserID == "" {
		return nil, errors.New(
			"actor user_id is required",
		)
	}

	if targetUserID == "" {
		return nil, errors.New(
			"target user_id is required",
		)
	}

	if actorUserID == targetUserID {
		return nil, ErrSelfRoleChange
	}

	if !isValidRole(role) {
		return nil, ErrInvalidRole
	}

	targetUser, err :=
		service.repository.GetByUserID(
			ctx,
			targetUserID,
		)
	if err != nil {
		if errors.Is(
			err,
			ErrNotFound,
		) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf(
			"Get user before role update: %w",
			err,
		)
	}

	if targetUser.Role == role {
		return targetUser, nil
	}

	now := time.Now().UTC()

	err = service.repository.UpdateRole(
		ctx,
		targetUserID,
		role,
		now,
	)
	if err != nil {
		if errors.Is(
			err,
			ErrNotFound,
		) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf(
			"Update user role: %w",
			err,
		)
	}

	targetUser.Role = role
	targetUser.UpdatedAt = now

	return targetUser, nil
}

func ParseRole(
	value string,
) (Role, error) {
	role := Role(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)

	if !isValidRole(role) {
		return "", ErrInvalidRole
	}

	return role, nil
}

func isValidRole(role Role) bool {
	switch role {
	case Role("admin"):
		return true

	case Role("operator"):
		return true

	case Role("viewer"):
		return true

	default:
		return false
	}
}
