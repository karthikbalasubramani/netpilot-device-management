package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role represents the authorization level assigned to a NetPilot user.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// Status represents the current account state of a NetPilot user.
type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusPending  Status = "pending"
)

// User represents an account that can authenticate with NetPilot.
type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	UserID string `bson:"user_id" json:"user_id"`
	Name   string `bson:"name"    json:"name"`
	Email  string `bson:"email"   json:"email"`

	// PasswordHash stores only the securely hashed representation of a user's
	// password. The plain-text password must never be persisted.
	PasswordHash string `bson:"password_hash" json:"-"`

	Role   Role   `bson:"role"   json:"role"`
	Status Status `bson:"status" json:"status"`

	LastLoginAt *time.Time `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
